GOPATH := ${GOPATH}:$(shell pwd)

hijack:
# @echo "hello jack"

install:
    #maybe we should find other better package management system to shoot go get:(
    # @echo "shit"
build:
	@GOPATH=$(GOPATH) go install luaproxy
# buildarch:
#     for arch in arm 386 amd64; do for os in linux darwin freebsd; do\
#         go build -o bin/zoneproxy-$$os-$$arch zoneproxy &&\
#         go build -o bin/socks5-$$os-$$arch socks5 &&\
#         zip -r zoneproxy-$$os-$$arch.zip bin/zoneproxy-$$os-$$arch conf README.md; \
#     done done
test:
    #fake test, we should add test :(
    @GOPATH=$(GOPATH) go test dialer
clean:
	@rm -fr bin pkg
