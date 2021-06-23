MOCKS=PubSub Storage

.PHONY: mocks

mocks:
	mockery --recursive --name "$(shell echo "$(MOCKS)" | tr ' ' '|')"
