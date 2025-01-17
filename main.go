package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	c := NewController(cfg)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")

	r.NoRoute(c.NotFound())
	r.GET("/health/v1/ping", c.Ping())
	r.GET("/", c.Home())
	r.GET("/stage-1/candidates", c.Stage1Candidates())
	r.GET("/:stage/:zone", c.FillForm())
	r.GET("/:stage/:zone/preview", c.PreviewLocalForm())
	r.GET("/preview/:stage/:zone", c.PreviewOriginalLocalForm())
	r.GET("/assets/:type/:file", c.GetAsset())

	r.Run(":" + cfg.AppPort)
}
