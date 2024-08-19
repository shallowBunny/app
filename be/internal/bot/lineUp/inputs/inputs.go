package inputs

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	Undefined = iota
	NotInputingState
	ChoosingRoom
	MergingState
	ChoosingDay
	ChoosingHour
	EnteringSet
	EnteringDuration
	Validate
	MergeStep
	RebaseStep
	logStep
	closed      = "closed"
	unknown     = "?"
	Duration60  = "1h"
	Duration90  = "1.5"
	Duration120 = "2h"
	Duration150 = "2.5"
	Duration180 = "3h"
	Duration210 = "3.5"
	Duration240 = "4h"
	DurationMax = 60 * 10
)

type RoomSchedule struct {
	Room     string
	Schedule []string
	Delta    time.Duration
}

type Step int

type State struct {
	Step              Step
	Day               int
	Min               int
	Hour              int
	Dj                string
	Duration          int
	Room              string
	Inputs            []InputCommandResultSet
	WhichInputCommand string
}

type Inputs struct {
	States           map[int64]*State
	Days             []string
	Rooms            []string
	WhichRoomButtons []string
	WhichDaysButtons []string
}

func (i Inputs) CurrentInputCommand(chatID int64) string {
	v, ok := i.States[chatID]
	if !ok {
		return ""
	} else {
		return v.WhichInputCommand // NotInputingState
	}
}

func (i Inputs) IsUserInputing(chatID int64) bool {
	return i.CurrentInputCommand(chatID) != ""
}

func (i Inputs) IsUserInLogs(chatID int64) bool {
	return i.CurrentInputCommand(chatID) == LogCommand
}

type InputCommandResultSet struct {
	Room     string
	Dj       string
	Day      int
	Hour     int
	Minute   int
	Duration int
}

type InputCommandResult struct {
	Answer  string
	Buttons []string
	Sets    []InputCommandResultSet
}

const (
	whichRoomMessage       = "Which room?"
	invalidRoom            = "Invalid input, please click a button to choose the room"
	whichDay               = "Which day?"
	invalidDay             = "Invalid input, please click a button to choose the day"
	whichHourMessage       = `Which hour? i.e "21" or "21 30"`
	invalidHour            = `Invalid input, please enter something like "11" or "11 30"`
	whichDj                = "Enter the Dj name"
	invalidDj              = "Invalid input, please enter the Dj name"
	whichDuration          = "which duration for this set? (click button or input duration in minutes)"
	invalidDurationTooLong = "You entered %d but max duration is %d minutes, please try again"
	invalidDuration        = "Invalid input, please enter Duration in minutes"
	validatedMessage       = "Validated, thanks"
	cancelledMessage       = "Cancelled, no changes in lineup where made"
	validateErrorMessage   = "Invalid input"
	internalErrorMessage   = "Internal error"
	mergeMessage           = `
Click Submit to submit your changes to moderation
Edit to make more changes
Delete to delete all your changes`
	MergeSubmitMessage = "Merge request sent to moderation, thanks!"
	MergeDeleteMessage = "Cancelled merge request and deleted all your changes"
	MergeEditMessage   = "Cancelled merge request, you can keep editing your changes"
	logMessage         = "Starting log streaming."
	RebaseCommand      = "rebase"
	cancelButton       = "🔴"
	stopLoggingMessage = "Stopped log streaming."
)

var (
	logButtons           = []string{cancelButton}
	whichDjButtons       = []string{cancelButton} //[]string{closed, unknown}
	whichDurationButtons = []string{Duration60, Duration90, Duration120, Duration150, Duration180, Duration210, Duration240, cancelButton}
	MergeSubmitCommand   = "Submit"
	MergeEditCommand     = "edit"
	MergeDeleteCommand   = "delete"
	mergeButtons         = []string{MergeSubmitCommand, MergeEditCommand, MergeDeleteCommand}
	whichDurationValue   = []int{60, 90, 120, 150, 180, 210, 240, 0}
	ValidateCommand      = "validate"
	ContinueCommand      = "continue"
	editCommand          = "edit"
	cancelCommand        = cancelButton
	InputCommand         = "input"
	LogCommand           = "log"
	MergeCommand         = "merge"
	WhichRoomButtons     = []string{}
	validationMsg        = "\n\nClick the validate button to confirm\nContinue to enter an extra set right afte this one\nEdit to change the last entered set\n" + cancelButton + " to cancel"
	validationButtons    = []string{ValidateCommand, ContinueCommand, editCommand, cancelCommand}
	whichHourButtons     = []string{cancelButton}

	RebaseAcceptCommand = "accept"
	RebaseRefuseCommand = "refuse"

	RebaseAcceptMessage = "Merge request rebased on master"
	RebaseRefuseMessage = "Merge request refused"

	RebaseMessage = "\nAccept or refuse this MR"
	rebaseButtons = []string{RebaseAcceptCommand, RebaseRefuseCommand}
	emptyButtons  = []string{}
)

