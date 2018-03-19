all:
	-mkdir bin
	GOPATH=$(PWD) go build -o bin/drivers src/drivers/*.go
	cp -r public bin