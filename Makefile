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
	docker-compose up
