package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	AppEnv             string
	AppHostname        string
	AppPath            string
	AppPort            string
	AppTrustedProxies  []string
	AppBaseURL         *url.URL
	TurnstileSiteKey   string
	TurnstileSecretKey string
	Zones              map[string][]string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		AppEnv:             os.Getenv("APP_ENV"),
		AppHostname:        os.Getenv("APP_HOSTNAME"),
		AppPath:            os.Getenv("APP_PATH"),
		AppPort:            os.Getenv("APP_PORT"),
		AppTrustedProxies:  strings.Split(strings.ReplaceAll(os.Getenv("APP_TRUSTED_PROXIES"), " ", ""), ","),
		TurnstileSiteKey:   os.Getenv("TURNSTILE_SITE_KEY"),
		TurnstileSecretKey: os.Getenv("TURNSTILE_SECRET_KEY"),
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

func (r Config) VerifyTurnstileToken(token string) (bool, error) {
	verifyURL := "https://challenges.cloudflare.com/turnstile/v0/siteverify"

	data := map[string]string{
		"secret":   r.TurnstileSecretKey,
		"response": token,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(verifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	result := TurnstileSiteverifyResponse{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, fmt.Errorf("verification failed: %v", result.ErrorCodes)
	}

	return true, nil
}

type TurnstileSiteverifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
	Messages   []string `json:"messages"`
}
