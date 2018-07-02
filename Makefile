install:
	go build -o socker cmd/socker/socker.go
	chown root:root socker
	chmod +s socker
	mv socker /usr/bin/
.PHONY: clean
clean:
	-rm socker
