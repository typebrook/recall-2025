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
	r.SetTrustedProxies(cfg.AppTrustedProxies)

	r.NoRoute(c.NotFound())
	r.GET("/health/v1/ping", c.Ping())
	r.GET("/", c.Home())
	r.GET("/:stage/:zone", c.FillForm())
	r.POST("/:stage/:zone/preview", c.VerifyTurnstile(), c.PreviewLocalForm())
	r.GET("/preview/:stage/:zone", c.PreviewOriginalLocalForm())
	r.GET("/assets/:type/:file", c.GetAsset())
	r.GET("/thank-you", c.ThankYou())

	r.Run(":" + cfg.AppPort)
}
