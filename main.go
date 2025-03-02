package main

import (
	"html/template"
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	c := NewController(cfg)

	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"urlencode": url.QueryEscape,
	})
	r.LoadHTMLGlob("templates/*.html")
	r.SetTrustedProxies(cfg.AppTrustedProxies)

	r.NoRoute(c.NotFound())
	r.GET("/health/v1/ping", c.Ping())
	r.GET("/", c.Home())
	r.GET("/authorization-letter", c.AuthorizationLetter())
	r.GET("/apis/constituencies", c.SearchRecallConstituency())
	r.GET("/legislators/:name", c.Participate())
	r.POST("/legislators/:name/preview", c.VerifyTurnstile(), c.PreviewLocalForm())
	r.GET("/legislators/:name/thank-you", c.ThankYou())
	r.GET("/preview/stages/:stage/:name", c.PreviewOriginalLocalForm())
	r.GET("/robots.txt", c.RobotsTxt())
	r.GET("/sitemap.xml", c.Sitemap())
	r.GET("/assets/:type/:file", c.GetAsset())

	r.Run(":" + cfg.AppPort)
}
