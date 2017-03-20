.PHONY: all clean docker run

all: gokoori

gokoori:
	go build

clean:
	rm gokoori

docker:
	docker run -tiP -v $(shell readlink -f cruise-config.xml):/etc/go/cruise-config.xml -p 8153:8153 gocd/gocd-server

run: gokoori
	./gokoori
