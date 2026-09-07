.PHONY: dev build docs clean

dev:
	air

docs:
	swag init -g cmd/api_info.go --parseDependency --parseInternal

build: docs
	go build -o tmp/main cmd/main.go

clean:
	rm -rf tmp/
