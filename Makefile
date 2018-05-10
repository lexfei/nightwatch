SHELL := /bin/bash 
BASEDIR = $(shell pwd)

all: nightwatch

nightwatch:
	@make -C ${BASEDIR}/cmd/
	@echo binary file is: ${BASEDIR}/nightwatch

gotool:
	@-gofmt -w  .
	@-go tool vet . |& grep -v vendor

clean:
	rm -f nightwatch
	find . -name "[._]*.s[a-w][a-z]" | xargs -i rm -f {}

help:
	@echo "all - make nightwatch & run go tool"
	@echo "nightwatch - make api"
	@echo "gotool - run gofmt & go too vet"
	@echo "clean - do some clean job"

.PHONY: all gotool clean help nightwatch
