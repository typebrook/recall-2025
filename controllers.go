package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	MayorName = "高虹安"
	MayorCity = "新竹市"
)

type Controller struct {
	*Config
	Templates *template.Template
}

func NewController(cfg *Config, tmpl *template.Template) *Controller {
	return &Controller{
		Config:    cfg,
		Templates: tmpl,
	}
}

func (ctrl *Controller) CalcDaysLeft() error {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return err
	}

	now := time.Now().In(loc)
	ctrl.Config.CalcDaysLeft(now)
	return nil
}

func (ctrl *Controller) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		ctrl.renderTemplate(w, "home.html", map[string]interface{}{
			"BaseURL":        ctrl.AppBaseURL.String(),
			"Municipalities": ctrl.Municipalities,
			"Areas":          ctrl.Areas,
		})
	} else {
		ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
	}
}

func (ctrl *Controller) AuthorizationLetter(w http.ResponseWriter, r *http.Request) {
	ctrl.renderTemplate(w, "authorization-letter.html", map[string]interface{}{
		"BaseURL": ctrl.AppBaseURL.String(),
	})
}

func (ctrl *Controller) SearchRecallConstituency(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var qp RequestQuerySearchRecallConstituency

	m := r.FormValue("municipality")
	mid, err := strconv.ParseUint(m, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "municipality error"})
		return
	}
	qp.MunicipalityId = mid

	if d := r.FormValue("district"); d != "" {
		val, err := strconv.ParseUint(d, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "district error"})
			return
		}
		qp.DistrictId = &val
	}

	if wd := r.FormValue("ward"); wd != "" {
		val, err := strconv.ParseUint(wd, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ward error"})
			return
		}
		qp.WardId = &val
	}

	exists, divisions, legislators := ctrl.HasRecallLegislators(qp.MunicipalityId, qp.DistrictId, qp.WardId)
	if !exists {
		writeJSON(w, http.StatusNotFound, RespSearchRecallConstituency{
			Message: http.StatusText(http.StatusNotFound),
		})
		return
	}

	writeJSON(w, http.StatusOK, RespSearchRecallConstituency{
		Message: http.StatusText(http.StatusOK),
		Result: &ResultSearchRecallConstituency{
			Divisions:   divisions,
			Legislators: legislators,
		},
	})
}

func (ctrl *Controller) Participate(w http.ResponseWriter, r *http.Request, name string) {
	l := ctrl.GetRecallLegislator(name)
	if l == nil || l.RecallStatus != RecallStatusOngoing {
		http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		address = l.MunicipalityName
	}

	switch l.RecallStage {
	case 1, 2:
		ctrl.renderTemplate(w, "fill-form.html", map[string]interface{}{
			"BaseURL":          ctrl.AppBaseURL.String(),
			"PreviewURL":       l.ParticipateURL.JoinPath("preview").String(),
			"Address":          address,
			"TurnstileSiteKey": ctrl.TurnstileSiteKey,
			"Legislator":       l,
		})
	case 3, 4:
		ctrl.renderTemplate(w, "vote-reminder.html", map[string]interface{}{
			"BaseURL":    ctrl.AppBaseURL.String(),
			"Legislator": l,
		})
	default:
		http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
	}
}

func (ctrl *Controller) PreviewLocalForm(w http.ResponseWriter, r *http.Request, name string) {
	l := ctrl.GetRecallLegislator(name)
	if l == nil || l.RecallStatus != RecallStatusOngoing || !l.IsPetitioning() {
		ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusConflict, "候選人不處於連署階段", ctrl.AppBaseURL, ctrl.AppBaseURL))
		return
	}

	qp, err := getRequestForm(r)
	if err != nil {
		ctrl.renderTemplate(w, "4xx.html", ViewHttp4xxError{
			HttpStatusCode: http.StatusBadRequest,
			ErrorMessage:   "輸入有誤",
			ReturnURL:      ctrl.AppBaseURL.String(),
		})
		return
	}

	up := RequestUriStageLegislator{Name: name, Stage: l.RecallStage}
	data, err := qp.ToPreviewData(ctrl.Config, &up, l)
	if err != nil {
		ctrl.renderTemplate(w, "4xx.html", ViewHttp4xxError{
			HttpStatusCode: http.StatusBadRequest,
			ErrorMessage:   err.Error(),
			ReturnURL:      ctrl.AppBaseURL.String(),
		})
		return
	}

	tmpfile := l.GetTmplFilename()
	ctrl.renderTemplate(w, tmpfile, data)
}

