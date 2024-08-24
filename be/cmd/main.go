package main

import (
	"context"
	"flag"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shallowBunny/app/be/internal/bot"
	"github.com/shallowBunny/app/be/internal/bot/config"
	"github.com/shallowBunny/app/be/internal/bot/dao"
	DaoDb "github.com/shallowBunny/app/be/internal/bot/dao/daoDb"
	DaoMem "github.com/shallowBunny/app/be/internal/bot/dao/daoMem"

	"github.com/shallowBunny/app/be/internal/bot/telegram"
)

func createServer(b *bot.Bot) *http.Server {

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"}, // Update with your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Forwarded-For"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/api", b.GetLineUp)
	r.GET("/manifest", b.GetManifest)
	r.GET("/manifest.webmanifest", b.GetManifest)

	go r.Run()

	// Create an HTTP server using the Gin router
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", b.Config.Port),
		Handler: r, // Gin engine as the HTTP handler
	}
	return server
}

func isLocalhostTesting() bool {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get hostname")
	}
	return hostname == os.Getenv("SHALLOWBUNNY_LOCALHOST")
}

func initLogging(fileName string, checkConfig bool) *os.File {

	if checkConfig {
		consoleWriter := zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: true, // Disable colors
		}
		log.Logger = zerolog.New(consoleWriter).
			With().
			Timestamp().
			Logger()

		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return nil
	} else {
		if isLocalhostTesting() {
			consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}
			log.Logger = zerolog.New(consoleWriter).
				With().
				Timestamp().
				Caller(). // Adjust the skip frame count as needed
				Logger()
			log.Debug().Msg("isLocalhostTesting is true: not using logFile, logging initialized for console with colors and caller information")
			return nil
		} else {
			// Open files for different log levels
			debugFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to open debug log file")
			}
			log.Logger = zerolog.New(debugFile).
				With().
				Timestamp().
				Caller(). // Adjust the skip frame count as needed
				Logger()
			zerolog.SetGlobalLevel(zerolog.InfoLevel) // Skip debug level
			return debugFile                          // Return both file handles
		}
	}
}

func main() {

	loc, err := time.LoadLocation("CET")
	if err != nil {
		log.Error().Msg(err.Error())
		panic("ok")
	}
	// handle err
	time.Local = loc // -> this is setting the global timezone

	configFile := "configs/default.yaml"
	configFileArg := flag.String("config", "", "config file")
	checkConfig := flag.Bool("checkConfig", false, "check config")

	flag.Parse()

	if *configFileArg != "" {
		configFile = *configFileArg
	}

	config, err := config.New(configFile, *checkConfig)
	if err != nil {
		panic(err)
	}
	// Initialize logging and get the file handles
	logFile := initLogging(config.LogFile, *checkConfig)
	if logFile != nil {
		defer logFile.Close() // Ensure files are closed when main exits
	}
	if *configFileArg != "" {
		log.Info().Msg(fmt.Sprintf("using config file from arg: %v", configFile))
	}

	apiToken := config.TelegramToken

	if isLocalhostTesting() {
		log.Debug().Msg("isLocalhostTesting is true: using env SHALLOWBUNNY_TELEGRAM_API_TOKEN and port 8082")
		envToken := os.Getenv("SHALLOWBUNNY_TELEGRAM_API_TOKEN")
		if envToken != "" {
			apiToken = envToken
		}
		config.Port = 8082
	}

	var dao dao.Dao

	if *checkConfig {
		dao = DaoMem.New()
	} else {
		redisclient := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		dao = DaoDb.New(apiToken, redisclient)
	}

	bot := bot.New(dao, config)

	if *checkConfig {
		log.Info().Msg("Checked config: OK")
		log.Info().Msg(bot.PrintLinupForCheckConfig())

	} else {

		gin.SetMode(gin.ReleaseMode)

		var server *http.Server

		if config.Port != 0 {
			log.Info().Msg("starting rest api")
			server = createServer(bot)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatal().Msg(err.Error())
				}
			}()
		} else {
			log.Info().Msg("skipping rest api")
		}

		if apiToken != "" {
			log.Info().Msg("starting telegram bot")
			telegram := telegram.New(apiToken, bot)
			go func() {
				restartMsg := fmt.Sprintf("restarted using %v key: %v", configFile, dao.GetKey())
				bot.SendAdminsMessage(restartMsg)
				bot.Log(0, restartMsg, "")
			}()
			go telegram.Listen()
		} else {
			log.Info().Msg("no telegram token. skipping telegram")
		}

		if apiToken != "" || server != nil {
			// Create a channel to listen for termination signals
			quit := make(chan os.Signal, 1)

			// Relay SIGINT, SIGTERM to the quit channel
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			// Block until we receive a signal
			<-quit

			if server != nil {
				log.Info().Msg("Shutting down rest api...")
				// Create a context with a timeout for graceful shutdown
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// Attempt a graceful shutdown
				if err := server.Shutdown(ctx); err != nil {
					log.Error().Msg(fmt.Sprintf("Rest api to shutdown:%v", err))
				}
			}
		}
		log.Info().Msg("Server exiting")
	}
}
