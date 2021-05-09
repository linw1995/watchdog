fasttest:
	go test ./... -coverprofile=coverage.out -v
	go tool cover -func=coverage.out
fulltest:
	go test ./... -bench=. -coverprofile=coverage.out -v
	go tool cover -func=coverage.out
