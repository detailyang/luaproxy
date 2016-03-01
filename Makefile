GOPATH := ${GOPATH}:$(shell pwd)

hijack:
	@echo "hello jack"

install:
    #maybe we should find other better package management system to shoot go get:(
build:
	@GOPATH=$(GOPATH) go install luaproxy
test:
    #fake test, we should add test :(
    @GOPATH=$(GOPATH) go test dialer
clean:
	@rm -fr bin pkg
