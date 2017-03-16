all: test build install

build:
	go build

test:
	go test -v ./...
	echo $(HOME)

install:
	mkdir -p ~/bin/
	mkdir -p ~/.config/systemd/user/
	mkdir -p ~/.config/godown/
	cp godown.service ~/.config/systemd/user/
	cp godown.timer ~/.config/systemd/user/
	cp config.json.example ~/.config/godown/
	systemctl --user enable godown.timer
	systemctl --user start godown.timer
	sed -i -e 's#^ExecStart.*#ExecStart='"${HOME}/bin/godown ${HOME}/.config/godown/config.json#" godown.service
