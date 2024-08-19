package lineUp

import (
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shallowBunny/app/be/internal/bot/lineUp/inputs"
)

type test struct {
	input string
	want  string
}

var (
	roomA = "roomA"
	dj    = "pierre"

	roomTurm = "🗼 Turmbühne"
	roomTanz = "🏜️ Tanzwüste"
	roomSonn = "🌞 Sonnendeck"

	rooms = []string{roomA, roomTurm, roomTanz, roomSonn}

	roomSchedule2 = map[string][]string{
		roomA:    {"0 23:00 120 0 MADmoiselle"},
		roomTurm: {"3 02:00 120 1 Animal Trainer"},
		roomTanz: {"2 03:00 120 1 Ava Irandoost"},
		roomSonn: {"4 18:00 120 1 Bassphilia"},
	}

	roomSchedule = map[string][]string{
		roomA: {"0 23:00 120 0 MADmoiselle"},
	}
)

func TestLineUpInput(t *testing.T) {

	tt := time.Now()
	currentTime := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
	startTime := currentTime

	lu := New(startTime, 3, true, true, false, nil, rooms, roomSchedule)

	got := lu.Dump()
	want := `roomA:
- '0 23:00 120 0 MADmoiselle'
`
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected: \n<%v>, got: \n<%v>", want, got)
	}

	l, r := lu.InputCommand(0, inputs.InputCommand)
	log.Debug().Msg(r.Answer)
	if l != lu {
		t.Fatalf("New lineup should not be created")
	}
	l, r = lu.InputCommand(0, roomA)
	if l != lu {
		t.Fatalf("New lineup should not be created")
	}
	log.Debug().Msg(r.Answer)
	l, r = lu.InputCommand(0, r.Buttons[0])
	if l != lu {
		t.Fatalf("New lineup should not be created")
	}
	log.Debug().Msg(r.Answer)
	l, r = lu.InputCommand(0, "23:30")
	if l != lu {
		t.Fatalf("New lineup should not be created")
	}
	log.Debug().Msg(r.Answer)
	l, r = lu.InputCommand(0, dj)
	if l != lu {
		t.Fatalf("New lineup should not be created")
	}
	log.Debug().Msg(r.Answer)
	l, r = lu.InputCommand(0, "180")
	if l != lu {
		t.Fatalf("New lineup should not be created")
	}
	log.Debug().Msg(r.Answer)
	l, r = lu.InputCommand(0, inputs.ValidateCommand)
	if l == lu {
		t.Fatalf("No new lineup created")
	}
	log.Debug().Msg(r.Answer)

	// check premier lineup non modifie
	got = lu.Dump()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected: \n<%v>, got: \n<%v>", want, got)
	}

	// check deuxieme lineup modifie
	got = l.Dump()
	want = `roomA:
- '0 23:30 180 0 pierre'
`
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected: \n<%v>, got: \n<%v>", want, got)
	}

	// check events for both lineup

	got = lu.Events(startTime.Add(time.Hour * 23))
	want = "MADmoiselle started in roomA\n"
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected: \n<%v>, got: \n<%v>", want, got)
	}

	got = l.Events(startTime.Add(time.Hour*23 + 29*time.Minute))
	want = ""
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected: \n<%v>, got: \n<%v>", want, got)
	}

	got = l.Events(startTime.Add(time.Hour*23 + 30*time.Minute))
	want = "pierre started in roomA\n"
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected: \n<%v>, got: \n<%v>", want, got)
	}

}

func TestFindRoom(t *testing.T) {
	roomSchedule = map[string][]string{
		roomA: {"0 23:00 120 0 MADmoiselle", "0 23:00 120 0 Mad max"},
	}
	startTime := time.Now().Add(time.Hour)
	lu := New(startTime, 3, true, true, false, nil, rooms, roomSchedule2)

	inputTest := []test{
		{input: "Tanzwuste", want: roomTanz},
		{input: "Caca", want: ""},
		{input: "sonnenDuck", want: roomSonn},
	}

	for _, tc := range inputTest {
		_, got := lu.FindRoom(tc.input, 3)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func TestFindDJ(t *testing.T) {
	roomSchedule = map[string][]string{
		roomA: {"3 18:00 60 0 Robyn Schulkowsky & Gebrüder Teichmann"},
	}
	startTime := time.Now().Add(time.Hour)
	lu := New(startTime, 3, true, true, false, nil, rooms, roomSchedule)

	day := time.Now().Add(time.Hour * 24 * 3).Format("Monday")

	inputTest := []test{
		{input: "Robyn", want: searchedMessage + "✅ Robyn Schulkowsky & Gebrüder Teichmann is playing " + day + " at 18:00 in roomA\n"},
		{input: "Cacaboudin", want: searchedMessage + "Not found. 😔\n"},
		{input: "Robin", want: searchedMessage + "✅ Robyn Schulkowsky & Gebrüder Teichmann is playing " + day + " at 18:00 in roomA\n"},
	}

	for _, tc := range inputTest {
		got := lu.FindDJ(tc.input, startTime)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: <%v>, got: <%v>", tc.want, got)
		}
	}
}
