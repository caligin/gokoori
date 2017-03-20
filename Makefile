.PHONY: docker

docker:
	docker run -tiP -v $(shell readlink -f cruise-config.xml):/etc/go/cruise-config.xml -p 8153:8153 gocd/gocd-server
