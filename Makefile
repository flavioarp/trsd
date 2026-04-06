APP_PROVIDER=provider
APP_CONSUMER=consumer

.PHONY: all build run clean

all: build

build:
	go build -o build/$(APP_PROVIDER) provider.go
	go build -o build/$(APP_CONSUMER) consumer.go

run-provider:
	./build/$(APP_PROVIDER)

run-provider-pdfa:
	./build/$(APP_PROVIDER) pdfa

run-consumer:
	./build/$(APP_CONSUMER)

clean:
	rm -f build/$(APP_PROVIDER) build/$(APP_CONSUMER)
	pkill -f $(APP_PROVIDER) || true
	pkill -f $(APP_CONSUMER) || true
