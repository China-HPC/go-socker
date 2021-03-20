build-linux64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o socker cmd/socker/socker.go
install:
	go build -o socker cmd/socker/socker.go
	chown root:root socker
	chmod +s socker
	mv socker /usr/bin/
	mkdir -p /var/lib/socker
	cp configs/images.yaml /var/lib/socker/
.PHONY: clean
clean:
	-rm socker
