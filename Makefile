GO = go
GOTEST = $(GO) test
GOCOVER = $(GO) tool cover

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: test/cover
test/cover:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCOVER) -func=coverage.out

.PHONY: test/cover/html
test/cover/html: test/cover
	$(GOCOVER) -html=coverage.out

.PHONY: test/full
test/full:
	$(GOTEST) -v -bench=. -coverprofile=coverage.out ./...
	$(GOCOVER) -func=coverage.out
