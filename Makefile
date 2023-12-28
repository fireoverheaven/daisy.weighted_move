.PHONY: *
all: clean mod build

build:
	go build weighted_move.go

mod_init:
	-go mod init github.com/fireoverheaven/daisy.weighted_move
	-go mod tidy
	
mod_deps:
	-go get github.com/jmcvetta/randutil
	-go get github.com/lmittmann/tint
	-go get -u all

mod: mod_init mod_deps

clean:
	-rm go.mod
	-rm go.sum
	-rm weighted_move

push: mod
	git add -A
	-git commit -m "update"
	-git push
