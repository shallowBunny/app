package inputs

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"math/rand"

	"github.com/araddon/dateparse"
	"github.com/rs/zerolog/log"
)

type test struct {
	input string
	want  InputCommandResult
}

var (
	rooms        = []string{"A", "B"}
	roomsButtons = append(rooms, cancelButton)
	days         = []string{"Fri", "Sat", "Sun", "Mon", "Tue", "Wed", "Thu"}
	daysButtons  = append(days, cancelButton)
	inputTest    = []test{
		{input: InputCommand, want: InputCommandResult{whichRoomMessage, roomsButtons, nil}},
		{input: "C", want: InputCommandResult{invalidRoom, roomsButtons, nil}},
		{input: "B", want: InputCommandResult{whichDay, daysButtons, nil}},
		{input: "Z", want: InputCommandResult{invalidDay, daysButtons, nil}},
		{input: "Sat", want: InputCommandResult{whichHourMessage, whichHourButtons, nil}},
		{input: "aaa", want: InputCommandResult{invalidHour, whichHourButtons, nil}},
		{input: "111", want: InputCommandResult{invalidHour, whichHourButtons, nil}},
		{input: "12:00", want: InputCommandResult{whichDj, whichDjButtons, nil}},
		{input: "", want: InputCommandResult{invalidDj, whichDjButtons, nil}},
		{input: "DJ", want: InputCommandResult{whichDuration, whichDurationButtons, nil}},
		{input: "tratata", want: InputCommandResult{invalidDuration, whichDurationButtons, nil}},
		{input: "120", want: InputCommandResult{`B Sat 12:00 to 14:00 DJ` + validationMsg, validationButtons, nil}},
		{input: "tratata", want: InputCommandResult{validateErrorMessage, validationButtons, nil}},
	}
	inputTestEdit = []test{
		{input: "C", want: InputCommandResult{invalidRoom, roomsButtons, nil}},
		{input: "B", want: InputCommandResult{whichDay, daysButtons, nil}},
		{input: "Z", want: InputCommandResult{invalidDay, daysButtons, nil}},
		{input: "Sat", want: InputCommandResult{whichHourMessage, []string{"12:00", cancelButton}, nil}},
		{input: "aaa", want: InputCommandResult{invalidHour, []string{"12:00", cancelButton}, nil}},
		{input: "111", want: InputCommandResult{invalidHour, []string{"12:00", cancelButton}, nil}},
		{input: "13:01", want: InputCommandResult{whichDj, append(whichDjButtons, "DJ"), nil}},
		{input: "", want: InputCommandResult{invalidDj, append(whichDjButtons, "DJ"), nil}},
		{input: "DJ", want: InputCommandResult{whichDuration, whichDurationButtons, nil}},
		{input: "tratata", want: InputCommandResult{invalidDuration, whichDurationButtons, nil}},
		{input: "120", want: InputCommandResult{`B Sat 13:01 to 15:01 DJ` + validationMsg, validationButtons, nil}},
		{input: "tratata", want: InputCommandResult{validateErrorMessage, validationButtons, nil}},
	}

	inputTestContinue = []test{
		{input: "", want: InputCommandResult{invalidDj, whichDjButtons, nil}},
		{input: "DJ2", want: InputCommandResult{whichDuration, whichDurationButtons, nil}},
		{input: "tratata", want: InputCommandResult{invalidDuration, whichDurationButtons, nil}},
		{input: "111", want: InputCommandResult{`B Sat 14:00 to 15:51 DJ2` + validationMsg, validationButtons, nil}},
		{input: "tratata", want: InputCommandResult{validateErrorMessage, validationButtons, nil}},
		{input: ContinueCommand, want: InputCommandResult{whichDj, whichDjButtons, nil}},
		{input: "DJ3", want: InputCommandResult{whichDuration, whichDurationButtons, nil}},
		{input: "tratata", want: InputCommandResult{invalidDuration, whichDurationButtons, nil}},
		{input: "10", want: InputCommandResult{`B Sat 15:51 to 16:01 DJ3` + validationMsg, validationButtons, nil}},
		{input: "tratata", want: InputCommandResult{validateErrorMessage, validationButtons, nil}},
		{input: ValidateCommand, want: InputCommandResult{validatedMessage, nil, []InputCommandResultSet{
			{
				Room:     "B",
				Dj:       "DJ",
				Day:      1,
				Hour:     12,
				Minute:   0,
				Duration: 120,
			},
			{
				Room:     "B",
				Dj:       "DJ2",
				Day:      1,
				Hour:     14,
				Minute:   0,
				Duration: 111,
			},
			{
				Room:     "B",
				Dj:       "DJ3",
				Day:      1,
				Hour:     15,
				Minute:   51,
				Duration: 10,
			},
		}}},
	}
)

