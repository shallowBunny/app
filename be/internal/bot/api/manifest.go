package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shallowBunny/app/be/internal/infrastructure/config"
)

type ManifestHandler struct {
	Config *config.Config
}

func NewManifestHandler(config *config.Config) *ManifestHandler {
	return &ManifestHandler{Config: config}
}

func (h *ManifestHandler) GetManifest(c *gin.Context) {
	manifest := Manifest{
		Name:            h.Config.Meta.MobileAppName,
		ShortName:       h.Config.Meta.MobileAppName,
		StartURL:        "/",
		Display:         "standalone",
		BackgroundColor: "#222123",
		Lang:            "en",
		Scope:           "/",
		Description:     "An app to display DJ sets",
		ThemeColor:      "#222123",
		Icons: []Icon{
			{
				Src:     h.Config.Meta.Prefix + "-192x192.png",
				Sizes:   "192x192",
				Type:    "image/png",
				Purpose: "any",
			},
			{
				Src:     h.Config.Meta.Prefix + "-180x180.png",
				Sizes:   "180x180",
				Type:    "image/png",
				Purpose: "maskable",
			},
			{
				Src:     h.Config.Meta.Prefix + "-192x192.png",
				Sizes:   "192x192",
				Type:    "image/png",
				Purpose: "maskable",
			},
		},
	}
	c.JSON(http.StatusOK, manifest)
}
