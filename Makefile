.PHONY: build

build:
	docker stop file-manage
	docker remove file-manage
	docker build . -t file-manage
	docker run -itd -p 8080:8080 --name file-manage file-manage
	docker logs -f file-manage