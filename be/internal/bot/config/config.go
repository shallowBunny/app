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
	BotUrl                    string `json:"botUrl"`
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
	Meta                               Meta
	BeginningSchedule                  time.Time
	BotMotd                            string
	BotOldLineupMessage                string
	TelegramDeleteLeftTheGroupMessages bool
	Admins                             []int
	Modos                              []int
	TelegramToken                      string
	ServerToken                        string
	Rooms                              []string
	Shedules                           map[string][]string
	NbDaysForInput                     int
	Buttons                            []string
	Input                              bool
	LogFile                            string
	CommandsHistoryLogFile             string
	NowSkipClosed                      bool
	Port                               int
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

	if isConfigCheck && v.IsSet("secrets") {
		errorString += "ConfigCheck: secrets not allowed\n"
	} else {
		c.TelegramToken = v.GetString("secrets.telegramToken")
		c.Admins = v.GetIntSlice("secrets.admins")
		c.Modos = v.GetIntSlice("secrets.modos")
		c.Port = v.GetInt("secrets.port")
		c.ServerToken = v.GetString("secrets.serverToken")
		c.CommandsHistoryLogFile = v.GetString("secrets.commandsHistoryLogFile")
		c.LogFile = v.GetString("secrets.logFile")
	}

	c.TelegramDeleteLeftTheGroupMessages = v.GetBool("public.telegramDeleteLeftTheGroupMessages")

	c.Buttons = v.GetStringSlice("buttons")
	c.Rooms = v.GetStringSlice("rooms")

	if len(c.Rooms) == 0 {
		errorString += "missing rooms\n"
	}

	c.Meta.NowShowShallowBunnyAd = v.GetBool("meta.nowShowShallowBunnyAd")
	c.Meta.NowShowDataSourceAd = v.GetBool("meta.nowShowDataSourceAd")
	c.Meta.NowShowSisyDuckAd = v.GetBool("meta.nowShowSisyDuckAd")
	c.Meta.NowShowPleaseSendData = v.GetBool("meta.nowShowPleaseSendData")
	c.Meta.NowTextAfterMap = v.GetString("meta.nowTextAfterMap")
	c.Meta.NowTextWhenFinished = v.GetString("meta.nowTextWhenFinished")
	c.Meta.BotUrl = v.GetString("meta.botUrl")
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

	c.BotMotd = v.GetString("botMotd")

	c.NbDaysForInput = v.GetInt("nbDaysForInput")
	if c.NbDaysForInput == 0 {
		errorString += "Missing nbDaysForInput\n"
	}

	c.BotOldLineupMessage = v.GetString("botOldLineupMessage")
	c.NowSkipClosed = v.GetBool("nowSkipClosed")

	c.Input = v.GetBool("input")

	beg := v.GetString("beginningSchedule")
	if beg == "" {
		errorString += "Missing beginningSchedule\n"
	} else {
		c.BeginningSchedule, err = dateparse.ParseLocal(beg)
		if err != nil {
			errorString += err.Error()
		}
	}

	if errorString != "" {
		return nil, errors.New(errorString)
	}
	return &c, nil
}
