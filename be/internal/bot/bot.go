package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ottoDaffy/go-diff/diffmatchpatch"
	"github.com/shallowBunny/app/be/internal/bot/config"
	"github.com/shallowBunny/app/be/internal/bot/dao"
	"github.com/shallowBunny/app/be/internal/bot/lineUp"
	"github.com/shallowBunny/app/be/internal/bot/lineUp/inputs"
	"github.com/shallowBunny/app/be/internal/bot/users"

	"github.com/rs/zerolog/log"
)

const (
	nothingToMergeMessage     = "You have nothing to merge, please use the /input command first"
	deltaDefault              = time.Hour
	durationEvent             = time.Hour * (5)
	distanceMaxRoom           = 3
	distanceMaxRoomWithSlash  = 6
	modifiedLineUpMessage     = "\n\n⚠️ You are viewing a modified version of the lineUp, please use the /merge command to share your Changes with others or /input to add more Changes ⚠️\n"
	MergedMessageAccepted     = "✅ Your merge request #%d has been accepted by %v, thanks!"
	MergedMessageRefused      = "💔 Your merge request #%d has been refused by %v."
	rebaseCommandErrorMessage = "No merge requests to rebase."
	stopNotificationsCommand  = "🔴"
	stoppedNoticationsMessage = "You stopped Dj changes notifications"
	startNotificationsCommand = "🟢"
	startedNoticationsMessage = "You enabled Dj changes notifications"
	helpCommand               = "Help"
	maxMnbRoomsForRoomButton  = 100
	maxSize                   = 4096
	noMotdMessage             = "No help available"
)

var (
	mergeRequestID = 0
)

type MergeRequests struct {
	Changes           []inputs.InputCommandResultSet
	UserId            int64
	User              string
	Created           time.Time
	ID                int
	Info              string
	BeginningSchedule time.Time
}

type Bot struct {
	dao                    dao.Dao
	users                  users.Users
	UsersLineUps           map[int64]*lineUp.LineUp // userId -> LineUp
	UsersMergeRequest      []MergeRequests
	RootLineUp             *lineUp.LineUp
	admins                 []int
	modos                  []int
	channel                chan Message
	config                 *config.Config
	commandsHistoryLogFile *os.File
	logs                   string
	roomsEmoticons         []string
	magicRoomButton        bool
}

const (
	KeyboardHidden = iota
	Keyboard
)

type KeyboardType int

type Message struct {
	UserID  int64
	Text    string
	Buttons []string
	Html    bool
}

func (b Bot) GetMessageChannel() chan Message {
	return b.channel
}

func (b Bot) IsAdmin(user int64) bool {
	for _, v := range b.admins {
		if v == int(user) {
			return true
		}
	}
	return false
}

func (b Bot) IsModo(user int64) bool {
	for _, v := range b.modos {
		if v == int(user) {
			return true
		}
	}
	return false
}

