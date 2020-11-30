all: build

build:
	GO111MODULE=on go build -o proxydetection ./main.go 
