build:
	go build -ldflags="-w -s" -gcflags=all="-l -B" -o ./bin/StatiStream ./cli/cli.go

run: build
	./bin/StatiStream stream config.yaml
