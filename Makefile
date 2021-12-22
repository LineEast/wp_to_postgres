.PHONY: build, clean

build:
	go build -o="build/wp_to_postgres/main" "cmd/wp_to_postgres/main.go"

clean:
	rm -rf build

.DEFAULT_GOAL := start