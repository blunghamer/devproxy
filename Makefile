GOARCH=amd64

all: windows linux
build: linux

windows:
	GOOS=windows GOARCH=$(GOARCH) go build -ldflags="-s -w" -o .build/windows/$(GOARCH)/devproxy.exe cmd/devproxy/main.go	

linux:
	GOOS=linux GOARCH=$(GOARCH) go build -ldflags="-s -w" -o .build/linux/$(GOARCH)/devproxy cmd/devproxy/main.go	

linuxinstall: linux
	sudo rm -f /usr/local/bin/devproxy
	sudo cp .build/linux/amd64/devproxy /usr/local/bin/
	mkdir -p ~/.devproxy
	cp devproxy.yaml ~/.devproxy/
	# enable service config for devproxy
	mkdir -p ~/.config/systemd/user
	rm -f ~/.config/systemd/user/devproxy.service
	cp devproxy.service ~/.config/systemd/user/devproxy.service
	systemctl --user enable devproxy.service
	systemctl --user start devproxy.service

checkservice:
	systemctl --user status devproxy.service