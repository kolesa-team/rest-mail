package main

import (
	"os"
	"syscall"

	"./logger"
	"./server"

	log "github.com/Sirupsen/logrus"
	"github.com/endeveit/go-snippets/config"
	gd "github.com/sevlyar/go-daemon"
	"github.com/urfave/cli"
)

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func main() {
	app := cli.NewApp()

	app.Name = "rest-mail"
	app.Usage = "REST mail sender"
	app.Version = "0.0.1"
	app.Author = "Igor Borodikhin"
	app.Email = "iborodikhin@gmail.com"
	app.Action = actionRun
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, b",
			Usage: "If provided, the service will be launched in debug mode",
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/rest-mail/config.cfg",
			Usage: "Path to the configuration file",
		},
	}

	app.Run(os.Args)
}

func actionRun(c *cli.Context) error {
	config.Instance(c.String("config"))

	if c.Bool("debug") {
		logger.Instance().Level = log.DebugLevel
	}

	gd.SetSigHandler(termHandler, syscall.SIGTERM)

	s := server.NewServer()
	go s.Listen(done, stop)

	return gd.ServeSignals()
}

// Обработчик SIGTERM
func termHandler(sig os.Signal) error {
	stop <- struct{}{}

	if sig == syscall.SIGTERM {
		<-done
	}

	return gd.ErrStop
}
