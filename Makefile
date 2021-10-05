.PHONY: test
test:
	docker-compose up -d
# testが失敗した場合は掃除してからstatus 1を返す
	go test  ./... || (docker-compose down && false) 
	docker-compose down