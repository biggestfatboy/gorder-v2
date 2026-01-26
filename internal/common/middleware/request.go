package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

func RequestLog(l *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIn(c, l)
		defer requestOut(c, l)
		c.Next()
	}
}

func requestIn(c *gin.Context, l *logrus.Entry) {
	c.Set("request_start", time.Now())
	body := c.Request.Body
	bodyBytes, _ := io.ReadAll(body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	var compactJson bytes.Buffer
	_ = json.Compact(&compactJson, bodyBytes)
	logrus.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		"start":      time.Now().Unix(),
		logging.Args: compactJson.String(),
		"from":       c.RemoteIP(),
		"uri":        c.Request.RequestURI,
	}).Info("_request_in")
}
func requestOut(c *gin.Context, l *logrus.Entry) {
	response, _ := c.Get("response")
	start, _ := c.Get("request_start")
	startTime := start.(time.Time)
	logrus.WithContext(c.Request.Context()).WithFields(logrus.Fields{
		logging.Cost:     time.Since(startTime).Milliseconds(),
		logging.Response: response,
	}).Info("_request_out")
}