func (i *Inputs) printSet(Room string, Day int, Hour int, Min int, Hour2 int, Min2 int, Dj string) string {
	DayString := i.Days[Day%len(i.Days)]

	//res += v.start.Format("MonDay") + " " + printTime(v.start) + " to " + printTime(v.end) + " " + v.Dj

	return fmt.Sprintf("%v %v %.2d:%.2d to %.2d:%.2d %v", Room, DayString, Hour, Min, Hour2, Min2, Dj)
}

// msgText , buttons, removeKeyboard, msgAdMin
func (i *Inputs) InputCommand(chatID int64, commandOrArg string) InputCommandResult {

	v, ok := i.States[chatID]
	if !ok {
		state := &State{Step: NotInputingState,
			WhichInputCommand: "",
		}
		i.States[chatID] = state
		v = state
	}

	switch v.Step {

	case NotInputingState:
		switch commandOrArg {
		case LogCommand:
			i.States[chatID] = &State{Step: logStep,
				Min:               -1,
				Hour:              -1,
				WhichInputCommand: LogCommand,
			}
			return InputCommandResult{logMessage, logButtons, nil}

		case InputCommand:
			i.States[chatID] = &State{Step: ChoosingRoom,
				Min:               -1,
				Hour:              -1,
				WhichInputCommand: InputCommand,
			}
			return InputCommandResult{whichRoomMessage, i.WhichRoomButtons, nil}
		case MergeCommand:
			i.States[chatID] = &State{Step: MergeStep,
				WhichInputCommand: MergeCommand,
			}
			return InputCommandResult{mergeMessage, mergeButtons, nil}
		case RebaseCommand:
			i.States[chatID] = &State{Step: RebaseStep,
				WhichInputCommand: RebaseCommand,
			}
			return InputCommandResult{RebaseMessage, rebaseButtons, nil}
		default:
			log.Error().Msg(fmt.Sprintf("InputCommand: %d <%v>", chatID, commandOrArg))
			return InputCommandResult{internalErrorMessage, nil, nil}
		}
	case ChoosingRoom:

		if commandOrArg == cancelButton {
			i.emptyState(chatID)
			return InputCommandResult{cancelledMessage, nil, nil}
		}

		foundRoom := false
		for _, v := range i.Rooms {
			if v == commandOrArg {
				foundRoom = true
				continue
			}
		}

		if !foundRoom {
			return InputCommandResult{invalidRoom, i.WhichRoomButtons, nil}
		}

		i.States[chatID].Step = ChoosingDay
		i.States[chatID].Room = commandOrArg
		return InputCommandResult{whichDay, i.WhichDaysButtons, nil}
	case ChoosingDay:

		if commandOrArg == cancelButton {
			i.emptyState(chatID)
			return InputCommandResult{cancelledMessage, nil, nil}
		}

		found := false
		for index, v := range i.Days {
			if v == commandOrArg {
				found = true
				i.States[chatID].Day = index
				continue
			}
		}
		if !found {
			return InputCommandResult{invalidDay, i.WhichDaysButtons, nil}
		}
		i.States[chatID].Step = ChoosingHour
		if i.States[chatID].Min != -1 && i.States[chatID].Hour != -1 {
			whichHoursButton := []string{fmt.Sprintf("%.2d:%.2d", i.States[chatID].Hour, i.States[chatID].Min), cancelButton}
			return InputCommandResult{whichHourMessage, whichHoursButton, nil}
		}
		return InputCommandResult{whichHourMessage, whichHourButtons, nil}
	case ChoosingHour:

		if commandOrArg == cancelButton {
			i.emptyState(chatID)
			return InputCommandResult{cancelledMessage, nil, nil}
		}

		Hour := -1
		Min := 0
		commandOrArg = strings.ReplaceAll(commandOrArg, ":", " ")
		fmt.Sscanf(commandOrArg, "%d %d", &Hour, &Min)
		if Hour == -1 || Min == -1 || Hour > 23 {
			if i.States[chatID].Min != -1 && i.States[chatID].Hour != -1 {
				whichHoursButton := []string{fmt.Sprintf("%.2d:%.2d", i.States[chatID].Hour, i.States[chatID].Min), cancelButton}
				return InputCommandResult{invalidHour, whichHoursButton, nil}
			}
			return InputCommandResult{invalidHour, whichHourButtons, nil}
		}
		i.States[chatID].Min = Min
		i.States[chatID].Hour = Hour
		i.States[chatID].Step = EnteringSet

		DjButtons := whichDjButtons
		if i.States[chatID].Dj != "" {
			DjButtons = append(DjButtons, i.States[chatID].Dj)
		}
		return InputCommandResult{whichDj, DjButtons, nil}

	case EnteringSet:

		if commandOrArg == cancelButton {
			i.emptyState(chatID)
			return InputCommandResult{cancelledMessage, nil, nil}
		}

		if commandOrArg == "" {
			DjButtons := whichDjButtons
			if i.States[chatID].Dj != "" {
				DjButtons = append(DjButtons, i.States[chatID].Dj)
			}
			return InputCommandResult{invalidDj, DjButtons, nil}
		}
		i.States[chatID].Step = EnteringDuration
		i.States[chatID].Dj = commandOrArg
		return InputCommandResult{whichDuration, whichDurationButtons, nil}

	case EnteringDuration:

		if commandOrArg == cancelButton {
			i.emptyState(chatID)
			return InputCommandResult{cancelledMessage, nil, nil}
		}

		Duration := -1
		for i, v := range whichDurationButtons {
			if v == commandOrArg {
				Duration = whichDurationValue[i]
			}
		}
		if Duration == -1 {
			fmt.Sscanf(commandOrArg, "%d", &Duration)
		}

		if Duration <= 0 || Duration > DurationMax {
			msg := invalidDuration
			if Duration > DurationMax {
				msg = fmt.Sprintf(invalidDurationTooLong, Duration, DurationMax)
			}
			return InputCommandResult{msg, whichDurationButtons, nil}
		}

		i.States[chatID].Duration = Duration

		Hour2 := (i.States[chatID].Hour + Duration/60)
		mm := (i.States[chatID].Min + Duration%60)
		if mm >= 60 {
			Hour2++
		}
		Min2 := mm % 60

		Hour2 = Hour2 % 24

		//log.Info().Msg(fmt.Sprintf("Hour: %v Hour+Duration/60: %v Duration=%v Duration/60=%v", i.States[chatID].Hour, i.States[chatID].Hour+Duration/60, Duration, Duration/60))

		printedSet := i.printSet(i.States[chatID].Room, i.States[chatID].Day, i.States[chatID].Hour, i.States[chatID].Min, Hour2, Min2, i.States[chatID].Dj)

		//log.Info().Msg(printedSet)

		//log.Info().Msg(fmt.Sprintf("%v - '%d %.2d:%.2d %d %v'", i.States[chatID].Room, i.States[chatID].Day, i.States[chatID].Hour, i.States[chatID].Min, i.States[chatID].Duration, i.States[chatID].Dj))

		i.States[chatID].Step = Validate
		return InputCommandResult{printedSet + validationMsg, validationButtons, nil}

	case MergeStep:
		switch commandOrArg {
		case MergeSubmitCommand:
			i.emptyState(chatID)
			return InputCommandResult{MergeSubmitMessage, nil, nil}
		case MergeDeleteCommand:
			i.emptyState(chatID)
			return InputCommandResult{MergeDeleteMessage, nil, nil}
		case MergeEditCommand:
			i.emptyState(chatID)
			return InputCommandResult{MergeEditMessage, nil, nil}
		default:
			return InputCommandResult{mergeMessage, mergeButtons, nil}
		}
	case logStep:
		i.emptyState(chatID)
		return InputCommandResult{stopLoggingMessage, nil, nil}
	case RebaseStep:
		switch commandOrArg {
		case RebaseAcceptCommand:
			i.emptyState(chatID)
			return InputCommandResult{RebaseAcceptMessage, nil, nil}
		case RebaseRefuseCommand:
			i.emptyState(chatID)
			return InputCommandResult{RebaseRefuseMessage, nil, nil}
		default:
			return InputCommandResult{RebaseMessage, rebaseButtons, nil}
		}

	case Validate:
		switch commandOrArg {
		case ValidateCommand:
			set := InputCommandResultSet{
				Room:     i.States[chatID].Room,
				Dj:       i.States[chatID].Dj,
				Day:      i.States[chatID].Day,
				Hour:     i.States[chatID].Hour,
				Minute:   i.States[chatID].Min,
				Duration: i.States[chatID].Duration,
			}
			i.States[chatID].Inputs = append(i.States[chatID].Inputs, set)
			res := i.States[chatID].Inputs
			i.emptyState(chatID)
			return InputCommandResult{validatedMessage, nil, res}
		case cancelCommand:
			i.emptyState(chatID)
			return InputCommandResult{cancelledMessage, nil, nil}
		case editCommand:
			i.States[chatID].Step = ChoosingRoom
			return InputCommandResult{whichRoomMessage, i.WhichRoomButtons, nil}
		case ContinueCommand:
			set := InputCommandResultSet{
				Room:     i.States[chatID].Room,
				Dj:       i.States[chatID].Dj,
				Day:      i.States[chatID].Day,
				Hour:     i.States[chatID].Hour,
				Minute:   i.States[chatID].Min,
				Duration: i.States[chatID].Duration,
			}
			i.States[chatID].Inputs = append(i.States[chatID].Inputs, set)
			i.States[chatID].Dj = ""
			i.States[chatID].Min += i.States[chatID].Duration % 60
			if i.States[chatID].Min > 59 {
				i.States[chatID].Hour += 1
				i.States[chatID].Min -= 60
			}
			i.States[chatID].Hour += i.States[chatID].Duration / 60
			if i.States[chatID].Hour > 23 {
				i.States[chatID].Hour -= 24
				i.States[chatID].Day++
			}
			i.States[chatID].Step = EnteringSet

			return InputCommandResult{whichDj, whichDjButtons, nil}
		default:
			return InputCommandResult{validateErrorMessage, validationButtons, nil}
		}

	default:
		log.Error().Msg(internalErrorMessage)
		i.emptyState(chatID)
		return InputCommandResult{internalErrorMessage, nil, nil}
	}

}

