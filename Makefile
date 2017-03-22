.PHONY: all clean docker

all: gokoori

gokoori: koori.go
	go build

clean:
	rm gokoori

docker:
	docker run -tiP -v $(shell readlink -f cruise-config.xml):/etc/go/cruise-config.xml -v  $(shell readlink -f go-users):/etc/go/go-users -p 8153:8153 gocd/gocd-server

