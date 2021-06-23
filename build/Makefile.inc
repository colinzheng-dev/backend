.PHONY: gen run dbtest mocks test

gen:
	go generate $(BASE)/...

run:
	gcloud config set project veganbaseapp
	docker-compose up -d

# VB_TEST_DB='postgres://localhost/vb_test?sslmode=disable'
dbtest:
	go test -v $(BASE)/test/db | ../../dev-tools/test-colours

mocks:
	mockery -recursive -name "$(shell echo "$(MOCKS)" | tr ' ' '|')"

test:
	go generate $(BASE)/...
	go test $(BASE)/...
