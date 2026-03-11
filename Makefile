PROVIDER=provider
CONSUMER=consumer

.PHONY: all clean

all: $(PROVIDER) $(CONSUMER)

$(PROVIDER): provider.go
	go build -o $(PROVIDER) provider.go

$(CONSUMER): consumer.go
	go build -o $(CONSUMER) consumer.go

clean:
	rm -f $(PROVIDER) $(CONSUMER)
	killall provider
	killall consumer