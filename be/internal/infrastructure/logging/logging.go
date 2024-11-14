package logging

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shallowBunny/app/be/internal/utils"
)

func InitLogging(fileName string) *os.File {
	if utils.IsLocalhostTesting() {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}
		log.Logger = zerolog.New(consoleWriter).
			With().
			Timestamp().
			Caller().
			Logger()
		log.Debug().Msg("isLocalhostTesting is true: not using logFile, logging initialized for console with colors and caller information")
		return nil
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // Skip debug level
		logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// will be used for checkConfig option
			consoleWriter := zerolog.ConsoleWriter{
				Out:     os.Stderr,
				NoColor: true, // Disable colors
			}
			log.Logger = zerolog.New(consoleWriter).
				With().
				Timestamp().
				Logger()
			return nil
		}
		// Configure ConsoleWriter for the file to disable JSON format and colors
		fileWriter := zerolog.ConsoleWriter{
			Out:        logFile,
			NoColor:    true,
			TimeFormat: "2006-01-02 15:04:05", // Adjust time format as needed
		}
		log.Logger = zerolog.New(fileWriter).
			With().
			Timestamp().
			//		Caller().
			Logger()
		return logFile
	}
}
