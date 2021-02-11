.PHONY: build clean

bin/nvidiasmi_exporter:
	@GOOS=linux GOARCH=amd64 go build -o bin/nvidiasmi_exporter src/*.go

build: bin/nvidiasmi_exporter

clean:
	@rm -rf ./bin