func getRequestForm(r *http.Request) (*RequestForm, error) {
	r.ParseForm()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return nil, fmt.Errorf("invalid name")
	}

	idNumber := strings.TrimSpace(r.FormValue("id-number"))
	if idNumber == "" {
		return nil, fmt.Errorf("invalid id-number")
	}

	birthYear := strings.TrimSpace(r.FormValue("birth-year"))
	if birthYear == "" {
		return nil, fmt.Errorf("invalid date")
	}

	birthMonth := strings.TrimSpace(r.FormValue("birth-month"))
	if birthMonth == "" {
		return nil, fmt.Errorf("invalid date")
	}

	birthDay := strings.TrimSpace(r.FormValue("birth-day"))
	if birthDay == "" {
		return nil, fmt.Errorf("invalid date")
	}

	address := strings.TrimSpace(r.FormValue("address"))
	if address == "" {
		return nil, fmt.Errorf("empty address")
	}

	mobileNumber := strings.TrimSpace(r.FormValue("mobile-number"))

	return &RequestForm{
		Name:         name,
		IdNumber:     idNumber,
		BirthYear:    birthYear,
		BirthMonth:   birthMonth,
		BirthDay:     birthDay,
		Address:      sanitizeAddress(address),
		MobileNumber: mobileNumber,
	}, nil
}

func (ctrl *Controller) ThankYou(w http.ResponseWriter, r *http.Request, name string) {
	l := ctrl.GetRecallLegislator(name)
	if l == nil || l.RecallStatus != RecallStatusOngoing {
		http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
		return
	}

	ctrl.renderTemplate(w, "thank-you.html", map[string]interface{}{
		"BaseURL":        ctrl.AppBaseURL.String(),
		"ParticipateURL": l.ParticipateURL,
		"CalendarURL":    l.CalendarURL,
		"CsoURL":         l.CsoURL,
	})
}

func (ctrl *Controller) MParticipate(w http.ResponseWriter, r *http.Request) {
	ctrl.renderTemplate(w, "mayor-fill-form.html", map[string]interface{}{
		"BaseURL":          ctrl.AppBaseURL.String(),
		"PreviewURL":       ctrl.AppBaseURL.JoinPath("mayor", "preview").String(),
		"Address":          MayorCity,
		"TurnstileSiteKey": ctrl.TurnstileSiteKey,
	})
}

func (ctrl *Controller) MPreviewLocalForm(w http.ResponseWriter, r *http.Request) {
	qp, err := getRequestForm(r)
	if err != nil {
		ctrl.renderTemplate(w, "4xx.html", ViewHttp4xxError{
			HttpStatusCode: http.StatusBadRequest,
			ErrorMessage:   "輸入有誤",
			ReturnURL:      ctrl.AppBaseURL.String(),
		})
		return
	}

	data, err := qp.ToMayorPreviewData(ctrl.Config)
	if err != nil {
		ctrl.renderTemplate(w, "4xx.html", ViewHttp4xxError{
			HttpStatusCode: http.StatusBadRequest,
			ErrorMessage:   err.Error(),
			ReturnURL:      ctrl.AppBaseURL.String(),
		})
		return
	}
	ctrl.renderTemplate(w, "stage-2-"+MayorName+".html", data)
}

