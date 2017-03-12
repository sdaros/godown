all: build install

build:
	go build

test:
	go test -v ./...

install:
	mkdir -p ~/bin/
	mkdir -p ~/.config/systemd/user/
	mkdir -p ~/.config/godown/
#	cp godown.service ~/.config/systemd/user/
# cp godown.time ~/.config/systemd/user/
# cp config.json ~/.config/godown/
	systemctl --user enable godown.timer
	systemctl --user start godown.timer
