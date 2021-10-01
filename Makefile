GOARCH=amd64

all: windows linux
build: linux

windows:
	GOOS=windows GOARCH=$(GOARCH) go build -ldflags="-s -w" -o .build/windows/$(GOARCH)/devproxy.exe cmd/devproxy/main.go	

linux:
	GOOS=linux GOARCH=$(GOARCH) go build -ldflags="-s -w" -o .build/linux/$(GOARCH)/devproxy cmd/devproxy/main.go	

checkservice:
	systemctl --user status devproxy.service