package api

import (
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/shallowBunny/app/be/internal/bot"
	"github.com/shallowBunny/app/be/internal/bot/lineUp"
	"github.com/shallowBunny/app/be/internal/bot/lineUp/inputs"
	"github.com/shallowBunny/app/be/internal/infrastructure/config"
	"github.com/shallowBunny/app/be/internal/utils"

	"github.com/rs/zerolog/log"
)

type BotHandler struct {
	Bot *bot.Bot
}

// NewManifestHandler initializes a new ManifestHandler with the necessary config
func NewBotHandler(bot *bot.Bot) *BotHandler {
	return &BotHandler{Bot: bot}
}

type Response struct {
	Meta config.Meta  `json:"meta"`
	Sets []lineUp.Set `json:"sets"`
}

// Define the Manifest struct
type Manifest struct {
	Name            string `json:"name"`
	ShortName       string `json:"short_name"`
	StartURL        string `json:"start_url"`
	Display         string `json:"display"`
	BackgroundColor string `json:"background_color"`
	Lang            string `json:"lang"`
	Scope           string `json:"scope"`
	Description     string `json:"description"`
	ThemeColor      string `json:"theme_color"`
	Icons           []Icon `json:"icons"`
}

// Define the Icon struct
type Icon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

func (b *BotHandler) GetLineUp(c *gin.Context) {
	var response Response
	ip := utils.GetClientIPByRequest(c.Request)
	go b.Bot.StatsUsingUserIp(ip)
	response.Sets = b.Bot.RootLineUp.Sets
	response.Meta = b.Bot.GetConfig().Meta
	response.Meta.Rooms = b.Bot.GetConfig().Lineup.Rooms
	b.Bot.Log(0, c.Request.UserAgent(), ip)
	c.JSON(http.StatusOK, response)
}

func convertLineupToInputCommandResultSets(lineup config.Lineup) []inputs.InputCommandResultSet {
	var results []inputs.InputCommandResultSet

	for room, sets := range lineup.Sets {
		for _, set := range sets {
			result := inputs.InputCommandResultSet{
				Room:     room,
				Dj:       set.Dj,
				Day:      set.Day,
				Hour:     set.Hour,
				Minute:   set.Minute,
				Duration: set.Duration,
			}
			results = append(results, result)
		}
	}
	return results
}

func (b *BotHandler) TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token != "Bearer "+b.Bot.GetConfig().ServerToken || b.Bot.GetConfig().ServerToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (b *BotHandler) UpdateLineUp(c *gin.Context) {
	var lineup config.Lineup

	// Bind the JSON body to the Lineup struct
	if err := c.ShouldBindJSON(&lineup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process the lineup data here (e.g., update your configuration, save to a database, etc.)
	log.Printf("Received Lineup: %+v\n", lineup)

	mr := bot.NewMergeRequest(lineup.BeginningSchedule, convertLineupToInputCommandResultSets(lineup), 0, "api", utils.GetClientIPByRequest(c.Request))

	err := b.Bot.ChecForDuplicateMergeRequest(mr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = b.Bot.CheckMergeRequest(mr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b.Bot.CreateMergeRequest(*mr)

	// Respond to the client
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Created MR %v with changes", mr.ID),
	})
}

func (b *BotHandler) Message(c *gin.Context) {
	var json struct {
		AdminMsg string `json:"adminMsg"` // Expecting "adminMsg" in the JSON body
	}

	// Parse the incoming JSON body
	if err := c.ShouldBindJSON(&json); err != nil {
		// If parsing fails, respond with an error
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the adminMsg from the parsed JSON
	adminMsg := json.AdminMsg

	// Send the message to admins using the bot
	b.Bot.SendAdminsMessage(adminMsg)
	b.Bot.Log(666, "", adminMsg)

	// Respond back to the client
	c.JSON(http.StatusOK, gin.H{"status": "Message sent to admins"})
}

type RestartRequest struct {
	MergedBy    string `json:"merged_by"`
	CreatedBy   string `json:"created_by"`
	PrUrl       string `json:"pr_url"`
	PushType    string `json:"push_type"`
	PushedBy    string `json:"pushed_by"`
	PusherName  string `json:"pusher_name"`
	PusherEmail string `json:"pusher_email"`
}

func (b *BotHandler) Restart(c *gin.Context) {

	// Parse the JSON request body
	var req RestartRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	var restartMsg string
	if req.PushType == "merge" {
		restartMsg = fmt.Sprintf("Restart: Merged by: %s, Created by: %s, PR URL: %s", req.MergedBy, req.CreatedBy, req.PrUrl)
	} else if req.PushType == "force push" {
		restartMsg = fmt.Sprintf("Restart: Force-pushed by: %s, Pusher: %s (%s), PR URL: %s", req.PushedBy, req.PusherName, req.PusherEmail, req.PrUrl)
	} else {
		restartMsg = fmt.Sprintf("Restart: Unknown push type. Pushed by: %s, Pusher: %s (%s), PR URL: %s", req.PushedBy, req.PusherName, req.PusherEmail, req.PrUrl)
	}

	b.Bot.SendAdminsMessage(restartMsg)

	// Send the response back to the client
	c.JSON(http.StatusOK, gin.H{
		"message": "Server is restarting...",
	})

	// Trigger the graceful shutdown by sending a SIGTERM to the current process
	pid := os.Getpid()
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		log.Printf("Failed to send SIGTERM: %v", err)
	}
}
