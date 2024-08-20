package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	Meta                      Meta
	BeginningSchedule         time.Time
	Rooms                     []string
	Shedules                  map[string][]string
	Motd                      string
	Days                      int
	Buttons                   []string
	TelegramToken             string
	Input                     bool
	LogFile                   string
	CommandsHistoryLogFile    string
	Admins                    []int
	Modos                     []int
	NowSkipClosed             bool
	Port                      int
	PrintThisIsLastWeekLineup bool
}

// ExtractFilename extracts the filename from a given path without the leading path and extension
func extractFilename(fullPath string) string {
	// Get the base name of the file
	base := filepath.Base(fullPath)

	// Remove the extension
	ext := filepath.Ext(base)
	filename := strings.TrimSuffix(base, ext)

	return filename
}

func New(fileName string) *Config {

	c := Config{
		Shedules: make(map[string][]string),
	}

	data, err := os.ReadFile(fileName)

	if err != nil {
		panic(err)
	}

	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(data))

	if err != nil {
		panic(err)
	}

	c.TelegramToken = v.GetString("telegramToken")

	c.Buttons = v.GetStringSlice("buttons")

	c.Rooms = v.GetStringSlice("rooms")

	c.Admins = v.GetIntSlice("admins")
	c.Modos = v.GetIntSlice("modos")
	c.Port = v.GetInt("port")

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
	c.Meta.MobileAppName = v.GetString("meta.mobileAppName")
	c.Meta.Prefix = v.GetString("meta.prefix")
	c.Meta.Title = v.GetString("meta.title")
	if c.Meta.RoomYouAreHereEmoticon == "" {
		panic("no RoomYouAreHereEmoticon")
	}
	if c.Meta.Prefix == "" {
		c.Meta.Prefix = extractFilename(fileName)
	}
	if c.Meta.Title == "" {
		c.Meta.Title = c.Meta.Prefix
	}

	if len(c.Rooms) == 0 {
		log.Error().Msg(fmt.Sprintf("c.rooms: %v", c.Rooms))
		panic("no rooms")
	}

	for _, room := range c.Rooms {
		c.Shedules[room] = v.GetStringSlice(room)
		if c.Shedules[room] == nil {
			log.Error().Msg(fmt.Sprintf("No shedule for: %v", room))
		}
	}

	c.Motd = v.GetString("motd")
	if c.Motd == "" {
		panic("no Motd")
	}

	c.Days = v.GetInt("days")
	if c.Days == 0 {
		panic("no Days")
	}

	beg := v.GetString("beg")
	if beg == "" {
		panic("no beg")
	}

	c.LogFile = v.GetString("logFile")
	if c.LogFile == "" {
		panic("no LogFile")
	}
	c.CommandsHistoryLogFile = v.GetString("commandsHistoryLogFile")

	c.PrintThisIsLastWeekLineup = v.GetBool("printThisIsLastWeekLineup")

	c.NowSkipClosed = v.GetBool("nowSkipClosed")

	c.Input = v.GetBool("input")

	c.BeginningSchedule, err = dateparse.ParseLocal(beg)
	if err != nil {
		log.Error().Msg(err.Error())
		panic(err.Error())
	}

	return &c
}
