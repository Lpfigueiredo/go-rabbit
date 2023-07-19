.PHONY: env
env:
	cp .env.local .env

.PHONY: sender
sender:
	go run sender

.PHONY: consumer
consumer:
	go run consumer
