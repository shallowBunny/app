package lineUp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/shallowBunny/app/be/internal/bot/config"
	"github.com/shallowBunny/app/be/internal/bot/lineUp/inputs"
	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/ijt/go-anytime"
	"github.com/rs/zerolog/log"
)

type Set struct {
	Dj    string    `json:"dj"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Room  string    `json:"room"`
	kind  int
}

type Event struct {
	time     time.Time
	dj       string
	room     string
	priority int
}

func printTime(t time.Time) string {
	return t.Format("15:04")
}

func printTimeWithDay(t time.Time) string {
	return t.Format("Mon 15:04")
}

type LineUp struct {
	Sets    []Set
	events  []Event
	Inputs  inputs.Inputs
	Changes []inputs.InputCommandResultSet
	config  *config.Config
}

const (
	UnknownDJ   = "?"
	closed      = "🚫 closed"
	noDataRoom  = "⚠️ no data"
	openedFloor = "✅"
	//noData              = "⚠️ No data available yet ⚠️"
	missingData         = "\n\n⚠️ Some data is missing ⚠️"
	here                = " <- you are here"
	minSizeDJSearch     = 2
	minSizeDJSearchText = "Enter more than 2 characters for searching a DJ.\n"
	searchedMessage     = "Searched in DJ sets:\n"
)

func (l LineUp) DuplicateLineUp() *LineUp {
	new := &LineUp{
		Sets:    l.Sets,
		events:  l.events,
		Inputs:  l.Inputs,
		Changes: l.Changes,
		config:  l.config,
	}
	return new
}

func New(config *config.Config) *LineUp {
	days := []string{}

	for i := 0; i < config.NbDaysForInput; i++ {
		d := config.Lineup.BeginningSchedule.Add(time.Duration(24*i) * time.Hour).Format("Mon")
		days = append(days, d)
	}

	lineUp := &LineUp{
		Sets:    []Set{},
		events:  []Event{},
		Inputs:  inputs.New(days, config.Lineup.Rooms),
		Changes: []inputs.InputCommandResultSet{},
		config:  config,
	}

	for _, room := range config.Lineup.Rooms {
		sets, ok := config.Lineup.Sets[room]
		if !ok {
			log.Error().Msg(fmt.Sprintf("missing room <%v>", room))
			continue
		}

		for _, s := range sets {
			msg := lineUp.AddSet(lineUp.NewSet(s.Dj, room, s.Day, s.Hour, s.Minute, s.Duration, 0))
			if msg != "" {
				log.Error().Msg(msg)
			}
		}
	}

	emptyDurations := false
	for _, s := range lineUp.Sets {
		if s.End == s.Start {
			log.Error().Msg(fmt.Sprintf("empty duration for %v (%v)", s.Dj, s.Room))
			emptyDurations = true
		}
	}
	if emptyDurations {
		panic("empty durations")
	}

	return lineUp
}

func (l LineUp) GetSetsAndDurations() string {
	roomData := make(map[string]struct {
		setCount  int
		totalMins int
	})
	totalSets := 0
	totalDuration := 0
	var result strings.Builder

	// Loop through each set to count sets and calculate total duration for each room
	for _, set := range l.Sets {
		duration := int(set.End.Sub(set.Start).Minutes())
		if data, exists := roomData[set.Room]; exists {
			data.setCount++
			data.totalMins += duration
			roomData[set.Room] = data
		} else {
			roomData[set.Room] = struct {
				setCount  int
				totalMins int
			}{1, duration}
		}
		totalSets++
		totalDuration += duration
	}

	// Add the results for each room to the result string
	for room, data := range roomData {
		result.WriteString(fmt.Sprintf("Room: %s, Number of sets: %d, Total duration: %d minutes\n", room, data.setCount, data.totalMins))
	}

	// Add the total number of sets and total duration across all rooms
	result.WriteString(fmt.Sprintf("Total number of sets: %d\n", totalSets))
	result.WriteString(fmt.Sprintf("Total duration across all rooms: %d minutes\n", totalDuration))

	return result.String()
}

func (l *LineUp) Events(t time.Time) string {

	updatedEvents := []Event{}
	res := ""

	started := false
	for _, v := range l.events {
		if !v.time.After(t) {
			if !started {
				started = true
				res += v.dj + " started in " + v.room + "\n"
			} else {
				res += v.dj + " in " + v.room + "\n"
			}
		} else {
			updatedEvents = append(updatedEvents, v)
		}
	}
	l.events = updatedEvents
	return res
}

func (l LineUp) DumpEvents() string {
	res := ""
	var lastTime time.Time

	events := l.events
	sort.Slice(events, func(i, j int) bool {
		return events[i].time.Before(events[j].time)
	})

	for _, v := range events {
		if !v.time.Equal(lastTime) {
			res += "\n" + v.time.Format(time.Layout) + "\n"
			res += v.dj + " started in " + v.room + "\n"
		} else {
			res += v.dj + " in " + v.room + "\n"
		}
		lastTime = v.time
	}
	return res
}

func (l LineUp) IsUserInputing(chatID int64) bool {
	return l.Inputs.IsUserInputing(chatID)
}

func (l LineUp) IsUserInLogs(chatID int64) bool {
	return l.Inputs.IsUserInLogs(chatID)
}

func (l LineUp) CurrentInputCommand(chatID int64) string {
	return l.Inputs.CurrentInputCommand(chatID)
}

type InputCommandResult struct {
	Answer     string
	Buttons    []string
	AnswerModo string
	Changes    []inputs.InputCommandResultSet
}

func (l *LineUp) InputCommand(chatID int64, commandOrArg string) (*LineUp, InputCommandResult) {

	r := l.Inputs.InputCommand(chatID, commandOrArg)

	answerModo := ""
	var newLineup *LineUp = l

	if len(r.Sets) != 0 {
		if len(l.Changes) == 0 {
			newLineup = l.DuplicateLineUp()
		} else {
			newLineup = l
		}
		newLineup.Changes = append(newLineup.Changes, r.Sets...)

		log.Debug().Msg(fmt.Sprintf("List of changes for user %v <%v> in detached lineup", chatID, newLineup.Changes))

		for _, v := range r.Sets {
			s := newLineup.NewSet(v.Dj, v.Room, v.Day, v.Hour, v.Minute, v.Duration, 0)
			answerModo += "added " + newLineup.PrintSet(s) + "\n"
			answerModo += newLineup.AddSet(s)
		}
	}
	res := InputCommandResult{
		Answer:     r.Answer,
		Buttons:    r.Buttons,
		AnswerModo: answerModo,
	}

	return newLineup, res
}

func (l *LineUp) NewSet(djName string, room string, day int, hour int, min int, duration int, kind int) Set {
	t := l.config.Lineup.BeginningSchedule
	t1 := time.Date(t.Year(), t.Month(), t.Day(), hour, min, t.Second(), t.Nanosecond(), t.Location())
	t1 = t1.Add(time.Duration((int(time.Hour) * 24 * day)))

	set := Set{
		Dj:    djName,
		Start: t1,
		End:   t1.Add(time.Duration((int(time.Minute) * duration))),
		Room:  room,
		kind:  kind,
	}
	return set
}

func filterNonASCIIAndSpaces(input string) string {
	filtered := make([]rune, 0, len(input))
	for _, r := range input {
		if r <= unicode.MaxASCII && !unicode.IsSpace(r) {
			filtered = append(filtered, r)
		}
	}
	return string(filtered)
}

func (l *LineUp) FindRoom(source string, distanceMaxRoom int) (int, string) {
	source = strings.ToUpper(filterNonASCIIAndSpaces(source))

	minDistance := distanceMaxRoom + 1
	room := ""
	indexRoom := 0

	for i, v := range l.config.Lineup.Rooms {
		target := strings.ToUpper(filterNonASCIIAndSpaces(v))
		distance := levenshtein.DistanceForStrings([]rune(source), []rune(target), levenshtein.DefaultOptions)
		if distance <= distanceMaxRoom {
			if distance < minDistance {
				minDistance = distance
				room = v
				indexRoom = i
			}
		}
	}
	return indexRoom, room
}

func (l *LineUp) FindDJ(i string, when time.Time) string {

	if len(i) <= minSizeDJSearch {
		return minSizeDJSearchText
	}

	targets := strings.Fields(i)
	founds := make(map[string]bool)
	res := ""
	found := false
	for distance := 1; distance < 7; distance++ {
		for _, vv := range targets {
			target := strings.ToUpper(filterNonASCIIAndSpaces(vv))
			for _, vv := range l.Sets {
				if vv.Dj == UnknownDJ {
					continue
				}
				words := strings.Fields(vv.Dj)
				for _, source := range words {
					if len(source) < minSizeDJSearch {
						continue
					}
					source = strings.ToUpper(filterNonASCIIAndSpaces(source))
					d := levenshtein.DistanceForStrings([]rune(source), []rune(target), levenshtein.DefaultOptions)
					lastWasTrue := false
					if d < distance {
						foundThat := ""
						if vv.End.After(when) {
							foundThat += "✅ "
							foundThat += vv.Dj
							foundThat += " is playing "
							found = true
							lastWasTrue = true
						} else {
							found = true
							foundThat += "🚫 "
							foundThat += vv.Dj
							foundThat += " was playing "
						}
						foundThat += vv.Start.Format("Monday") + " at " + printTime(vv.Start) + " in " + vv.Room + "\n"
						_, ok := founds[foundThat]
						if !ok {
							if lastWasTrue {
								res += foundThat
							} else {
								res = foundThat + res
							}
							founds[foundThat] = true
						}
						break
					}
				}
			}
		}
		if found {
			break
		}
	}
	if !found {
		return searchedMessage + "Not found. 😔\n"
	}

	return searchedMessage + res
}

func (l *LineUp) AddSet(s Set) string {
	roomKnown := false
	msg := ""

	for _, v := range l.config.Lineup.Rooms {
		if v == s.Room {
			roomKnown = true
			break
		}
	}
	if !roomKnown {
		msg += fmt.Sprintf("Skipped  set <%v> because unknown room <%v>\n", l.PrintSet(s), s.Room)
		return msg
	}

	resSet := []Set{}

	for _, v := range l.Sets {
		skip := false

		if v.End.After(s.Start) && v.Start.Before(s.End) && v.Room == s.Room {
			skip = true
		}

		if v.End == v.Start {
			log.Trace().Msg(fmt.Sprintf("empty duration for %v", v.Dj))
			v.End = s.Start
		}

		if !skip {
			resSet = append(resSet, v)
		} else {
			msg += v.Room + " deleted <" + l.PrintSet(v) + "> because it collided with <" + l.PrintSet(s) + ">\n"
		}
	}
	resSet = append(resSet, s)

	sort.Slice(resSet, func(i, j int) bool {
		if resSet[i].Room != resSet[j].Room {
			for _, v := range l.config.Lineup.Rooms {
				if v == resSet[i].Room {
					return false
				}
				if v == resSet[j].Room {
					return true
				}
			}
			return resSet[i].Room > resSet[j].Room
		} else {
			return resSet[i].Start.Before(resSet[j].Start)
		}

	})

	l.Sets = resSet

	//log.Trace().Msg(fmt.Sprintf("start=%v", s.Start))

	l.computeEvents()
	return msg
}
func (l *LineUp) Init(config *config.Config) {
	l.config = config
	l.computeEvents()
}

func (l *LineUp) computeEvents() {
	events := []Event{}
	for _, v := range l.Sets {
		priority := 0
		for i, v2 := range l.config.Lineup.Rooms {
			if v2 == v.Room {
				priority = i
				break
			}
		}
		if v.Start.After(time.Now()) {
			events = append(events, Event{time: v.Start, dj: v.Dj, room: v.Room, priority: priority})
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].priority > events[j].priority
	})
	l.events = events
}

// SameDay function checks if two dates are on the same day
func sameDay(date1, date2 time.Time) bool {
	// Extract year, month, and day for both dates and compare
	return date1.Year() == date2.Year() &&
		date1.Month() == date2.Month() &&
		date1.Day() == date2.Day()
}

func (l LineUp) printRoom(sets []Set, oldData bool, oldLineupMessage string, currentTime time.Time, youAre string, filterNomSalle string) string {
	var res string
	var closingTime time.Time

	printedYouAreHere := oldData

	log.Trace().Msg(fmt.Sprintf("printRoom: oldLineup:%v currentTime:%v", oldLineupMessage, currentTime))

	sort.Slice(sets, func(i, j int) bool {
		return sets[i].Start.Before(sets[j].Start)
	})

	currentDay := -1 // Initialize to a value that will not match any day in sets
	foundData := false

	for _, set := range sets {
		if set.Dj != UnknownDJ {
			foundData = true
		}
		if closingTime.Before(set.End) {
			closingTime = set.End
		}
	}
	if !foundData {
		return l.config.BotNoDataAvailableYet
	}

	for _, set := range sets {
		if set.Dj == UnknownDJ {
			continue
		}

		// Check if we need to add the day header
		if set.Start.Day() != currentDay {
			// Print "you are here" before changing the day if not yet printed and it's still today
			if !printedYouAreHere && currentDay != -1 && currentDay == currentTime.Day() {
				res += youAre + here + "\n"
				printedYouAreHere = true
			}
			if currentDay != -1 {
				res += "\n"
			}
			if sameDay(currentTime, set.Start) {
				res += "Today:\n"
			} else {
				res += set.Start.Format("Monday") + ":\n"
			}
			currentDay = set.Start.Day()
		}

		// Print "you are here" before the next event if it's the first future event
		if set.Start.After(currentTime) && !printedYouAreHere {
			res += youAre + here + "\n"
			printedYouAreHere = true
		}

		res += printTime(set.Start) + " " + set.Dj
		if filterNomSalle == "" {
			res += " " + set.Room
		}
		res += "\n"
	}

	// Check if the closing time is after the current time and add "you are here" if not yet printed
	if closingTime.After(currentTime) {
		if !printedYouAreHere {
			res += youAre + here + "\n"
		}
		res += printTime(closingTime) + " closing\n"
	} else {
		res += printTime(closingTime) + " closed\n"
	}

	res += oldLineupMessage

	return res
}

func (l LineUp) PrintForMerge(filterNomSalle string) string {
	s := []Set{}
	for _, v := range l.Sets {
		if v.Room != filterNomSalle || v.Dj == UnknownDJ {
			continue
		}
		s = append(s, v)
	}
	if len(s) == 0 {
		return ""
	}
	res := ""

	sort.Slice(s, func(i, j int) bool {
		return s[i].Start.Before(s[j].Start)
	})

	res += "\n" + filterNomSalle + ":\n\n"

	var lastClosing time.Time

	for _, v := range s {

		if !lastClosing.IsZero() && v.Start != lastClosing {
			res += printTimeWithDay(lastClosing) + " to " + printTime(v.Start) + " ⏸️\n"
		}
		res += v.Start.Format("Mon") + " " + printTime(v.Start) + " to " + printTime(v.End) + " " + v.Dj
		res += "\n"

		lastClosing = v.End
	}
	return res
}

func (l LineUp) Print(youAreHere string, filter int, filterNomSalle string) string {
	current := time.Now()
	s := []Set{}
	var oldData bool = true
	var oldLineupMessage = l.config.BotOldLineupMessage
	for _, v := range l.Sets {
		if v.End.After(current) {
			oldLineupMessage = ""
			oldData = false
		}
		if filter != -1 {
			if v.kind != filter {
				continue
			}
		}
		if filterNomSalle != "" {
			if v.Room != filterNomSalle {
				continue
			}
		}
		s = append(s, v)
	}

	res := ""
	if filterNomSalle != "" {
		res += "Lineup in " + filterNomSalle + "\n" + "\n"
	}

	if len(s) == 0 {
		log.Warn().Msg(fmt.Sprintf("Print: returning %s", l.config.BotNoDataAvailableYet))
		res += l.config.BotNoDataAvailableYet
		return res
	}
	return res + l.printRoom(s, oldData, oldLineupMessage, current, youAreHere, filterNomSalle)
}

func (l LineUp) getDayNumber(t time.Time) int {
	// TOFIX
	diff := int(t.Sub(l.config.Lineup.BeginningSchedule).Hours())
	if diff < 0 {
		log.Error().Msg("getDayNumber on date after beginning")
		return 0
	}
	sd := l.config.Lineup.BeginningSchedule
	added := time.Hour
	currentDay := l.config.Lineup.BeginningSchedule.Day()
	dayNumber := 0
	for {
		currentTime := sd.Add(added)
		added += time.Hour
		if currentTime.Day() != currentDay {
			dayNumber++
			currentDay = currentTime.Day()
		}
		if sd.Add(added).After(t) {
			return dayNumber
		}
	}
}

func (l LineUp) PrintSet(v Set) string {
	return "- '" + strconv.Itoa(l.getDayNumber(v.Start)) + " " + printTime(v.Start) + " " + strconv.Itoa(int(v.End.Sub(v.Start).Minutes())) + " " + strconv.Itoa(v.kind) + " " + v.Dj + "'"
}

func (l LineUp) Dump() string {

	foundAny := false
	res := ""
	//current := time.Now()
	room := ""

	var lastClosing time.Time
	for i, v := range l.Sets {
		if v.Room != room {
			room = v.Room
			if i != 0 {
				res += "\n"
			}
			res += room + ":\n"
			lastClosing = time.Time{}
		} else {
			if i != 0 {
				res += "\n"
			}
		}
		if !lastClosing.IsZero() && v.Start != lastClosing {
			res += "# hole: " + printTime(lastClosing) + " to " + printTime(v.Start) + "\n"
		}
		if !lastClosing.IsZero() && v.Start.Before(lastClosing) {
			res += "# wrong data? " + lastClosing.String() + " to " + v.Start.String() + "\n"
		}
		lastClosing = v.End

		res += l.PrintSet(v)

		foundAny = true
	}
	res += "\n"
	if !foundAny {
		res = "No data"
	}

	return res
}

func (l LineUp) Hole() string {

	foundAny := false
	res := ""
	room := ""

	var lastClosing time.Time
	var lastDJ string
	for _, v := range l.Sets {
		if v.Room != room {
			room = v.Room
			lastClosing = time.Time{}
		}
		if !lastClosing.IsZero() && v.Start != lastClosing {
			res += room + " gap: " + printTimeWithDay(lastClosing) + " to " + printTimeWithDay(v.Start) + " (" + lastDJ + " -> " + v.Dj + ")\n"
		}
		if !lastClosing.IsZero() && v.Start.Before(lastClosing) {
			res += room + " " + v.Dj + "🚫 wrong data? " + lastClosing.String() + " to " + v.Start.String() + "\n"
		}
		lastClosing = v.End
		lastDJ = v.Dj

		foundAny = true
	}
	res += "\n"
	if !foundAny {
		res = "No data"
	}

	return res
}

func (l LineUp) PrintCurrent() string {
	return l.PrintCurrentForTime(nil)
}

func (l LineUp) calculatePause(closingTime time.Time, room string) *time.Duration {
	for _, v := range l.Sets {
		if v.Room == room {
			if v.Start.After(closingTime) {
				distance := v.Start.Sub(closingTime)
				return &distance
			}
		}
	}
	return nil
}

func (l LineUp) PrintCurrentForTime(when *string) string {

	var room string
	var res string

	nbDjs := 0
	foundCurrent := false
	nextFound := false
	//isLast := true
	current := time.Now()
	var currentClosingTime time.Time

	if when != nil {
		v, err := anytime.Parse(*when, current)
		if err != nil {
			log.Debug().Msg(fmt.Sprintf("%v parsing <%v>", err.Error(), *when))
			res = "teleporting command failed, I couldnt parse your input.\n"
		} else {
			current = v
			res = "teleporting to " + current.String() + "\n"
		}
	}

	foundRoom := make(map[string]bool)

	roomsFound := 0

	for i, v := range l.Sets {

		if v.Room != room {
			foundRoom[v.Room] = true
			roomsFound++
			permanentelyClosed := false
			if !nextFound && i != 0 {
				if foundCurrent {
					res += " (closing at " + printTime(currentClosingTime) + ")"
				} else {
					if !l.config.NowSkipClosed {
						res += room + " " + closed
					} else {
						permanentelyClosed = true
					}
				}
			}
			if !permanentelyClosed {
				res += "\n"
			}
			room = v.Room
			foundCurrent = false
			nextFound = false
		}

		if (v.Start.Before(current) || v.Start.Equal(current)) && v.End.After(current) {
			res += room + " " + openedFloor + " " + v.Dj
			if v.Dj != UnknownDJ {
				nbDjs++
			}
			foundCurrent = true
			currentClosingTime = v.End
			continue
		}

		if v.Start.After(current) && !nextFound {
			if !foundCurrent {
				dj := v.Dj
				res += room + " " + closed //+ " "
				if dj == UnknownDJ {
					res += " until"
				} else {
					res += " (" + dj
				}

				if dj != UnknownDJ {
					nbDjs++
					if current.Day() != v.Start.Day() {
						res += ", " + v.Start.Format("Mon")
					}
				} else {
					res += " " + v.Start.Format("Mon")
				}
				res += " at"
				res += " " + printTime(v.Start)
				if dj != UnknownDJ {
					res += ")"
				}
				nextFound = true
				continue
			} else {
				if currentClosingTime != v.Start {
					pauseTime := l.calculatePause(currentClosingTime, room)
					if pauseTime == nil || *pauseTime > time.Hour*2 {
						res += " (closing at " + printTime(currentClosingTime) + ")"
					} else {
						res += fmt.Sprintf(" (%v at %v after %vmin pause)", v.Dj, printTime(v.Start), pauseTime.Minutes())
					}
				} else {
					res += " (" + v.Dj + " at " + printTime(v.Start) + ")"
				}
				nextFound = true
				continue
			}
		}

	}

	if !nextFound {
		if foundCurrent {
			res += " (closing at " + printTime(currentClosingTime) + ")"
		} else {
			if !l.config.NowSkipClosed && room != "" {
				res += room + " " + closed
			}
		}
	}

	for _, v := range l.config.Lineup.Rooms {
		ok := foundRoom[v]
		if !ok {
			res += "\n" + v + " " + noDataRoom
		}
	}

	if res == "" || res == "\n" {
		log.Warn().Msg(fmt.Sprintf("PrintCurrentForTime: returning %s", l.config.BotNoDataAvailableYet))
		res = l.config.BotNoDataAvailableYet
	} else if roomsFound != len(l.config.Lineup.Rooms) {
		res += missingData
	}

	return res
}