func (ctrl *Controller) MThankYou(w http.ResponseWriter, r *http.Request) {
	ctrl.renderTemplate(w, "mayor-thank-you.html", map[string]interface{}{
		"BaseURL":        ctrl.AppBaseURL.String(),
		"ParticipateURL": ctrl.AppBaseURL.JoinPath("mayor"),
		"CsoURL":         "https://www.facebook.com/hc.thebigrecall",
	})
}

func (ctrl *Controller) VerifyTurnstile(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	token := r.FormValue("cf-turnstile-response")
	if token == "" {
		ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL, ctrl.AppBaseURL))
		return false
	}
	success, err := ctrl.VerifyTurnstileToken(token)
	if err != nil || !success {
		ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusForbidden, "驗證失敗，請回到首頁重新輸入", ctrl.AppBaseURL, ctrl.AppBaseURL))
		return false
	}
	return true
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (ctrl *Controller) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if ctrl.AppEnv == AppEnvProduction {
		if err := ctrl.Templates.ExecuteTemplate(w, name, data); err != nil {
			http.Error(w, "Template rendering error", http.StatusInternalServerError)
		}
	} else {
		if t, err := template.ParseFiles("templates/tmpl.html", "templates/"+name); err != nil {
			http.Error(w, fmt.Errorf("Template parsing error: %v", err).Error(), http.StatusInternalServerError)
		} else if err := t.ExecuteTemplate(w, name, data); err != nil {
			http.Error(w, fmt.Errorf("Template rendering error: %v", err).Error(), http.StatusInternalServerError)
		}
	}
}

type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

type SitemapURLSet struct {
	XMLName     xml.Name      `xml:"urlset"`
	Xmlns       string        `xml:"xmlns,attr"`
	SitemapURLs []*SitemapURL `xml:"url"`
}

func (ctrl *Controller) Ping(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "v0.0.1"})
}

func (ctrl *Controller) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/robots.txt")
	if err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	tmpl.Execute(w, map[string]interface{}{
		"BaseURL":       ctrl.AppBaseURL.String(),
		"DisallowPaths": ctrl.DisallowPaths,
	})
}

func (ctrl *Controller) Sitemap(w http.ResponseWriter, r *http.Request) {
	date := "2025-03-02"
	urls := []*SitemapURL{
		{ctrl.AppBaseURL.String(), date, "daily", "1.0"},
		{ctrl.AppBaseURL.JoinPath("authorization-letter").String(), "2025-02-26", "yearly", "1.0"},
		{ctrl.AppBaseURL.JoinPath("mayor").String(), "2025-03-12", "weekly", "0.9"},
		{ctrl.AppBaseURL.JoinPath("mayor", "thank-you").String(), "2025-03-12", "weekly", "0.9"},
	}

	for _, l := range ctrl.RecallLegislators {
		if l.RecallStatus == "ONGOING" {
			urls = append(urls,
				&SitemapURL{l.ParticipateURL.String(), date, "weekly", "0.9"},
				&SitemapURL{l.ParticipateURL.JoinPath("thank-you").String(), date, "weekly", "0.8"},
			)
		}
	}

	sitemap := SitemapURLSet{
		Xmlns:       "http://www.sitemaps.org/schemas/sitemap/0.9",
		SitemapURLs: urls,
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	xml.NewEncoder(w).Encode(sitemap)
}

func (ctrl *Controller) GetAsset(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/assets/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	filePath := path.Join("assets", parts[0], parts[1])
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	if ctrl.AppEnv == AppEnvProduction {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	http.ServeFile(w, r, filePath)
}

func (ctrl *Controller) LegislatorRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/legislators/")
	parts := strings.Split(path, "/")

	if len(parts) == 1 && r.Method == http.MethodGet {
		ctrl.Participate(w, r, parts[0])
		return
	}

	if len(parts) == 2 {
		name := parts[0]
		switch parts[1] {
		case "preview":
			if r.Method == http.MethodPost {
				if !ctrl.VerifyTurnstile(w, r) {
					ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusBadRequest, "不合法的請求", ctrl.AppBaseURL, ctrl.AppBaseURL))
					return
				} else {
					ctrl.PreviewLocalForm(w, r, name)
					return
				}
			}
		case "thank-you":
			if r.Method == http.MethodGet {
				ctrl.ThankYou(w, r, name)
				return
			}
		}
	}

	ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
}

