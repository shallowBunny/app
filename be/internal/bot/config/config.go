package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/araddon/dateparse"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Meta struct {
	AboutBigIcon              string `json:"aboutBigIcon"`
	AboutShowShallowBunnyIcon bool   `json:"aboutShowShallowBunnyIcon"`
	AboutShowSisyDuckIcon     bool   `json:"aboutShowSisyDuckIcon"`
	NowBotUrl                 string `json:"nowBotUrl"`
	NowMapImage               string `json:"nowMapImage"`
	NowShowDataSourceAd       bool   `json:"nowShowDataSourceAd"`
	NowShowShallowBunnyAd     bool   `json:"nowShowShallowBunnyAd"`
	NowShowSisyDuckAd         bool   `json:"nowShowSisyDuckAd"`
	NowShowPleaseSendData     bool   `json:"nowShowPleaseSendData"`
	NowTextAfterMap           string `json:"nowTextAfterMap"`
	NowTextWhenFinished       string `json:"nowTextWhenFinished"`
	MobileAppName             string `json:"mobileAppName"`
	Prefix                    string `json:"prefix"`
	RoomYouAreHereEmoticon    string `json:"roomYouAreHereEmoticon"`
	Title                     string `json:"title"`
}

type Config struct {
	Meta                   Meta
	BeginningSchedule      time.Time
	Rooms                  []string
	Shedules               map[string][]string
	Motd                   string
	NbDaysForInput         int
	Buttons                []string
	TelegramToken          string
	Input                  bool
	LogFile                string
	CommandsHistoryLogFile string
	Admins                 []int
	Modos                  []int
	NowSkipClosed          bool
	Port                   int
	OldLineupMessage       string
}

func New(fileName string, isConfigCheck bool) (*Config, error) {

	errorString := ""

	c := Config{
		Shedules: make(map[string][]string),
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	c.TelegramToken = v.GetString("telegramToken")
	if isConfigCheck && v.IsSet("telegramToken") {
		errorString += "ConfigCheck: telegramToken not allowed\n"
	}
	c.Buttons = v.GetStringSlice("buttons")
	c.Rooms = v.GetStringSlice("rooms")

	if len(c.Rooms) == 0 {
		errorString += "missing rooms\n"
	}
	c.Admins = v.GetIntSlice("admins")
	if isConfigCheck && v.IsSet("admins") {
		errorString += "ConfigCheck: admins not allowed\n"
	}
	c.Modos = v.GetIntSlice("modos")
	if isConfigCheck && v.IsSet("modos") {
		errorString += "ConfigCheck: modos not allowed\n"
	}
	c.Port = v.GetInt("port")
	if isConfigCheck && v.IsSet("port") {
		errorString += "ConfigCheck: port not allowed\n"
	}
	c.Meta.NowShowShallowBunnyAd = v.GetBool("meta.nowShowShallowBunnyAd")
	c.Meta.NowShowDataSourceAd = v.GetBool("meta.nowShowDataSourceAd")
	c.Meta.NowShowSisyDuckAd = v.GetBool("meta.nowShowSisyDuckAd")
	c.Meta.NowShowPleaseSendData = v.GetBool("meta.nowShowPleaseSendData")
	c.Meta.NowTextAfterMap = v.GetString("meta.nowTextAfterMap")
	c.Meta.NowTextWhenFinished = v.GetString("meta.nowTextWhenFinished")
	c.Meta.NowBotUrl = v.GetString("meta.nowBotUrl")
	c.Meta.AboutBigIcon = v.GetString("meta.aboutBigIcon")
	c.Meta.AboutShowShallowBunnyIcon = v.GetBool("meta.aboutShowShallowBunnyIcon")
	c.Meta.AboutShowSisyDuckIcon = v.GetBool("meta.aboutShowSisyDuckIcon")
	c.Meta.NowMapImage = v.GetString("meta.nowMapImage")
	c.Meta.RoomYouAreHereEmoticon = v.GetString("meta.roomYouAreHereEmoticon")
	if c.Meta.RoomYouAreHereEmoticon == "" {
		errorString += "missing meta.roomYouAreHereEmoticon"
	}
	c.Meta.MobileAppName = v.GetString("meta.mobileAppName\n")
	c.Meta.Prefix = v.GetString("meta.prefix")
	if c.Meta.Prefix == "" {
		errorString += "missing meta.prefix\n"
	}
	c.Meta.Title = v.GetString("meta.title")
	if c.Meta.Title == "" {
		errorString += "missing meta.title\n"
	}
	for _, room := range c.Rooms {
		c.Shedules[room] = v.GetStringSlice(room)
		if !v.IsSet(room) {
			log.Warn().Msg(fmt.Sprintf("missing sets for %v", room))
		}
	}

	c.Motd = v.GetString("motd")

	c.NbDaysForInput = v.GetInt("nbDaysForInput")
	if c.NbDaysForInput == 0 {
		errorString += "Missing nbDaysForInput\n"
	}

	beg := v.GetString("beg")
	if beg == "" {
		errorString += "Missing beg\n"
	}

	c.LogFile = v.GetString("logFile")
	if isConfigCheck {
		if v.IsSet("logFile") {
			errorString += "ConfigCheck: logFile not allowed\n"
		}
	} else if c.LogFile == "" {
		errorString += "missing: logFile\n"
	}

	c.CommandsHistoryLogFile = v.GetString("commandsHistoryLogFile")
	if isConfigCheck {
		if v.IsSet("commandsHistoryLogFile") {
			errorString += "ConfigCheck: commandsHistoryLogFile not allowed\n"
		}
	}

	c.OldLineupMessage = v.GetString("oldLineupMessage")
	c.NowSkipClosed = v.GetBool("nowSkipClosed")

	c.Input = v.GetBool("input")

	c.BeginningSchedule, err = dateparse.ParseLocal(beg)
	if err != nil {
		return nil, errors.New(errorString + err.Error())
	}

	if errorString != "" {
		return nil, errors.New(errorString)
	}
	return &c, nil
}
