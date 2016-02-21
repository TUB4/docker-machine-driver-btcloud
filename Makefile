default: build

clean:
	rm bin/docker-machine-*

build:
	GOGC=off go build -o bin/docker-machine-driver-btcloud bin/main.go

install: build
	cp ./bin/docker-machine-driver-btcloud /usr/local/bin/
