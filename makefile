APP=URL_shortner_backend
APP_EXECUTABLE="./target/$(APP)"


build:
	mkdir -p target/
	go build -o $(APP_EXECUTABLE)
run:
	make build 
	chmod +x $(APP_EXECUTABLE)
	clear
	$(APP_EXECUTABLE)
test:
	go mod tidy
	go mod vendor 
	go test ./...
coverage:
	go mod tidy 
	go mod vendor
	go test -v ./... -coverprofile=coverage.out
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
