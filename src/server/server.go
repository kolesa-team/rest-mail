package server

import (
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"../logger"

	"github.com/Sirupsen/logrus"
	"github.com/braintree/manners"
	hCli "github.com/endeveit/go-snippets/cli"
	"github.com/endeveit/go-snippets/config"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"gopkg.in/gomail.v2"
)

type Server struct {
	running, enableAccessLog bool
	server                   *manners.GracefulServer
	mailer                   gomail.SendCloser
	dialer                   *gomail.Dialer
	mailChan                 chan *gomail.Message
	isMailerOpen             bool
	mutext                   sync.RWMutex
}

func NewServer() *Server {
	s := Server{
		mailChan: make(chan *gomail.Message),
	}

	host, err := config.Instance().String("smtp", "host")
	hCli.CheckError(err)

	port, err := config.Instance().Int("smtp", "port")
	hCli.CheckError(err)

	user, err := config.Instance().String("smtp", "user")
	hCli.CheckError(err)

	password, err := config.Instance().String("smtp", "password")
	hCli.CheckError(err)

	s.dialer = gomail.NewDialer(host, port, user, password)

	return &s
}

func (s *Server) Listen(done, stop chan struct{}) {
	address, err := config.Instance().String("http", "address")
	hCli.CheckError(err)

	s.server = manners.NewWithServer(&http.Server{
		Addr:        address,
		Handler:     s.mux(),
		ReadTimeout: time.Duration(1) * time.Second,
	})
	s.server.SetKeepAlivesEnabled(true)

	go func() {
		for {
			time.Sleep(time.Duration(1) * time.Second)
			if _, ok := <-stop; ok {
				s.server.Close()
			}
		}
	}()

	go func() {
		var err error

		for {
			select {
			case m, ok := <-s.mailChan:
				if !ok {
					return
				}

				if !s.isMailerOpen {
					if s.mailer, err = s.dialer.Dial(); err != nil {
						logger.Instance().WithFields(logrus.Fields{
							"error": err,
						}).Error("Error while connecting to SMTP server")
					} else {
						s.isMailerOpen = true
					}
				}

				if err = gomail.Send(s.mailer, m); err != nil {
					logger.Instance().WithFields(logrus.Fields{
						"error": err,
					}).Error("Error while sending message")
				} else {
					logger.Instance().WithFields(logrus.Fields{
						"from": m.GetHeader("From"),
						"to":   m.GetHeader("To"),
					}).Debug("Message sent")
				}
			case <-time.After(30 * time.Second):
				if s.isMailerOpen {
					if err := s.mailer.Close(); err != nil {
						logger.Instance().WithFields(logrus.Fields{
							"error": err,
						}).Error("Error while closing connection to SMTP server")
					}

					s.isMailerOpen = false
				}
			}
		}
	}()

	logger.Instance().Info("Starting daemon")
	s.server.ListenAndServe()

	done <- struct{}{}
}

func (s *Server) Close() {
	s.server.Close()
}

func (s *Server) mux() *web.Mux {
	m := web.New()

	m.Use(middleware.RealIP)
	m.Use(mwJson)

	if s.enableAccessLog {
		m.Use(mwLogger)
	}

	m.Use(mwRecoverer)

	m.Post("/", func(c web.C, w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		headers := make(map[string]string)

		for k, _ := range r.Header {
			k = strings.ToLower(k)

			if strings.HasPrefix(k, "x-") {
				headers[k[2:]] = r.Header.Get(k)
			}
		}

		for _, v := range strings.Split(headers["to"], ",") {
			message := gomail.NewMessage()

			message.SetBody(r.Header.Get("Content-Type"), string(body))
			message.SetAddressHeader("To", strings.Trim(v, " "), "")

			for key, value := range headers {
				if key != "to" {
					message.SetHeader(strings.ToUpper(key[:1])+key[1:], value)
				}
			}

			s.mailChan <- message
		}

		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	})

	return m
}
