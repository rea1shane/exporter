VERSION  = $(shell cat VERSION)
REVISION = $(shell git rev-parse HEAD)
BRANCH   = $(shell git rev-parse --abbrev-ref HEAD)
USER     = $(shell whoami)
HOSTNAME = $(shell hostname)
DATE     = $(shell date "+%Y%m%d-%H:%M:%S")

.PHONY: build
build:
	@go build -o example_exporter \
	  -ldflags \
	  "-X github.com/prometheus/common/version.Version=$(VERSION) \
		-X github.com/prometheus/common/version.Revision=$(REVISION) \
		-X github.com/prometheus/common/version.Branch=$(BRANCH) \
		-X github.com/prometheus/common/version.BuildUser=$(USER)@$(HOSTNAME) \
		-X github.com/prometheus/common/version.BuildDate=$(DATE)" \
	  .
