.PHONY: all clean test

SERVICES=\
	api-gateway \
	blob-service \
	category-service \
	email-service \
	item-service \
	search-service \
	site-service \
	user-service \
	cart-service \
	social-service \
	purchase-service \
	payment-service \
	webhook-service

all:
	@go clean ./...
	@go generate ./...
	@for s in $(SERVICES); do echo $$s; cd ./services/$$s; go build; cd ../..; done

clean:
	go clean ./...

test:
	go clean ./...
	go generate ./...
	go build ./...
	go test ./...

mocks:
	make -C chassis mocks
	@for s in $(SERVICES); do echo $$s; cd ./services/$$s; make mocks; cd ../..; done
