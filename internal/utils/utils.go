package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type Utils struct {
	Logger *slog.Logger
}

func NewUtils(logger *slog.Logger) *Utils {
	return &Utils{Logger: logger}
}

func (u *Utils) reportServerError(r *http.Request, err error) {
	var (
		message = err.Error()
		method  = r.Method
		url     = r.URL.String()
		trace   = string(debug.Stack())
	)

	requestAttrs := slog.Group("request", "method", method, "url", url)
	u.Logger.Error(message, requestAttrs, "trace", trace)
}

func (u *Utils) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	u.reportServerError(r, err)

	message := fmt.Sprintf("The server encountered a problem and could not process your request: %s", err.Error())
	http.Error(w, message, http.StatusInternalServerError)
}

func (u *Utils) NotFound(w http.ResponseWriter, _ *http.Request) {
	message := "The requested resource could not be found"
	http.Error(w, message, http.StatusNotFound)
}

func (u *Utils) BadRequest(w http.ResponseWriter, _ *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
