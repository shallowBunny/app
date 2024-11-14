package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/rs/zerolog/log"
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
	Rooms                     []string  `json:"rooms" yaml:"-"`
	Title                     string    `json:"title" yaml:"title"`
	BeginningSchedule         time.Time `json:"beginningSchedule" yaml:"-"`
	TimeZone                  string    `json:"timeZone" yaml:"timeZone"`
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
	MapImageDirectory                  string   `yaml:"secrets.mapImageDirectory,omitempty"`
	NbDaysForInput                     int      `yaml:"nbDaysForInput"`
	Buttons                            []string `yaml:"buttons"`
	ReadSetsFromRedisOnRestart         bool     `yaml:"readSetsFromRedisOnRestart"`
	LogFile                            string   `yaml:"secrets.logFile,omitempty"`
	CommandsHistoryLogFile             string   `yaml:"secrets.commandsHistoryLogFile,omitempty"`
	NowSkipClosed                      bool     `yaml:"nowSkipClosed"`
	Port                               int      `yaml:"port"`
	BeginningSchedule                  string   `yaml:"beginningSchedule"`

	Demo   bool   `yaml:"demo"`
	Meta   Meta   `yaml:"meta"`
	Lineup Lineup `yaml:"lineup"`
}

type Set struct {
	Day      int       `yaml:"day" json:"day"`
	Duration int       `yaml:"duration" json:"duration"`
	Dj       string    `yaml:"dj" json:"dj"`
	Hour     int       `yaml:"hour" json:"hour"`
	Minute   int       `yaml:"minute" json:"minute"`
	Meta     []SetMeta `yaml:"meta" json:"meta"`
}

type SetMeta struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

type Lineup struct {
	BeginningSchedule time.Time        `yaml:"-" json:"-"`
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
		c.MapImageDirectory = v.GetString("secrets.mapImageDirectory")
		c.CommandsHistoryLogFile = v.GetString("secrets.commandsHistoryLogFile")
		c.LogFile = v.GetString("secrets.logFile")
		c.Demo = v.GetBool("secrets.demo")
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

	c.Meta.TimeZone = v.GetString("meta.timeZone")
	if c.Meta.TimeZone == "" {
		c.Meta.TimeZone = "CET"
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

	loc, err := time.LoadLocation(c.Meta.TimeZone)
	if err != nil {
		errorString += err.Error()
	} else {
		time.Local = loc // -> this is setting the global timezone
	}

	beginningSchedule, err := dateparse.ParseLocal(v.GetString("beginningSchedule"))
	if err != nil {
		if !c.Demo {
			errorString += "Error on beginningSchedule\n" + err.Error()
		}
	} else {
		c.Lineup.BeginningSchedule = beginningSchedule
	}

	if c.Demo {
		now := time.Now().AddDate(0, 0, -1)
		c.Lineup.BeginningSchedule = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		log.Warn().Msg(fmt.Sprintf("demo mode, forcing BeginningSchedule to %v", c.Lineup.BeginningSchedule))
	}

	c.Meta.BeginningSchedule = c.Lineup.BeginningSchedule

	cetLocation, err := time.LoadLocation(c.Meta.TimeZone)
	if err != nil {
		errorString += err.Error()
	}
	cetTime := c.Lineup.BeginningSchedule.In(cetLocation)
	c.BeginningSchedule = cetTime.Format(time.RFC3339)

	if errorString != "" {
		return nil, errors.New(errorString)
	}

	return &c, nil
}
