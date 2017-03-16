all: test build install

build:
	go build

test:
	go test -v ./...

install:
	mkdir -p ~/bin/
	mkdir -p ~/.config/systemd/user/
	mkdir -p ~/.config/godown/
	cp godown ~/bin/
	cp godown.service ~/.config/systemd/user/
	cp godown.timer ~/.config/systemd/user/
	cp config.json.example ~/.config/godown/
	chmod 600 ~/.config/godown/config.json.example
	systemctl --user enable godown.timer
	systemctl --user start godown.timer
	sed -i -e 's#^ExecStart.*#ExecStart='"${HOME}/bin/godown ${HOME}/.config/godown/config.json#" godown.service