type RequestQuerySearchRecallConstituency struct {
	MunicipalityId uint64
	DistrictId     *uint64
	WardId         *uint64
}

type RespSearchRecallConstituency struct {
	Message string                          `json:"message"`
	Result  *ResultSearchRecallConstituency `json:"result,omitempty"`
}

type ResultSearchRecallConstituency struct {
	Divisions   Divisions         `json:"divisions,omitempty"`
	Legislators RecallLegislators `json:"legislators,omitempty"`
}

type RequestForm struct {
	Name         string
	IdNumber     string
	BirthYear    string
	BirthMonth   string
	BirthDay     string
	Address      string
	MobileNumber string
}

type RequestUriStageLegislator struct {
	Name  string
	Stage uint64
}

func (ctrl *Controller) PreviewOriginalLocalForm(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/preview/stages/"), "/")
	if len(parts) != 2 {
		http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
		return
	}

	name := parts[1]
	stage, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
		return
	}

	var data *PreviewData
	if name != MayorName {
		l := ctrl.GetRecallLegislator(name)
		if l == nil {
			http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
			return
		}

		data = &PreviewData{
			BaseURL:          ctrl.AppBaseURL.String(),
			ParticipateURL:   l.ParticipateURL,
			RedirectURL:      l.ParticipateURL.JoinPath("thank-you").String(),
			PoliticianName:   name,
			ConstituencyName: l.ConstituencyName,
			RecallStage:      stage,
			Name:             "邱吉爾",
			IdNumber:         IdNumber{"A", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
			BirthYear:        "888",
			BirthMonth:       "11",
			BirthDate:        "30",
			MobileNumber:     "0987654321",
			Address:          "某某市某某區某某里某某路三段 123 號七樓一段超長的地址一段超長的地址一段超長的地址一段超長的地址一段超長的地址",
		}
	} else {
		participateURL := ctrl.AppBaseURL.JoinPath("mayor")
		data = &PreviewData{
			BaseURL:          ctrl.AppBaseURL.String(),
			ParticipateURL:   participateURL,
			RedirectURL:      participateURL.JoinPath("thank-you").String(),
			PoliticianName:   name,
			ConstituencyName: MayorCity,
			RecallStage:      stage,
			Name:             "邱吉爾",
			IdNumber:         IdNumber{"A", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
			BirthYear:        "888",
			BirthMonth:       "11",
			BirthDate:        "30",
			MobileNumber:     "0987654321",
			Address:          "某某市某某區某某里某某路三段 123 號七樓一段超長的地址一段超長的地址一段超長的地址一段超長的地址一段超長的地址",
		}
	}

	if stage == 2 {
		tmpl := fmt.Sprintf("stage-2-%s.html", name)
		ctrl.renderTemplate(w, tmpl, data)
		return
	}

	ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
}

type PreviewData struct {
	BaseURL          string
	ParticipateURL   *url.URL
	RedirectURL      string
	PoliticianName   string
	ConstituencyName string
	RecallStage      uint64
	ImagePrefix      string
	Name             string
	IdNumber         IdNumber
	BirthYear        string
	BirthMonth       string
	BirthDate        string
	MobileNumber     string
	Address          string
}

type IdNumber struct {
	D0 string
	D1 string
	D2 string
	D3 string
	D4 string
	D5 string
	D6 string
	D7 string
	D8 string
	D9 string
}

func (r RequestForm) ToPreviewData(cfg *Config, up *RequestUriStageLegislator, l *RecallLegislator) (*PreviewData, error) {
	if !isValidIdNumber(r.IdNumber) {
		return nil, fmt.Errorf("身份證輸入錯誤")
	}

	if l.RecallStage == 1 {
		if r.MobileNumber != "" && !isValidMobileNumber(r.MobileNumber) {
			return nil, fmt.Errorf("手機號碼輸入錯誤")
		}
	}

	stage := strconv.FormatUint(up.Stage, 10)
	redirectURL := l.ParticipateURL.JoinPath("thank-you")
	imagePrefix := fmt.Sprintf("stage-%s-%s", stage, up.Name)

	data := &PreviewData{
		BaseURL:          cfg.AppBaseURL.String(),
		ParticipateURL:   l.ParticipateURL,
		RedirectURL:      redirectURL.String(),
		PoliticianName:   up.Name,
		ConstituencyName: l.ConstituencyName,
		RecallStage:      up.Stage,
		ImagePrefix:      imagePrefix,
		Name:             r.Name,
		BirthYear:        r.BirthYear,
		BirthMonth:       r.BirthMonth,
		BirthDate:        r.BirthDay,
		MobileNumber:     r.MobileNumber,
		Address:          r.Address,
	}

	for i := 0; i < len(r.IdNumber); i++ {
		switch i {
		case 0:
			data.IdNumber.D0 = string(r.IdNumber[i])
		case 1:
			data.IdNumber.D1 = string(r.IdNumber[i])
		case 2:
			data.IdNumber.D2 = string(r.IdNumber[i])
		case 3:
			data.IdNumber.D3 = string(r.IdNumber[i])
		case 4:
			data.IdNumber.D4 = string(r.IdNumber[i])
		case 5:
			data.IdNumber.D5 = string(r.IdNumber[i])
		case 6:
			data.IdNumber.D6 = string(r.IdNumber[i])
		case 7:
			data.IdNumber.D7 = string(r.IdNumber[i])
		case 8:
			data.IdNumber.D8 = string(r.IdNumber[i])
		case 9:
			data.IdNumber.D9 = string(r.IdNumber[i])
		}
	}

	return data, nil
}

func (r RequestForm) ToMayorPreviewData(cfg *Config) (*PreviewData, error) {
	if !isValidIdNumber(r.IdNumber) {
		return nil, fmt.Errorf("身份證輸入錯誤")
	}

	redirectURL := cfg.AppBaseURL.JoinPath("mayor", "thank-you")
	imagePrefix := "stage-2-" + MayorName

	data := &PreviewData{
		BaseURL:          cfg.AppBaseURL.String(),
		ParticipateURL:   cfg.AppBaseURL.JoinPath("mayor"),
		RedirectURL:      redirectURL.String(),
		PoliticianName:   MayorName,
		ConstituencyName: MayorCity,
		RecallStage:      2,
		ImagePrefix:      imagePrefix,
		Name:             r.Name,
		BirthYear:        r.BirthYear,
		BirthMonth:       r.BirthMonth,
		BirthDate:        r.BirthDay,
		MobileNumber:     r.MobileNumber,
		Address:          r.Address,
	}

	for i := 0; i < len(r.IdNumber); i++ {
		switch i {
		case 0:
			data.IdNumber.D0 = string(r.IdNumber[i])
		case 1:
			data.IdNumber.D1 = string(r.IdNumber[i])
		case 2:
			data.IdNumber.D2 = string(r.IdNumber[i])
		case 3:
			data.IdNumber.D3 = string(r.IdNumber[i])
		case 4:
			data.IdNumber.D4 = string(r.IdNumber[i])
		case 5:
			data.IdNumber.D5 = string(r.IdNumber[i])
		case 6:
			data.IdNumber.D6 = string(r.IdNumber[i])
		case 7:
			data.IdNumber.D7 = string(r.IdNumber[i])
		case 8:
			data.IdNumber.D8 = string(r.IdNumber[i])
		case 9:
			data.IdNumber.D9 = string(r.IdNumber[i])
		}
	}

	return data, nil
}
