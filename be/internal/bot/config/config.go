package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Meta struct {
	AboutBigIcon              string    `json:"aboutBigIcon" yaml:"aboutBigIcon"`
	AboutShowShallowBunnyIcon bool      `json:"aboutShowShallowBunnyIcon" yaml:"aboutShowShallowBunnyIcon"`
	AboutShowSisyDuckIcon     bool      `json:"aboutShowSisyDuckIcon" yaml:"aboutShowSisyDuckIcon"`
	BotUrl                    string    `json:"botUrl" yaml:"botUrl"`
	NowMapImage               string    `json:"nowMapImage" yaml:"nowMapImage"`
	NowShowDataSourceAd       bool      `json:"nowShowDataSourceAd" yaml:"nowShowDataSourceAd"`
	NowShowShallowBunnyAd     bool      `json:"nowShowShallowBunnyAd" yaml:"nowShowShallowBunnyAd"`
	NowShowSisyDuckAd         bool      `json:"nowShowSisyDuckAd" yaml:"nowShowSisyDuckAd"`
	NowSubmitPR               string    `json:"nowSubmitPR" yaml:"nowSubmitPR"`
	NowTextAfterMap           string    `json:"nowTextAfterMap" yaml:"nowTextAfterMap"`
	NowTextWhenFinished       string    `json:"nowTextWhenFinished" yaml:"nowTextWhenFinished"`
	MobileAppName             string    `json:"mobileAppName" yaml:"mobileAppName"`
	Prefix                    string    `json:"prefix" yaml:"prefix"`
	RoomYouAreHereEmoticon    string    `json:"roomYouAreHereEmoticon" yaml:"roomYouAreHereEmoticon"`
	Rooms                     []string  `json:"rooms" yaml:"rooms"`
	Title                     string    `json:"title" yaml:"title"`
	BeginningSchedule         time.Time `json:"beginningSchedule" yaml:"-"`
}

type Config struct {
	BotAllowInput                      bool     `yaml:"botAllowInput"`
	BotMotd                            string   `yaml:"botMotd"`
	BotOldLineupMessage                string   `yaml:"botOldLineupMessage"`
	BotNoDataAvailableYet              string   `yaml:"botNoDataAvailableYet"`
	TelegramDeleteLeftTheGroupMessages bool     `yaml:"telegramDeleteLeftTheGroupMessages"`
	Admins                             []int    `yaml:"secrets.admins,omitempty"`
	Modos                              []int    `yaml:"secrets.modos,omitempty"`
	TelegramToken                      string   `yaml:"secrets.telegramToken,omitempty"`
	ServerToken                        string   `yaml:"secrets.serverToken,omitempty"`
	NbDaysForInput                     int      `yaml:"nbDaysForInput"`
	Buttons                            []string `yaml:"buttons"`
	ReadSetsFromRedisOnRestart         bool     `yaml:"readSetsFromRedisOnRestart"`
	LogFile                            string   `yaml:"secrets.logFile,omitempty"`
	CommandsHistoryLogFile             string   `yaml:"secrets.commandsHistoryLogFile,omitempty"`
	NowSkipClosed                      bool     `yaml:"nowSkipClosed"`
	Port                               int      `yaml:"port"`
	BeginningScheduleString            string   `yaml:"beginningSchedule"`
	Meta                               Meta     `yaml:"meta"`
	Lineup                             Lineup   `yaml:"lineup"`
}

type Set struct {
	Day      int    `yaml:"day" json:"day"`
	Duration int    `yaml:"duration" json:"duration"`
	Dj       string `yaml:"dj" json:"dj"`
	Hour     int    `yaml:"hour" json:"hour"`
	Minute   int    `yaml:"minute" json:"minute"`
}

type Lineup struct {
	BeginningSchedule time.Time        `yaml:"-" json:"beginningSchedule"`
	Rooms             []string         `yaml:"rooms" json:"rooms"`
	Sets              map[string][]Set `yaml:"sets" json:"sets"`
}

func (config Config) WriteConfigToFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	return encoder.Encode(config)
}

func WriteConfigToFile2(config Config, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // For pretty printing
	return encoder.Encode(config)
}

func New(fileName string, isConfigCheck bool) (*Config, error) {

	errorString := ""

	c := Config{}

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

	c.TelegramDeleteLeftTheGroupMessages = v.GetBool("telegramDeleteLeftTheGroupMessages")

	c.Buttons = v.GetStringSlice("buttons")

	c.Meta.NowShowShallowBunnyAd = v.GetBool("meta.nowShowShallowBunnyAd")
	c.Meta.NowShowDataSourceAd = v.GetBool("meta.nowShowDataSourceAd")
	c.Meta.NowShowSisyDuckAd = v.GetBool("meta.nowShowSisyDuckAd")
	c.Meta.NowSubmitPR = v.GetString("meta.nowSubmitPR")
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
	c.Meta.MobileAppName = v.GetString("meta.mobileAppName")
	c.Meta.Prefix = v.GetString("meta.prefix")
	if c.Meta.Prefix == "" {
		errorString += "missing meta.prefix\n"
	}
	c.Meta.Title = v.GetString("meta.title")
	if c.Meta.Title == "" {
		errorString += "missing meta.title\n"
	}

	v2 := v.Sub("lineup")
	if v2 == nil {
		errorString += "Error: 'lineup' key not found in configuration\n"
	} else {
		c.Lineup.Sets = make(map[string][]Set)
		lineup := Lineup{
			Sets: make(map[string][]Set),
		}
		if err := v2.Unmarshal(&lineup); err != nil {
			errorString += fmt.Sprintf("Error unmarshalling lineup: %v\n", err)
		}

		for _, room := range lineup.Rooms {
			sets, ok := lineup.Sets[strings.ToLower(room)]
			if ok {
				c.Lineup.Sets[room] = sets
			}
		}
		if len(lineup.Rooms) == 0 {
			errorString += "missing rooms\n"
		}
		c.Lineup.Rooms = lineup.Rooms
	}

	c.BotMotd = v.GetString("botMotd")

	c.NbDaysForInput = v.GetInt("nbDaysForInput")
	if c.NbDaysForInput == 0 {
		errorString += "Missing nbDaysForInput\n"
	}

	c.BotOldLineupMessage = v.GetString("botOldLineupMessage")
	c.NowSkipClosed = v.GetBool("nowSkipClosed")

	c.ReadSetsFromRedisOnRestart = v.GetBool("readSetsFromRedisOnRestart")
	c.BotAllowInput = v.GetBool("botAllowInput")

	c.BotNoDataAvailableYet = v.GetString("botNoDataAvailableYet")
	if c.BotNoDataAvailableYet == "" {
		c.BotNoDataAvailableYet = "⚠️ No data available yet ⚠️"
	}

	c.BeginningScheduleString = v.GetString("beginningSchedule")
	if c.BeginningScheduleString == "" {
		errorString += "Missing beginningSchedule\n"
	} else {
		c.Lineup.BeginningSchedule, err = dateparse.ParseLocal(c.BeginningScheduleString)
		if err != nil {
			errorString += err.Error()
		}
		c.Meta.BeginningSchedule = c.Lineup.BeginningSchedule
	}

	if errorString != "" {
		return nil, errors.New(errorString)
	}

	return &c, nil
}