func TestInputWithValidation(t *testing.T) {

	tests := append(inputTest, test{input: ValidateCommand, want: InputCommandResult{validatedMessage, nil, []InputCommandResultSet{{
		Room:     "B",
		Dj:       "DJ",
		Day:      1,
		Hour:     12,
		Minute:   0,
		Duration: 120}}}})

	i := New(days, rooms)

	for _, tc := range tests {
		got := i.InputCommand(0, tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}

	for _, tc := range tests {
		got := i.InputCommand(0, tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestInputWithCancellation(t *testing.T) {

	tests := append(inputTest, test{input: cancelCommand, want: InputCommandResult{cancelledMessage, nil, nil}})

	for ii := 1; ii < len(tests); ii++ {

		i := New(days, rooms)

		for zz, tc := range tests {
			if zz == ii {
				got := i.InputCommand(0, cancelCommand)
				want := InputCommandResult{cancelledMessage, nil, nil}
				if !reflect.DeepEqual(want, got) {
					if !reflect.DeepEqual(want.Buttons, got.Buttons) {
						t.Fatalf("Answer Buttons on %d expected: %v, got: %v", ii, want.Buttons, got.Buttons)
					}
					if !reflect.DeepEqual(want.Answer, got.Answer) {
						t.Fatalf("Answer Cancel on %d expected: <%v>, got: <%v>", ii, want.Answer, got.Answer)
					}
					if !reflect.DeepEqual(want.Sets, got.Sets) {
						t.Fatalf("Answer Cancel on %d expected: %v, got: %v", ii, want.Sets, got.Sets)
					}

					t.Fatalf("Cancel on %d expected: %v, got: %v", ii, tc.want, got)
				}
				break
			} else {
				got := i.InputCommand(0, tc.input)
				if !reflect.DeepEqual(tc.want, got) {
					t.Fatalf("expected: %v, got: %v", tc.want, got)
				}
			}
		}

		for _, tc := range tests {
			got := i.InputCommand(0, tc.input)
			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("expected: %v, got: %v", tc.want, got)
			}
		}
	}

}

func TestInputWithEdit(t *testing.T) {

	tests := append(inputTest, test{input: editCommand, want: InputCommandResult{whichRoomMessage, roomsButtons, nil}})

	tests2 := append(inputTestEdit, test{input: ValidateCommand, want: InputCommandResult{validatedMessage, nil, []InputCommandResultSet{{
		Room:     "B",
		Dj:       "DJ",
		Day:      1,
		Hour:     13,
		Minute:   1,
		Duration: 120}}}})

	i := New(days, rooms)

	for _, tc := range tests {
		got := i.InputCommand(0, tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
	for _, tc := range tests2 {
		got := i.InputCommand(0, tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}

}

func TestInputWithContinue(t *testing.T) {

	tests := append(inputTest, test{input: ContinueCommand, want: InputCommandResult{whichDj, whichDjButtons, nil}})

	tests2 := inputTestContinue

	/*
		append(inputTestContinue, test{input: ValidateCommand, want: InputCommandResult{validatedMessage, nil, []InputCommandResultSet{{
			Room:     "B",
			Dj:       "DJ",
			Day:      1,
			Hour:     13,
			Minute:   1,
			Duration: 120}}}})
	*/

	i := New(days, rooms)

	for ii, tc := range tests {
		got := i.InputCommand(0, tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%d expected: %v, got: %v", ii, tc.want, got)
		}
	}
	for ii, tc := range tests2 {
		got := i.InputCommand(0, tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%d expected: %v, got: %v", ii, tc.want, got)
		}
	}

}

func TestRandomInputWithContinue(t *testing.T) {

	randomTests := []test{
		{input: InputCommand, want: InputCommandResult{whichRoomMessage, roomsButtons, nil}},
		{input: "B", want: InputCommandResult{whichDay, daysButtons, nil}},
		{input: "Fri", want: InputCommandResult{whichHourMessage, whichHourButtons, nil}},
		{input: "12:00", want: InputCommandResult{whichDj, whichDjButtons, nil}},
		{input: "DJ", want: InputCommandResult{whichDuration, whichDurationButtons, nil}},
		{input: "120", want: InputCommandResult{`B Fri 12:00 to 14:00 DJ` + validationMsg, validationButtons, nil}},
	}

	tt, err := dateparse.ParseLocal("Fri Jun 7 2024 14:00:00")
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = tt
	i := New(days, rooms)

	for ii, tc := range randomTests {
		got := i.InputCommand(0, tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%d expected: %v, got: %v", ii, tc.want, got)
		}
		log.Debug().Msg(fmt.Sprintf("%d <%v>", ii, got))
	}

	for nb := 0; nb < 100; nb++ {
		DJInput := fmt.Sprintf("DJ number %d", nb)
		duration := 1 + rand.Int()%DurationMax
		fin := tt.Add(time.Duration(duration) * time.Minute)

		Result := fmt.Sprintf("B %v to %v DJ number %d%v", tt.Format("Mon 15:04"), fin.Format("15:04"), nb, validationMsg)
		tt = tt.Add(time.Duration(duration) * time.Minute)
		ii2 := []test{{input: ContinueCommand, want: InputCommandResult{whichDj, whichDjButtons, nil}},
			{input: DJInput, want: InputCommandResult{whichDuration, whichDurationButtons, nil}},
			{input: strconv.Itoa(duration), want: InputCommandResult{Result, validationButtons, nil}},
		}

		for ii, tc := range ii2 {
			got := i.InputCommand(0, tc.input)
			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("%d expected: %v, got: %v", ii, tc.want, got)
			}
		}
	}

}
