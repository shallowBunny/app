package main

import (
	"context"
	"flag"
	"net/http"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/shallowBunny/app/be/internal/bot"
	"github.com/shallowBunny/app/be/internal/bot/api"
	"github.com/shallowBunny/app/be/internal/infrastructure/config"
	"github.com/shallowBunny/app/be/internal/infrastructure/logging"
	"github.com/shallowBunny/app/be/internal/infrastructure/middleware"
	dao "github.com/shallowBunny/app/be/internal/infrastructure/repository"
	DaoDb "github.com/shallowBunny/app/be/internal/infrastructure/repository/daoDb"
	DaoMem "github.com/shallowBunny/app/be/internal/infrastructure/repository/daoMem"
	"github.com/shallowBunny/app/be/internal/utils"

	"github.com/shallowBunny/app/be/internal/bot/telegram"
)

func createServer(b *bot.Bot) *http.Server {

	r := gin.New()

	r.Use(middleware.ZerologMiddleware()) // Use zerolog middleware instead of default logger

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"}, // Update with your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Forwarded-For"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	botHandler := api.NewBotHandler(b)

	r.GET("/api", botHandler.GetLineUp)
	r.POST("/api", botHandler.TokenAuthMiddleware(), botHandler.Restart)
	r.PUT("/api", botHandler.TokenAuthMiddleware(), botHandler.UpdateLineUp)

	manifestHandler := api.NewManifestHandler(b.GetConfig())
	r.GET("/manifest", manifestHandler.GetManifest)
	r.GET("/manifest.webmanifest", manifestHandler.GetManifest)

	// Create an HTTP server using the Gin router
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", b.GetConfig().Port),
		Handler: r, // Gin engine as the HTTP handler
	}
	return server
}

func runRestartScript(runRestartScriptArg string) (string, error) {
	cmd := exec.Command("./restart.sh", runRestartScriptArg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("runRestartScript failed: %v\n%s", err, string(output))
	}
	return string(output), nil
}

func main() {

	configFileArg := flag.String("config", "", "use given config file")
	checkConfig := flag.Bool("check", false, "check config")
	restartScriptArg := flag.String("script", "", "restart script")

	flag.Parse()

	if *configFileArg == "" {
		fmt.Println("Error: --config is a mandatory flag")
		flag.Usage() // Display the usage information
		os.Exit(1)   // Exit the program with a non-zero status
	}

	var restartScriptOutput string
	var restartScriptError error
	if *restartScriptArg != "" {
		restartScriptOutput, restartScriptError = runRestartScript(*restartScriptArg)
	}

	config, err := config.New(*configFileArg, *checkConfig)
	if err != nil {
		panic(err)
	}

	// Initialize logging and get the file handles
	logFile := logging.InitLogging(config.LogFile)
	if logFile != nil {
		defer logFile.Close() // Ensure files are closed when main exits
	}

	log.Info().Msg("using config " + config.LogFile)

	telegramToken := config.TelegramToken

	if utils.IsLocalhostTesting() {
		log.Debug().Msg("isLocalhostTesting is true: using env SHALLOWBUNNY_TELEGRAM_API_TOKEN and port 8082")
		envToken := os.Getenv("SHALLOWBUNNY_TELEGRAM_API_TOKEN")
		if envToken != "" {
			telegramToken = envToken
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
		dao = DaoDb.New(telegramToken, redisclient)
	}

	bot := bot.New(dao, config)

	if *checkConfig {
		log.Info().Msg("Checked config: OK")
		log.Info().Msg(bot.PrintLineupForCheckConfig())

	} else {

		gin.SetMode(gin.ReleaseMode)

		var server *http.Server
		quitTelegram := make(chan struct{})

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

		if telegramToken != "" {
			log.Info().Msg("starting telegram bot")
			telegram := telegram.New(telegramToken, bot)
			go func() {

				restartMsg := fmt.Sprintf("✅ Restarted using %v key: %v", *configFileArg, dao.GetKey())

				if config.TelegramDeleteLeftTheGroupMessages {
					restartMsg += "\n✅ TelegramDeleteLeftTheGroupMessages Activated"
				}

				if restartScriptError != nil {
					restartMsg += "\n⚠️ " + restartScriptError.Error() + "\n"
				} else {
					restartMsg += "\n✅ "
				}
				if restartScriptOutput != "" {
					restartMsg += restartScriptOutput
				}
				restartMsg += bot.RootLineUp.GetSetsAndDurations()
				bot.SendAdminsMessage(restartMsg)
				bot.Log(0, restartMsg, "")
			}()
			go telegram.Listen(quitTelegram)
		} else {
			log.Info().Msg("no telegram token. skipping telegram")
		}

		if telegramToken != "" || server != nil {
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

			if telegramToken != "" {
				close(quitTelegram)
			}
		}
		log.Info().Msg("Server exiting")
	}
}
