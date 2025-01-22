package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
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
	Zones              map[string]*Zone
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

	var baseURL *url.URL
	if !strings.HasPrefix(cfg.AppPath, "/") {
		cfg.AppPath = "/" + cfg.AppPath
	}

	scheme := "https"
	if cfg.AppHostname == "localhost" {
		scheme = "http"
	}

	rootPath := ""
	if cfg.AppPath == "/" {
		rootPath = cfg.AppHostname
	} else {
		rootPath = cfg.AppHostname + cfg.AppPath
	}

	baseURL, err := url.ParseRequestURI(scheme + "://" + rootPath)
	if err != nil {
		return nil, err
	}

	cfg.AppBaseURL = baseURL
	cfg.Zones = map[string]*Zone{
		"taipei-4":    &Zone{"taipei-4", "臺北市第四選區", "李彥秀", "內湖、南港", "臺北市", 4},
		"taipei-6":    &Zone{"taipei-6", "臺北市第六選區", "羅智強", "大安", "臺北市大安區", 6},
		"taipei-7":    &Zone{"taipei-7", "臺北市第七選區", "徐巧芯", "信義、南松山", "臺北市", 7},
		"newtaipei-7": &Zone{"newtaipei-7", "新北市第七選區", "葉元之", "板橋", "新北市板橋區", 7},
		"keelung-1":   &Zone{"keelung-1", "基隆市", "林沛祥", "基隆", "基隆市", 1},
	}

	return cfg, nil
}

func (r Config) HasZone(zone string) bool {
	_, exists := r.Zones[zone]
	return exists
}

func (r Config) GetZone(zone string) *Zone {
	z, exists := r.Zones[zone]
	if !exists {
		return nil
	}

	return z
}

type Zone struct {
	ZoneCode      string
	ZoneName      string
	CandidateName string
	Districts     string
	AddressPrefix string
	sort          int
}

func (r Zone) GetTopic() string {
	return r.CandidateName + " - " + r.ZoneName + " (" + r.Districts + ") "
}

const (
	AreaNameTaipei        = "臺北市"
	AreaNameNewTaipei     = "新北市"
	AreaNameKeelung       = "基隆市"
	AreaNameTaoyuan       = "桃園市"
	AreaNameHsinchuCity   = "新竹市"
	AreaNameHsinchuCounty = "新竹縣"
	AreaNameMiaoli        = "苗栗縣"
	AreaNameTaichung      = "臺中市"
	AreaNameChanghua      = "彰化縣"
	AreaNameNantou        = "南投縣"
	AreaNameHualien       = "花蓮縣"
	AreaNameTaitung       = "臺東縣"
	AreaNameKinmen        = "金門縣"
	AreaNameLienchiang    = "連江縣"
)

func (r Config) ToRecallListViewData() map[string][]*Zone {
	list := map[string][]*Zone{}
	for code, z := range r.Zones {
		pieces := strings.Split(code, "-")
		switch pieces[0] {
		case "taipei":
			if _, exists := list[AreaNameTaipei]; !exists {
				list[AreaNameTaipei] = []*Zone{}
			}
			list[AreaNameTaipei] = append(list[AreaNameTaipei], z)

		case "newtaipei":
			if _, exists := list[AreaNameNewTaipei]; !exists {
				list[AreaNameNewTaipei] = []*Zone{}
			}
			list[AreaNameNewTaipei] = append(list[AreaNameNewTaipei], z)

		case "keelung":
			if _, exists := list[AreaNameKeelung]; !exists {
				list[AreaNameKeelung] = []*Zone{}
			}
			list[AreaNameKeelung] = append(list[AreaNameKeelung], z)

		case "taoyuan":
			if _, exists := list[AreaNameTaoyuan]; !exists {
				list[AreaNameTaoyuan] = []*Zone{}
			}
			list[AreaNameTaoyuan] = append(list[AreaNameTaoyuan], z)

		case "hsinchucity":
			if _, exists := list[AreaNameHsinchuCity]; !exists {
				list[AreaNameHsinchuCity] = []*Zone{}
			}
			list[AreaNameHsinchuCity] = append(list[AreaNameHsinchuCity], z)

		case "hsinchucounty":
			if _, exists := list[AreaNameHsinchuCounty]; !exists {
				list[AreaNameHsinchuCounty] = []*Zone{}
			}
			list[AreaNameHsinchuCounty] = append(list[AreaNameHsinchuCounty], z)

		case "miaoli":
			if _, exists := list[AreaNameMiaoli]; !exists {
				list[AreaNameMiaoli] = []*Zone{}
			}
			list[AreaNameMiaoli] = append(list[AreaNameMiaoli], z)

		case "taichung":
			if _, exists := list[AreaNameTaichung]; !exists {
				list[AreaNameTaichung] = []*Zone{}
			}
			list[AreaNameTaichung] = append(list[AreaNameTaichung], z)

		case "changhua":
			if _, exists := list[AreaNameChanghua]; !exists {
				list[AreaNameChanghua] = []*Zone{}
			}
			list[AreaNameChanghua] = append(list[AreaNameChanghua], z)

		case "nantou":
			if _, exists := list[AreaNameNantou]; !exists {
				list[AreaNameNantou] = []*Zone{}
			}
			list[AreaNameNantou] = append(list[AreaNameNantou], z)

		case "hualien":
			if _, exists := list[AreaNameHualien]; !exists {
				list[AreaNameHualien] = []*Zone{}
			}
			list[AreaNameHualien] = append(list[AreaNameHualien], z)

		case "taitung":
			if _, exists := list[AreaNameTaitung]; !exists {
				list[AreaNameTaitung] = []*Zone{}
			}
			list[AreaNameTaitung] = append(list[AreaNameTaitung], z)

		case "kinmen":
			if _, exists := list[AreaNameKinmen]; !exists {
				list[AreaNameKinmen] = []*Zone{}
			}
			list[AreaNameKinmen] = append(list[AreaNameKinmen], z)

		case "lienchiang":
			if _, exists := list[AreaNameLienchiang]; !exists {
				list[AreaNameLienchiang] = []*Zone{}
			}
			list[AreaNameLienchiang] = append(list[AreaNameLienchiang], z)
		}
	}

	for _, zones := range list {
		sort.Slice(zones, func(i, j int) bool {
			return zones[i].sort < zones[j].sort
		})
	}

	return list
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
