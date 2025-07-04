APP=URL_shortner_backend
APP_EXECUTABLE="./bin/$(APP)"


build:
	mkdir -p bin/
	go build -o $(APP_EXECUTABLE) ./cmd/backend/main.go
run:
	make build 
	chmod +x $(APP_EXECUTABLE)
	clear
	$(APP_EXECUTABLE)
test:
	go mod tidy
	go mod vendor 
	go clean -testcache
	go test ./tests/...
coverage:
	go mod tidy 
	go mod vendor
	go clean -testcache
	go test -v ./tests/... -coverprofile=coverage.out -coverpkg=./internal/...
	go tool cover -html=coverage.out
check-quality: 
	go fmt ./... 
	go vet ./... 
all:
	make check-quality
	make test 
	make build
stress: 
	chmod +x ./scripts/run-all-stressTest.sh
	./scripts/run-all-stressTest.sh
