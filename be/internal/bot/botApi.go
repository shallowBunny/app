package bot

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/shallowBunny/app/be/internal/bot/config"
	"github.com/shallowBunny/app/be/internal/bot/lineUp"

	"github.com/rs/zerolog/log"
)

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

func getClientIPByRequest(req *http.Request) (ip string) {
	forwarded := req.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) >= 1 {
			return strings.TrimSpace(ips[0])
		}
	}
	return "?"
}

func (b *Bot) GetManifest(c *gin.Context) {
	manifest := Manifest{
		Name:            b.config.Meta.MobileAppName,
		ShortName:       b.config.Meta.MobileAppName,
		StartURL:        "/",
		Display:         "standalone",
		BackgroundColor: "#222123",
		Lang:            "en",
		Scope:           "/",
		Description:     "An app to display DJ sets",
		ThemeColor:      "#222123",
		Icons: []Icon{
			{
				Src:     b.config.Meta.Prefix + "-192x192.png",
				Sizes:   "192x192",
				Type:    "image/png",
				Purpose: "any",
			},
			{
				Src:     b.config.Meta.Prefix + "-180x180.png",
				Sizes:   "180x180",
				Type:    "image/png",
				Purpose: "maskable",
			},
			{
				Src:     b.config.Meta.Prefix + "-192x192.png",
				Sizes:   "192x192",
				Type:    "image/png",
				Purpose: "maskable",
			},
		},
	}
	c.JSON(http.StatusOK, manifest)
}

func (b Bot) GetLineUp(c *gin.Context) {
	var response Response
	ip := getClientIPByRequest(c.Request)
	response.Sets = b.RootLineUp.Sets
	response.Meta = b.config.Meta
	response.Meta.Rooms = b.RootLineUp.Rooms
	b.Log(0, c.Request.UserAgent(), ip)
	c.JSON(http.StatusOK, response)
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

func (b Bot) Restart(c *gin.Context) {

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

	b.SendAdminsMessage(restartMsg)

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

func (b Bot) TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token != "Bearer "+b.config.ServerToken || b.config.ServerToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