func (b Bot) Save() error {
	bytes, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	botString := string(bytes)
	res, err := PrettyString(botString)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	_ = res
	//log.Trace().Msg(res)

	return b.dao.SaveBot(b.config.Lineup.BeginningSchedule, botString)
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

// ExtractEmoticons extracts emoticons (emoji) from the input string and returns them as a single concatenated string
func ExtractEmoticons(input string) string {
	// Define a regex pattern to match emoticons (emoji)
	emojiPattern := `[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F700}-\x{1F77F}]|[\x{1F780}-\x{1F7FF}]|[\x{1F800}-\x{1F8FF}]|[\x{1F900}-\x{1F9FF}]|[\x{1FA00}-\x{1FA6F}]|[\x{1FA70}-\x{1FAFF}]|[\x{1FB00}-\x{1FBFF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{2B50}-\x{2B55}]|[\x{231A}-\x{231B}]|[\x{23E9}-\x{23F3}]|[\x{23F8}-\x{23FA}]|[\x{1F004}]|[\x{1F0CF}]|[\x{1F18E}]|[\x{1F191}-\x{1F19A}]|[\x{1F1E6}-\x{1F1FF}]|[\x{1F201}-\x{1F251}]|[\x{1F300}-\x{1F321}]|[\x{1F324}-\x{1F393}]|[\x{1F396}-\x{1F397}]|[\x{1F399}-\x{1F39B}]|[\x{1F39E}-\x{1F3F0}]|[\x{1F3F3}-\x{1F3F5}]|[\x{1F3F7}-\x{1F4FD}]|[\x{1F4FF}-\x{1F53D}]|[\x{1F549}-\x{1F54E}]|[\x{1F550}-\x{1F567}]|[\x{1F56F}-\x{1F570}]|[\x{1F573}-\x{1F57A}]|[\x{1F587}]|[\x{1F58A}-\x{1F58D}]|[\x{1F590}]|[\x{1F595}-\x{1F596}]|[\x{1F5A4}]|[\x{1F5A5}]|[\x{1F5A8}]|[\x{1F5B1}-\x{1F5B2}]|[\x{1F5BC}]|[\x{1F5C2}-\x{1F5C4}]|[\x{1F5D1}-\x{1F5D3}]|[\x{1F5DC}-\x{1F5DE}]|[\x{1F5E1}]|[\x{1F5E3}]|[\x{1F5E8}]|[\x{1F5EF}]|[\x{1F5F3}]|[\x{1F5FA}]|[\x{1F5FB}-\x{1F5FF}]|[\x{1F600}-\x{1F644}]|[\x{1F645}-\x{1F64F}]|[\x{1F680}-\x{1F6C5}]|[\x{1F6CB}-\x{1F6D2}]|[\x{1F6E0}-\x{1F6E5}]|[\x{1F6E9}]|[\x{1F6EB}-\x{1F6EC}]|[\x{1F6F0}]|[\x{1F6F3}-\x{1F6F6}]|[\x{1F6F7}-\x{1F6F8}]|[\x{1F6F9}]|[\x{1F6FA}]|[\x{1F7E0}-\x{1F7EB}]|[\x{1F90D}-\x{1F90F}]|[\x{1F910}-\x{1F93A}]|[\x{1F93C}-\x{1F945}]|[\x{1F947}-\x{1F94C}]|[\x{1F94D}-\x{1F94F}]|[\x{1F950}-\x{1F95E}]|[\x{1F95F}-\x{1F96B}]|[\x{1F96C}-\x{1F970}]|[\x{1F971}]|[\x{1F973}-\x{1F976}]|[\x{1F97A}]|[\x{1F97C}-\x{1F97F}]|[\x{1F980}-\x{1F984}]|[\x{1F985}-\x{1F991}]|[\x{1F992}-\x{1F997}]|[\x{1F998}-\x{1F9A2}]|[\x{1F9A5}-\x{1F9AA}]|[\x{1F9AE}-\x{1F9AF}]|[\x{1F9B0}-\x{1F9B9}]|[\x{1F9BC}-\x{1F9CC}]|[\x{1F9CD}-\x{1F9CF}]|[\x{1F9D0}-\x{1F9E6}]|[\x{1F9E7}-\x{1F9FF}]|[\x{1FA70}-\x{1FA74}]|[\x{1FA78}-\x{1FA7A}]|[\x{1FA80}-\x{1FA82}]|[\x{1FA90}-\x{1FA95}]`
	emojiRegex := regexp.MustCompile(emojiPattern)

	// Find all matches in the input string
	matches := emojiRegex.FindAllString(input, -1)

	// Concatenate all matches into a single string
	return strings.Join(matches, "")
}

func New(dao dao.Dao, config *config.Config) *Bot {
	var f *os.File
	var err error
	if config.CommandsHistoryLogFile != "" {
		f, err = os.OpenFile(config.CommandsHistoryLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
	}

	bot := &Bot{}

	gotBotFromDB := true

	if config.ReadSetsFromRedisOnRestart {
		botString, err := dao.GetBot(config.Lineup.BeginningSchedule)
		if err == nil {
			res, err := PrettyString(botString)
			if err != nil {
				log.Error().Msg(err.Error())
			}
			_ = res

			err = json.Unmarshal([]byte(botString), bot)
			if err != nil {
				log.Error().Msg(err.Error())
				gotBotFromDB = false
			}
			log.Info().Msg("loading bot from redis")
			emptyLineup := true
			for _, v := range bot.UsersLineUps {
				if len(v.Sets) != 0 {
					emptyLineup = false
				}
				v.Init(config)
			}
			bot.RootLineUp.Init(config)
			if len(bot.RootLineUp.Sets) != 0 {
				emptyLineup = false
			}
			if emptyLineup {
				log.Warn().Msg("empty lineups: using config")
				gotBotFromDB = false
			}
		} else {
			log.Warn().Msg("not found in redis: using config")
			gotBotFromDB = false
		}
	} else {
		log.Debug().Msg("input=false using config")
		gotBotFromDB = false
	}

	if !gotBotFromDB {
		bot = &Bot{
			UsersLineUps: make(map[int64]*lineUp.LineUp),
			RootLineUp:   lineUp.New(config),
		}
		gotBotFromDB = false
		log.Info().Msg("loading bot from config")
	}

	bot.dao = dao
	bot.users = users.New(dao, config.Lineup.BeginningSchedule)
	bot.commandsHistoryLogFile = f
	bot.config = config
	bot.channel = make(chan Message)
	bot.admins = config.Admins
	bot.modos = config.Modos
	bot.magicRoomButton = len(config.Lineup.Rooms) < maxMnbRoomsForRoomButton

	for _, v := range config.Lineup.Rooms {
		emo := ExtractEmoticons(v)
		log.Trace().Msg(fmt.Sprintf("Rooms:%v -> <%v>", v, emo))
		bot.roomsEmoticons = append(bot.roomsEmoticons, emo)
	}

	if bot.UsersLineUps == nil {
		panic("nil user lineup")
	}

	if !gotBotFromDB {
		err := bot.Save()
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}

	log.Debug().Msg(bot.RootLineUp.GetSetsAndDurations())

	bot.logs, err = dao.GetLogs()
	if err != nil {
		if err.Error() == "redis: nil" {
			log.Warn().Msg("empty dao.GetLogs")
		} else {
			log.Error().Msg(err.Error())
		}
	}

	go bot.SendEvents()

	return bot
}
func (b Bot) GetConfig() *config.Config {
	return b.config
}

// user = 0 pour les logs web
func (b *Bot) Log(user int64, command, userString string) {

	logString := "\"" + command + "\", //" + time.Now().Format("Mon 15:04") + " " + userString + " " + strconv.Itoa(int(user)) + "\n"

	if user != 0 && b.commandsHistoryLogFile != nil {
		if _, err := b.commandsHistoryLogFile.WriteString(logString); err != nil {
			log.Error().Msg(err.Error())
		}
	}

	if !b.IsAdmin(user) {
		logString := time.Now().Format("Mon 15:04") + " " + userString + " <" + command + ">\n"
		if user != 0 {
			b.logs = Trim(b.logs + logString)
			err := b.dao.SaveLogs(b.logs)
			if err != nil {
				log.Error().Msg(err.Error())
			}
		}
		for _, u := range b.admins {
			userId := int64(u)
			lineup := b.GetLineUpForUser(userId)
			if lineup.IsUserInLogs(userId) {
				b.sendMessage(userId, logString)
			}
		}
	}
}

func Trim(a string) string {
	if len(a) > maxSize {
		r := strings.Split(a, "\n")
		zz := len(r)
		total := 0
		upTo := 0
		for zzz := zz - 1; zzz >= 0; zzz-- {
			total += len(r[zzz]) + 1
			if total > maxSize {
				upTo = zzz + 1
				break
			}
		}

		res := ""
		for zz := upTo; zz < len(r); zz++ {
			if r[zz] == "" {
				continue
			}
			res += r[zz] + "\n"
		}

		return res
	}
	return a
}

// SplitMessages function to trim and split large messages
func splitMessages(messages []Message) []Message {
	var result []Message

	for _, msg := range messages {
		if len(msg.Text) <= maxSize {
			result = append(result, msg)
		} else {
			start := 0
			for start < len(msg.Text) {
				end := start + maxSize
				if end > len(msg.Text) {
					end = len(msg.Text)
				} else {
					lastNewline := strings.LastIndex(msg.Text[start:end], "\n")
					if lastNewline != -1 {
						end = start + lastNewline + 1
					}
				}
				newMsg := Message{
					UserID:  msg.UserID,
					Text:    msg.Text[start:end],
					Buttons: msg.Buttons,
					Html:    msg.Html,
				}
				result = append(result, newMsg)
				start = end
			}
		}
	}
	return result
}

func (b Bot) SendAdminsMessage(input string) {
	msg := "#admin " + input
	for _, v := range b.admins {
		b.sendMessage(int64(v), msg)
	}
	log.Info().Msg(msg)
}

func (b Bot) SendModosMessage(input string) {
	msg := "#modo " + input
	for _, v := range b.modos {
		b.sendMessage(int64(v), msg)
	}
	log.Info().Msg(msg)
	log.Info().Msg(fmt.Sprintf("send to %v", b.modos))

}

func (b Bot) sendMessage(userId int64, msg string) {
	if b.channel != nil {
		buttons := b.GetButtonsForUser(userId)
		b.channel <- Message{
			UserID:  userId,
			Text:    msg,
			Buttons: buttons,
		}
	}
}

func (b *Bot) SendEvents() {
	maxUser := 0

	for {
		eventsText := b.RootLineUp.Events(time.Now())

		users := b.users.UsersWithNotifications()

		newUsers, totalUsers, _, _ := b.users.UsersStats()

		if totalUsers > maxUser {
			maxUser = totalUsers
			b.SendAdminsMessage(fmt.Sprintf("New max active users: %d new users: %d", maxUser, newUsers))
		}

		if eventsText != "" {

			log.Info().Msg(fmt.Sprintf("sending %d events", len(users)))
			for _, v := range users {
				msgForUser := eventsText
				if v > 0 { // SKIP pour les groups
					log.Debug().Msg(fmt.Sprintf("sending event for %v", v))
					b.sendMessage(v, msgForUser)
				} else {
					log.Debug().Msg(fmt.Sprintf("skipped sending event for %v", v))
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func (b *Bot) parseCommand(chatId int64, str string) (string, string) {

	lineup := b.GetLineUpForUser(chatId)
	command := str
	arg := strings.ToLower(nonAlphanumericRegex.ReplaceAllString(str, ""))
	if strings.Contains(str, " ") {
		command = str[:strings.Index(str, " ")]
		arg = str[strings.Index(str, " "):]
	}
	command = strings.ToLower(nonAlphanumericRegex.ReplaceAllString(command, ""))

	if lineup.IsUserInputing(chatId) {
		command = lineup.CurrentInputCommand(chatId)
		arg = str
	}

	if len(command) == 0 {
		command = str
	}

	return command, arg
}

func (b *Bot) DeleteUser(chatId int64) error {
	return b.users.DeleteUser(chatId)
}

func (b *Bot) ShowRoom(chatId int64, lineup *lineUp.LineUp, index int) string {
	if b.magicRoomButton {
		b.users.SetLastShownRoom(chatId, (index+1)%len(b.config.Lineup.Rooms))
	}
	return lineup.Print(b.config.Meta.RoomYouAreHereEmoticon, -1, b.config.Lineup.Rooms[index])
}

func (b *Bot) defaultCommand(orig string, lineup *lineUp.LineUp, chatId int64) string {

	for index, room := range b.roomsEmoticons {
		if orig == room {
			return b.ShowRoom(chatId, lineup, index)
		}
	}

	distance := distanceMaxRoom

	log.Debug().Msg(fmt.Sprintf("defaultCommand <%v>", orig))

	if len(orig) < 1 {
		log.Warn().Msg(fmt.Sprintf("defaultCommand: chatId:%v len(orig) < 1 <%v>", chatId, orig))
		return ""
	}

	if orig[:1] == "/" {
		distance = distanceMaxRoomWithSlash
	}
	index, room := lineup.FindRoom(orig, distance)
	if room != "" {
		return b.ShowRoom(chatId, lineup, index)
	} else {
		return lineup.FindDJ(orig, time.Now())
	}
}

func (b Bot) GetLineUpForUser(chatId int64) *lineUp.LineUp {
	l, ok := b.UsersLineUps[chatId]
	if ok {
		log.Debug().Msg(fmt.Sprintf("using local lineup for user %v", chatId))
		return l
	}
	return b.RootLineUp
}

func (b Bot) PrintLineupForCheckConfig() string {
	res := "\n\nLineup in each room:\n"
	for _, v := range b.config.Lineup.Rooms {
		res += b.RootLineUp.PrintForMerge(v)
	}
	return res
}

func (b Bot) compareLineUps(lineupA, lineupB *lineUp.LineUp) (string, error) {
	log.Debug().Msg("*** compareLineUps")

	dmp := diffmatchpatch.New()
	var err error
	res := ""
	changed := false

	// Use rooms from the bot's config
	for _, room := range b.config.Lineup.Rooms {
		f := lineupA.PrintForMerge(room)
		log.Debug().Msg(fmt.Sprintf("1/ compareLineUps: <%v:%v> ", f, room))

		f2 := lineupB.PrintForMerge(room)
		log.Debug().Msg(fmt.Sprintf("2/ compareLineUps: <%v:%v> ", f2, room))

		if f == f2 {
			log.Debug().Msg("same")
			continue
		}
		log.Debug().Msg("diff")

		changed = true
		fileAdmp, fileBdmp, dmpStrings := dmp.DiffLinesToChars(f, f2)
		diffs := dmp.DiffMain(fileAdmp, fileBdmp, false)
		diffs = dmp.DiffCharsToLines(diffs, dmpStrings)
		diffs = dmp.DiffCleanupSemantic(diffs)
		res += dmp.DiffPrettyText(diffs)
	}
	if !changed {
		err = errors.New("No change with current lineup")
	}
	log.Debug().Msg("end")

	log.Debug().Msg(fmt.Sprintf("compareLineUps: <%v> <%v> -> <%v> err:<%v>", lineupA.Dump(), lineupB.Dump(), res, err))

	return res, err
}

func (b *Bot) ProcessCommand(chatId int64, text, user string) []Message {
	command, arg := b.parseCommand(chatId, text)
	log.Debug().Msg(fmt.Sprintf("%v sent <%v> command <%v> arg <%v>", user, text, command, arg))
	return b.runCommand(chatId, command, arg, text, user)
}

func NewMergeRequest(beginningSchedule time.Time, changes []inputs.InputCommandResultSet, chatId int64, user string, answer string) *MergeRequests {
	mr := MergeRequests{
		Changes:           changes,
		UserId:            chatId,
		User:              user,
		Created:           time.Now(),
		ID:                mergeRequestID,
		Info:              answer,
		BeginningSchedule: beginningSchedule,
	}
	mergeRequestID++
	return &mr
}

func (b *Bot) CreateMergeRequest(mr MergeRequests) {
	b.UsersMergeRequest = append(b.UsersMergeRequest, mr)
	modoMsg := fmt.Sprintf("new merge request #%d from %v, use /rebase command to merge\n%v", mr.ID, mr.User, mr.Info)
	log.Debug().Msg(fmt.Sprintf("new merge request from %v <%v>", mr.User, mr))
	log.Debug().Msg(modoMsg)
	b.SendModosMessage(modoMsg)
}

func (b Bot) ChecForDuplicateMergeRequest(r *MergeRequests) error {
	for _, mr := range b.UsersMergeRequest {
		if len(mr.Changes) == len(r.Changes) {
			foundDifference := false
			for i := range mr.Changes {
				if mr.Changes[i] != r.Changes[i] {
					foundDifference = true
					continue
				}
			}
			if !foundDifference {
				return errors.New("Similar MR exists")
			}
		}
	}
	return nil
}

func (b Bot) CheckMergeRequest(r *MergeRequests) (string, error) {
	var answer string
	var err error

	l := b.RootLineUp.DuplicateLineUp()
	answer += fmt.Sprintf("Merge request %d from %v (submitted %v)\n\n", r.ID, r.User, r.Created.Format("Mon 15:04"))
	for _, v := range r.Changes {
		s := l.NewSet(v.Dj, v.Room, v.Day, v.Hour, v.Minute, v.Duration, 0)
		log.Debug().Msg("added " + l.PrintSet(s) + "\n")
		log.Debug().Msg(l.AddSet(s))
	}
	compare, err := b.compareLineUps(b.RootLineUp, l)
	answer += compare
	log.Debug().Msg(fmt.Sprintf("compare:<%v>", compare))
	return answer, err
}

func (b *Bot) runCommand(chatId int64, command, arg, orig string, user string) []Message {

	messages := []Message{}

	answer := ""
	var buttons []string
	res := ""
	lineUp := b.GetLineUpForUser(chatId)

	if !b.users.DoesUserExists(chatId) {
		log.Info().Msg("new user")
		if b.config.BotMotd != "" {
			messages = append(messages, Message{
				Text:    b.config.BotMotd,
				Buttons: nil,
				UserID:  chatId,
				Html:    true,
			})
		}
	}

	var adminMsg string
	var html bool

	log.Debug().Msg(fmt.Sprintf("command <%v> <%v>", command, stopNotificationsCommand))

	switch command {
	case "start":
		err := b.users.SetUserAsNew(chatId)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		answer += lineUp.PrintCurrent()
	case strings.ToLower(helpCommand):
		if b.config.BotMotd == "" {
			answer = noMotdMessage
		} else {
			answer = b.config.BotMotd + "\n\n"
			html = true
		}
	case "stop", stopNotificationsCommand:
		err := b.users.SetNotificationsUser(chatId, false)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		answer = stoppedNoticationsMessage
	case startNotificationsCommand:
		err := b.users.SetNotificationsUser(chatId, true)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		answer = startedNoticationsMessage
	case "p", "all":
		res += lineUp.Print(b.config.Meta.RoomYouAreHereEmoticon, -1, "")
		answer = res
	case "now":
		answer = lineUp.PrintCurrent()
	case "t":
		if arg != "" {
			res += lineUp.PrintCurrentForTime(&arg)
		} else {
			res += lineUp.PrintCurrent()
		}
		answer = res
	case "events":
		if b.IsAdmin(chatId) {
			answer = lineUp.DumpEvents()
		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
	case "print":
		if b.IsAdmin(chatId) {
			answer = b.PrintLineupForCheckConfig()
		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
	case "dump":
		if b.IsAdmin(chatId) {
			answer = lineUp.Dump()
			log.Debug().Msg(answer)

		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
	case "hole":
		if b.IsAdmin(chatId) {
			answer = lineUp.Hole()
			log.Debug().Msg(answer)

		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
		// TODO checker ce qui se passe si 2 users se mettent en rebase en meme temps
	case inputs.RebaseCommand:
		if b.IsModo(chatId) {
			if len(b.UsersMergeRequest) == 0 {
				answer += rebaseCommandErrorMessage
			} else {
				switch lineUp.CurrentInputCommand(chatId) {
				case "":

					a, err := b.CheckMergeRequest(&b.UsersMergeRequest[0])
					if err != nil {
						log.Error().Msg(fmt.Sprintf("CheckMergeRequest %v", err.Error()))
					}
					answer += a
					html = true
					newLineup, inputCommandResult := lineUp.InputCommand(chatId, arg)
					if newLineup != lineUp {
						log.Error().Msg(fmt.Sprintf("new lineup on rebase command %d", chatId))
					}
					answer += inputCommandResult.Answer
					buttons = inputCommandResult.Buttons
				default:
					newLineup, inputCommandResult := lineUp.InputCommand(chatId, arg)
					if newLineup != lineUp {
						log.Error().Msg(fmt.Sprintf("new lineup on rebase command %d", chatId))
					}
					switch inputCommandResult.Answer {
					case inputs.RebaseAcceptMessage:
						r := b.UsersMergeRequest[0]
						for _, v := range r.Changes {
							s := b.RootLineUp.NewSet(v.Dj, v.Room, v.Day, v.Hour, v.Minute, v.Duration, 0)
							answer += "added " + b.RootLineUp.PrintSet(s) + "\n"
							answer += b.RootLineUp.AddSet(s)
						}
						b.UsersMergeRequest = b.UsersMergeRequest[1:]
						b.sendMessage(r.UserId, fmt.Sprintf(MergedMessageAccepted, r.ID, user))
					case inputs.RebaseRefuseMessage:
						r := b.UsersMergeRequest[0]
						b.UsersMergeRequest = b.UsersMergeRequest[1:]
						b.sendMessage(r.UserId, fmt.Sprintf(MergedMessageRefused, r.ID, user))
					default:
						log.Error().Msg(fmt.Sprintf("unknown answer returned from inputCommand <%v>", inputCommandResult.Answer))
					}
					answer = inputCommandResult.Answer
					if len(b.UsersMergeRequest) == 0 {
						answer += " (No more merge request pending)"
					} else {
						answer += fmt.Sprintf(" (Remaining merge requests: %d)", len(b.UsersMergeRequest))
					}
					buttons = inputCommandResult.Buttons
				}
				b.Save()
			}
			//answer += fmt.Sprintf("%v", b.UsersMergeRequest)
		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
	case inputs.MergeCommand:
		if b.config.BotAllowInput || b.IsAdmin(chatId) {

			switch lineUp.CurrentInputCommand(chatId) {
			case "":
				if lineUp != b.RootLineUp {

					html = true
					answer, _ = b.compareLineUps(b.RootLineUp, lineUp)
					log.Debug().Msg(answer)
					newLineup, inputCommandResult := lineUp.InputCommand(chatId, arg)
					if newLineup != lineUp {
						log.Error().Msg(fmt.Sprintf("new lineup on merge command %d", chatId))
					}

					answer += inputCommandResult.Answer
					buttons = inputCommandResult.Buttons
				} else {
					answer = nothingToMergeMessage
				}
			default:
				newLineup, inputCommandResult := lineUp.InputCommand(chatId, arg)
				if newLineup != lineUp {
					log.Debug().Msg(fmt.Sprintf("created new lineup for user %d", chatId))
					lineUp = newLineup
					b.UsersLineUps[chatId] = newLineup
				}
				switch inputCommandResult.Answer {
				// reponse a merge
				case inputs.MergeSubmitMessage:
					mr := NewMergeRequest(b.config.Lineup.BeginningSchedule, newLineup.Changes, chatId, user, answer)
					b.CreateMergeRequest(*mr)
					delete(b.UsersLineUps, chatId)
					lineUp = b.RootLineUp
					answer += fmt.Sprintf(" (#%d)", mr.ID)

				case inputs.MergeDeleteMessage:
					delete(b.UsersLineUps, chatId)
					lineUp = b.RootLineUp
				default:
					log.Error().Msg(fmt.Sprintf("unknown answer returned from inputCommand <%v>", inputCommandResult.Answer))
				}

				answer = inputCommandResult.Answer
				buttons = inputCommandResult.Buttons
			}
			b.Save()
		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
	case inputs.InputCommand:
		if b.config.BotAllowInput || b.IsAdmin(chatId) {
			newLineup, inputCommandResult := lineUp.InputCommand(chatId, arg)
			if newLineup != lineUp {
				log.Debug().Msg(fmt.Sprintf("created new lineup for user %d", chatId))
				lineUp = newLineup
				b.UsersLineUps[chatId] = newLineup
			}
			answer = inputCommandResult.Answer
			buttons = inputCommandResult.Buttons
			b.Save()
		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
	case inputs.LogCommand:
		if b.IsAdmin(chatId) {
			if !lineUp.IsUserInputing(chatId) {
				answer += b.logs
			}
			newLineup, inputCommandResult := lineUp.InputCommand(chatId, arg)
			if newLineup != lineUp {
				err := fmt.Sprintf("ERROR: shouldnt created new lineup for user %d\n", chatId)
				log.Error().Msg(err)
				answer += err
			}
			answer += inputCommandResult.Answer
			buttons = inputCommandResult.Buttons
			b.Save()
			newUsers, totalUsers, deleted, notifications := b.users.UsersStats()
			answer += fmt.Sprintf("TotalUsers: %d new:%d deleted:%d notifs:%d", totalUsers, newUsers, deleted, notifications)
		} else {
			answer = b.defaultCommand(orig, lineUp, chatId)
		}
	default:
		answer = b.defaultCommand(orig, lineUp, chatId)
	}

	if adminMsg != "" {
		b.SendAdminsMessage(adminMsg)
	}

	if len(buttons) == 0 {
		buttons = b.GetButtonsForUser(chatId)
	}

	if lineUp != b.RootLineUp && !lineUp.IsUserInputing(chatId) {
		answer += modifiedLineUpMessage
	}

	if buttons == nil {
		log.Debug().Msg(fmt.Sprintf("buttons: %v = nil", buttons))

	} else {
		log.Debug().Msg(fmt.Sprintf("buttons: %v not nil", buttons))

	}

	if answer != "" && answer != "\n" {
		messages = append(messages, Message{
			Text:    answer,
			Buttons: buttons,
			UserID:  chatId,
			Html:    html,
		})
	} else {
		log.Warn().Msg("skipped empty message")
	}

	return splitMessages(messages)
}

func (b Bot) GetButtonsForUser(chatId int64) []string {
	lineUp := b.GetLineUpForUser(chatId)
	if lineUp.IsUserInputing(chatId) {
		return nil
	}
	buttons := []string{}

	for _, v := range b.config.Buttons {
		if b.magicRoomButton {
			if v == helpCommand {
				lastShown := b.users.GetLastShownRoom(chatId)

				if lastShown >= len(b.roomsEmoticons) {
					lastShown = 0
				}
				buttons = append(buttons, b.roomsEmoticons[lastShown])
			}
		}
		buttons = append(buttons, v)
	}
	log.Trace().Msg(fmt.Sprintf("rooms emoticons <%v>", b.roomsEmoticons))

	hasUserNotifications, err := b.users.HasUserNotifications(chatId)
	if err != nil {
		log.Warn().Msg(err.Error())
	}
	if hasUserNotifications {
		buttons = append(buttons, stopNotificationsCommand)
	} else {
		buttons = append(buttons, startNotificationsCommand)
	}

	if lineUp != b.RootLineUp {
		buttons = append(buttons, inputs.MergeCommand)
	}
	if b.IsModo(chatId) && len(b.UsersMergeRequest) != 0 {
		buttons = append(buttons, inputs.RebaseCommand)
	}
	if b.IsAdmin(chatId) {
		buttons = append(buttons, inputs.LogCommand)
	}
	return buttons
}

func (b *Bot) GroupChange(chatId int64, userString, group string) {
	msg := fmt.Sprintf("%v userString:%v group:%v", chatId, userString, group)
	if !b.users.DoesUserExists(chatId) {
		b.SendAdminsMessage(msg)
	} else {
		log.Debug().Msg(msg)
	}
}
