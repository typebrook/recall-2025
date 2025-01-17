package main

import (
	"net/url"
	"os"
)

type Config struct {
	AppEnv      string
	AppHostname string
	AppPath     string
	AppPort     string
	AppBaseURL  *url.URL
	Zones       map[string][]string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		AppEnv:      os.Getenv("APP_ENV"),
		AppHostname: os.Getenv("APP_HOSTNAME"),
		AppPath:     os.Getenv("APP_PATH"),
		AppPort:     os.Getenv("APP_PORT"),
	}

	baseURL, err := url.ParseRequestURI("https://" + cfg.AppHostname + cfg.AppPath)
	if err != nil {
		return nil, err
	}

	cfg.AppBaseURL = baseURL
	cfg.Zones = map[string][]string{
		"taipei-6": []string{
			"羅智強 - 臺北市第六選區 (大安)",
			"臺北市大安區",
		},
		"taipei-7": []string{
			"徐巧芯 - 臺北市第七選區 (信義、南松山)",
			"臺北市",
		},
	}
	return cfg, nil
}

func (r Config) HasZone(zone string) bool {
	_, exists := r.Zones[zone]
	return exists
}

func (r Config) GetZoneTopic(zone string) (string, string) {
	val, exists := r.Zones[zone]
	if !exists {
		return "", ""
	}

	return val[0], val[1]
}
