package main

import (
	"net/url"
)

type ViewHttp4xxError struct {
	BaseURL        string
	HttpStatusCode int
	ErrorMessage   string
	ReturnURL      string
}

func GetViewHttpError(code int, message string, baseURL, returnURL *url.URL) *ViewHttp4xxError {
	return &ViewHttp4xxError{
		BaseURL:        baseURL.String(),
		HttpStatusCode: code,
		ErrorMessage:   message,
		ReturnURL:      returnURL.String(),
	}
}
