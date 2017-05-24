export GOPATH=$(CURDIR)/.go

APP_NAME = rest-mail
DEBIAN_TMP = $(CURDIR)/deb
VERSION = `$(CURDIR)/out/$(APP_NAME) -v | cut -d ' ' -f 3`

$(CURDIR)/out/$(APP_NAME): $(CURDIR)/src/main.go
	go build -o $(CURDIR)/out/$(APP_NAME) $(CURDIR)/src/main.go

dep-install:
	go get github.com/endeveit/go-snippets/config
	go get github.com/endeveit/go-snippets/cli
	go get github.com/Sirupsen/logrus
	go get github.com/zenazn/goji/web/middleware
	go get github.com/braintree/manners
	go get github.com/sevlyar/go-daemon
	go get github.com/urfave/cli
	go get gopkg.in/gomail.v2

fmt:
	gofmt -s=true -w $(CURDIR)/src

run:
	go run -race $(CURDIR)/src/main.go -c=$(CURDIR)/data/config.cfg -b

strip: $(CURDIR)/out/$(APP_NAME)
	strip $(CURDIR)/out/$(APP_NAME)

clean:
	rm -f $(CURDIR)/out/*

debug:
	echo $(GOPATH)
