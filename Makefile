.PHONY: all clean docker
GO_BUILD=CGO_ENABLED=0 go build


all: gokoori gokoori.exe

gokoori: koori.go
	$(GO_BUILD)

gokoori.exe: koori.go
	GOOS=windows $(GO_BUILD)

clean:
	rm -f gokoori
	rm -f gokoori.exe

docker:
	docker run -tiP -v $(shell readlink -f cruise-config.xml):/etc/go/cruise-config.xml -v  $(shell readlink -f go-users):/etc/go/go-users -p 8153:8153 gocd/gocd-server

