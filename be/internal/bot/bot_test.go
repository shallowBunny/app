package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/shallowBunny/app/be/internal/bot/config"
	DaoDb "github.com/shallowBunny/app/be/internal/bot/dao/daoDb"
	DaoMem "github.com/shallowBunny/app/be/internal/bot/dao/daoMem"
	"github.com/shallowBunny/app/be/internal/bot/lineUp/inputs"
	"github.com/tj/assert"
)

type test struct {
	input time.Time
	want  string
}

const (
	adminID int64 = -123
)

func TestSerialisation(t *testing.T) {
	config, err := config.New("../../configs/bot_test.yaml", false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	tt := time.Now()
	log.Debug().Msg("TestSerialisation")

	currentTime := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
	config.Lineup.BeginningSchedule = currentTime

	redisclient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	dao := DaoDb.New("apiToken", redisclient)

	dao.DeleteBot(config.Lineup.BeginningSchedule)

	_ = dao
	dao2 := DaoMem.New()
	_ = dao2

	bot := New(dao2, config)

	log.Debug().Msg("TestSerialisation 1")

	bot.channel = nil
	bot2 := New(dao2, config)
	bot2.channel = nil

	log.Debug().Msg("TestSerialisation 2")
	log.Error().Msg(fmt.Sprintf("<%v>", bot.RootLineUp))

	log.Debug().Msg("TestSerialisation 3")

	err = dao.DeleteBot(config.Lineup.BeginningSchedule)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("<%v>", err.Error()))
	}

}

