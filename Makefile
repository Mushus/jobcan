.PHONY: fmt
.PHONY: run
.PHONY: build

fmt:
	gofmt -w ./

run:
	go run ./jobcan.go

build:
	if [ ! -e ./build ]; then mkdir ./build; fi;
	if [ ! -e ./build/release ]; then mkdir ./build/release; fi;
	GOARCH=amd64 GOOS=windows go build -o ./build/windows-amd64/jobcan.exe jobcan.go
	zip ./build/release/jobcan_windows_amd64.zip ./build/windows-amd64/jobcan.exe
	GOARCH=amd64 GOOS=darwin go build -o ./build/darwin-amd64/jobcan jobcan.go
	zip ./build/release/jobcan_macos_amd64.zip ./build/darwin-amd64/jobcan
	GOARCH=amd64 GOOS=linux go build -o ./build/linux-amd64/jobcan jobcan.go
	zip ./build/release/jobcan_linux_amd64.zip ./build/linux-amd64/jobcan