func (i *Inputs) emptyState(chatID int64) {
	_, ok := i.States[chatID]
	if ok {
		i.States[chatID].Step = NotInputingState
		i.States[chatID].WhichInputCommand = ""
		i.States[chatID].Min = -1
		i.States[chatID].Hour = -1
	} else {
		log.Error().Msg(fmt.Sprintf("emptyState on non existing state %v", chatID))
	}
}

func (i Inputs) printEditing(chatID int64) (string, []string) {

	/*
		editTime := i.States[chatID].time
		text := i.States[chatID].RoomSchedule.Room + " " + editTime.Format("Mon") + "-" + printTime(editTime) + " " + printTime(editTime.Add(i.States[chatID].RoomSchedule.Delta)) + " enter the Dj name or use one of the buttons"

		buttons := []string{}

		nbScheds := len(i.States[chatID].RoomSchedule.Schedule)
		if nbScheds != 0 {
			schedules := i.States[chatID].RoomSchedule.Schedule
			lastEnteredDj := schedules[nbScheds-1]
			buttons = append(buttons, lastEnteredDj)
		}


	*/

	text := "enter the Dj name or use one of the buttons"

	buttons := []string{}

	buttons = append(buttons, closed)
	buttons = append(buttons, unknown)
	/*
		var numericKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(buttons...))
		msg.ReplyMarkup = numericKeyboard
	*/
	return text, buttons
}

func New(Days, Rooms []string) Inputs {
	return Inputs{
		States:           make(map[int64]*State),
		Days:             Days,
		Rooms:            Rooms,
		WhichRoomButtons: append(Rooms, cancelButton),
		WhichDaysButtons: append(Days, cancelButton),
	}
}
