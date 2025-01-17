package main

import (
	"net/url"
)

type ViewHttp4xxError struct {
	HttpStatusCode int
	ErrorMessage   string
	ReturnURL      string
}

func GetViewHttpError(code int, message string, returnURL *url.URL) *ViewHttp4xxError {
	return &ViewHttp4xxError{
		HttpStatusCode: code,
		ErrorMessage:   message,
		ReturnURL:      returnURL.String(),
	}
}
