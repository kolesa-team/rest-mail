package server

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"../logger"

	log "github.com/Sirupsen/logrus"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/mutil"
)

// Middleware для выставления заголовка типа ответа и кодировки
func mwJson(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		lw := mutil.WrapWriter(w)
		lw.Header().Set("Content-Type", "application/json; charset=utf-8")

		h.ServeHTTP(lw, r)
	}

	return http.HandlerFunc(fn)
}

// Middleware для логирования запросов
func mwLogger(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		lw := mutil.WrapWriter(w)
		ts := time.Now()

		h.ServeHTTP(lw, r)

		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}

		logger.Instance().WithFields(log.Fields{
			"remote_addr":   r.RemoteAddr,
			"method":        r.Method,
			"request_url":   r.URL.String(),
			"status":        lw.Status(),
			"response_time": time.Now().Sub(ts).String(),
		}).Info("Request processed")
	}

	return http.HandlerFunc(fn)
}

// Middleware для обработки panic()
func mwRecoverer(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Instance().WithFields(log.Fields{
					"remote_addr": r.RemoteAddr,
					"method":      r.Method,
					"request_url": r.URL.String(),
					"error":       fmt.Sprintf("%+v", err),
					"error_stack": string(debug.Stack()),
				}).Error("An error occurred while handling request.")

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
