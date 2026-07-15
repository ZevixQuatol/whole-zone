package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Code        string              `json:"code"`
	Message     string              `json:"message"`
	FieldErrors map[string][]string `json:"fieldErrors,omitempty"`
	Retryable   bool                `json:"retryable"`
	TraceID     string              `json:"traceId,omitempty"`
}

func Abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Error{Code: code, Message: message, Retryable: status >= http.StatusInternalServerError})
}