func TestEvents(t *testing.T) {

	config, err := config.New("../../configs/bot_test.yaml", false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	tt := time.Now()
	currentTime := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
	config.Lineup.BeginningSchedule = currentTime

	log.Debug().Msg(fmt.Sprintf("using currentTime = %v", currentTime))

	dao := DaoMem.New()

	bot := New(dao, config)

	currentTime = currentTime.Add(24 * time.Hour)

	inputTest := []test{
		{input: currentTime,
			want: ""},
		{input: currentTime.Add(time.Hour * 1),
			want: "A started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 2),
			want: "B started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 3),
			want: "E started in 🔨\nC in 🍵\n"},
		{input: currentTime.Add(time.Hour * 4),
			want: "D started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 5),
			want: ""},
		{input: currentTime.Add(time.Hour * 6),
			want: "F started in 🔨\n"},
	}

	for i, tc := range inputTest {
		got := bot.RootLineUp.Events(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%v expected: <%v>, got: <%v>", i, tc.want, got)
		}
	}

}

// go test -run TestInputForEvents ./... -v

func createBotForTestInputMergeAndRebase(config *config.Config, userID int64, currentTime time.Time) *Bot {
	dao := DaoMem.New()
	bot := New(dao, config)
	bot.channel = nil
	inputCommands := []string{inputs.InputCommand, "🍵", currentTime.Format("Mon"), "2:30", "DJ FART", "90", inputs.ValidateCommand}
	for _, tc := range inputCommands {
		answer := bot.ProcessCommand(userID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}
	return bot
}

func TestInputMergeAndRebase(t *testing.T) {

	t.Log("Test started")

	// create config object
	config, err := config.New("../../configs/bot_test.yaml", false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	tt := time.Now()
	currentTime := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
	config.Lineup.BeginningSchedule = currentTime
	log.Debug().Msg(fmt.Sprintf("using currentTime = %v", currentTime))
	var userID int64 = 123
	currentTime = currentTime.Add(24 * time.Hour)
	dumpBotInitial := `🔨:
- '1 03:00 180 [] E'
- '1 06:00 180 [] F'
🍵:
- '1 01:00 60 [] A'
- '1 02:00 60 [] B'
- '1 03:00 60 [] C'
- '1 04:00 60 [] D'
`
	// create bot for config object, with an input command
	bot := createBotForTestInputMergeAndRebase(config, userID, currentTime)

	// dump root Linup
	got := bot.RootLineUp.Dump()
	if !reflect.DeepEqual(dumpBotInitial, got) {
		t.Fatalf("expected: <%v>, got: <%v>", dumpBotInitial, got)
	}

	// test the events on the root lineup
	inputTest := []test{
		{input: currentTime,
			want: ""},
		{input: currentTime.Add(time.Hour * 1),
			want: "A started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 2),
			want: "B started in 🍵\n"},
		{input: currentTime.Add(time.Hour*2 + time.Minute*29),
			want: ""},
		{input: currentTime.Add(time.Hour*2 + time.Minute*30),
			want: ""},
		{input: currentTime.Add(time.Hour * 3),
			want: "E started in 🔨\nC in 🍵\n"},
		{input: currentTime.Add(time.Hour * 4),
			want: "D started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 5),
			want: ""},
		{input: currentTime.Add(time.Hour * 6),
			want: "F started in 🔨\n"},
	}
	for i, tc := range inputTest {
		got := bot.RootLineUp.Events(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("%v expected: <%v>, got: <%v>", i, tc.want, got)
		}
	}

	// test the events on the user lineup
	lu := bot.GetLineUpForUser(userID)
	if lu == bot.RootLineUp {
		t.Fatalf("no new lineup created for user after input")
	}
	inputTest2 := []test{
		{input: currentTime,
			want: ""},
		{input: currentTime.Add(time.Hour * 1),
			want: "A started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 2),
			want: ""},
		{input: currentTime.Add(time.Hour*2 + time.Minute*29),
			want: ""},
		{input: currentTime.Add(time.Hour*2 + time.Minute*30),
			want: "DJ FART started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 3),
			want: "E started in 🔨\n"},
		{input: currentTime.Add(time.Hour * 5),
			want: "D started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 6),
			want: "F started in 🔨\n"},
	}
	for i, tc := range inputTest2 {
		got := lu.Events(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("index: %v expected: <%v>, got: <%v>", i, tc.want, got)
		}
	}

	// refait un nouveau bot pour retenter le truc apres un rebase
	bot = createBotForTestInputMergeAndRebase(config, userID, currentTime)

	// merge par user
	inputCommands := []string{inputs.MergeCommand, inputs.MergeSubmitCommand}
	for _, tc := range inputCommands {
		answer := bot.ProcessCommand(userID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}
	// rebase par admin
	inputCommands2 := []string{inputs.RebaseCommand, inputs.RebaseAcceptCommand}
	for _, tc := range inputCommands2 {
		answer := bot.ProcessCommand(adminID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}

	dumpBotModified := `🔨:
- '1 03:00 180 [] E'
- '1 06:00 180 [] F'
🍵:
- '1 01:00 60 [] A'
# hole: 02:00 to 02:30
- '1 02:30 90 [] DJ FART'
- '1 04:00 60 [] D'
`

	// dump the root lineUp
	got = bot.RootLineUp.Dump()
	if !reflect.DeepEqual(dumpBotModified, got) {
		t.Fatalf("expected: <%v>, got: <%v>", dumpBotModified, got)
	}

	// test the root events
	for i, tc := range inputTest2 {
		got := bot.RootLineUp.Events(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("index: %v expected: <%v>, got: <%v>", i, tc.want, got)
		}
	}

}

func TestInputMultipleMergeAndRebase(t *testing.T) {

	// create config object
	config, err := config.New("../../configs/bot_test.yaml", false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	tt := time.Now()
	currentTime := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
	config.Lineup.BeginningSchedule = currentTime
	log.Debug().Msg(fmt.Sprintf("using currentTime = %v", currentTime))
	var userID int64 = 123
	currentTime = currentTime.Add(24 * time.Hour)

	// create bot for config object, with an input command
	dao := DaoMem.New()
	bot := New(dao, config)
	bot.channel = nil

	// 1er input par user
	inputCommands := []string{inputs.InputCommand, "🍵", currentTime.Format("Mon"), "2:30", "DJ FART", "90", inputs.ValidateCommand,
		inputs.InputCommand, "🍵", currentTime.Format("Mon"), "5:00", "DJ FART 2", "120", inputs.ValidateCommand,
	}
	for _, tc := range inputCommands {
		answer := bot.ProcessCommand(userID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}

	// 1er merge par user
	mergeByUsersCommands := []string{inputs.MergeCommand, inputs.MergeSubmitCommand}
	for _, tc := range mergeByUsersCommands {
		answer := bot.ProcessCommand(userID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}

	// rebase par admin
	rebaseCommands := []string{inputs.RebaseCommand, inputs.RebaseAcceptCommand}
	for _, tc := range rebaseCommands {
		answer := bot.ProcessCommand(adminID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}

	// 2eme input par user
	inputCommands = []string{inputs.InputCommand, "🍵", currentTime.Format("Mon"), "7:00", "DJ FART 3", "60", inputs.ValidateCommand,
		inputs.InputCommand, "🍵", currentTime.Format("Mon"), "8:00", "DJ FART 4", "60", inputs.ValidateCommand,
	}
	for _, tc := range inputCommands {
		answer := bot.ProcessCommand(userID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}

	// 2eme merge par user
	for _, tc := range mergeByUsersCommands {
		answer := bot.ProcessCommand(userID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}
	// 2eme rebase par admin
	for _, tc := range rebaseCommands {
		answer := bot.ProcessCommand(adminID, tc, "test")
		log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
	}

	dumpBotModified := `🔨:
- '1 03:00 180 [] E'
- '1 06:00 180 [] F'
🍵:
- '1 01:00 60 [] A'
# hole: 02:00 to 02:30
- '1 02:30 90 [] DJ FART'
- '1 04:00 60 [] D'
- '1 05:00 120 [] DJ FART 2'
- '1 07:00 60 [] DJ FART 3'
- '1 08:00 60 [] DJ FART 4'
`

	// dump the root lineUp
	if bot.GetLineUpForUser(userID) != bot.GetLineUpForUser(adminID) {
		t.Fatalf("user and admin lineup are different")
	}

	got := bot.GetLineUpForUser(userID).Dump()
	if !reflect.DeepEqual(dumpBotModified, got) {
		t.Fatalf("expected: <%v>, got: <%v>", dumpBotModified, got)
	}

	inputTest2 := []test{
		{input: currentTime,
			want: ""},
		{input: currentTime.Add(time.Hour * 1),
			want: "A started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 2),
			want: ""},
		{input: currentTime.Add(time.Hour*2 + time.Minute*29),
			want: ""},
		{input: currentTime.Add(time.Hour*2 + time.Minute*30),
			want: "DJ FART started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 3),
			want: "E started in 🔨\n"},
		{input: currentTime.Add(time.Hour * 4),
			want: "D started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 5),
			want: "DJ FART 2 started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 6),
			want: "F started in 🔨\n"},
		{input: currentTime.Add(time.Hour * 7),
			want: "DJ FART 3 started in 🍵\n"},
		{input: currentTime.Add(time.Hour * 8),
			want: "DJ FART 4 started in 🍵\n"},
	}

	// test the root events
	for i, tc := range inputTest2 {
		got := bot.RootLineUp.Events(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("index: %v expected: <%v>, got: <%v>", i, tc.want, got)
		}
	}

}

func TestUpdateLineUp(t *testing.T) {

	return
	// Create a new Gin router for testing
	router := gin.Default()

	// Create an instance of your Bot struct
	c, err := config.New("../../configs/bot_test.yaml", false)
	if err != nil {
		t.Fatalf(err.Error())
	}

	tt := time.Now()

	currentTime := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
	c.Lineup.BeginningSchedule = currentTime

	bot := New(DaoMem.New(), c)
	bot.channel = nil

	// Register the UpdateLineUp handler with the router
	router.PUT("/update-lineup", bot.UpdateLineUp)

	// Create a sample Lineup payload
	lineup := config.Lineup{
		Rooms: []string{
			"⌛ Zeitmaschine",
			"🌾 Mühle",
			"🏎 Turbo Tüff",
			"🚕 Furore",
			"🚚 Rave Rikscha",
			"🤡 Jesterfield",
		},
		Sets: map[string][]config.Set{
			"⌛ Zeitmaschine": {
				{
					Day:      1,
					Duration: 120,
					Dj:       "DJ Awesome",
					Hour:     22,
					Minute:   30,
				},
			},
		},
	}

	// Convert the Lineup struct to JSON
	jsonPayload, err := json.Marshal(lineup)
	if err != nil {
		t.Fatalf("Failed to marshal lineup: %v", err)
	}

	// Create a new HTTP request with the JSON payload
	req, err := http.NewRequest(http.MethodPut, "/update-lineup", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the response
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check if the status code is 200 OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Check the response body
	expectedResponse := `{"message":"Lineup updated successfully"}`
	assert.JSONEq(t, expectedResponse, w.Body.String())
}
