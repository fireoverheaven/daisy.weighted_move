.PHONY: *
all: clean mod build

build:
	go build weighted_move.go && echo "build"

mod:
	-go mod init github.com/fireoverheaven/daisy.weighted_move && echo "mod :: init"
	-go mod tidy && echo "mod :: tidy"
	-go get github.com/jmcvetta/randutil  && echo "get :: randutil"
	-go get github.com/stretchr/testify && echo "get :: testify"
	-go get -u all && echo "get :: -u all"

clean:
	-rm main
	-rm go.mod
	-rm go.sum
	-rm weighted_move

push: clean
	git add -A
	-git commit -m "update"
	-git push
