VERSION := $(shell cat VERSION)

default: xcompile

xcompile: linux_arm6 linux_arm7 linux_386 linux_amd64 mac

makedirs:
	rm -rf build
	mkdir -p build/linux/arm6/
	mkdir -p build/linux/arm7/
	mkdir -p build/linux/amd64/
	mkdir -p build/linux/386/
	mkdir -p build/macos/

linux_arm6:
	mkdir -p build/linux/arm7/
	rm -rf build/linux/arm7/*
	env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "-X main.$(VERSION)" -o build/linux/arm6/gmux -v

linux_arm7:
	mkdir -p build/linux/arm7/
	rm -rf build/linux/arm7/*
	env GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-X main.$(VERSION)" -o build/linux/arm7/gmux -v

linux_386:
	mkdir -p build/linux/386/
	rm -rf build/linux/386/*
	env GOOS=linux GOARCH=arm go build -ldflags "-X main.$(VERSION)" -o build/linux/386/gmux -v

linux_amd64:
	mkdir -p build/linux/amd64/
	rm -rf build/linux/amd64/*
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X main.$(VERSION)" -o build/linux/amd64/gmux -v

mac:
	mkdir -p build/macos/
	rm -rf build/macos/*
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.$(VERSION)" -o  build/macos/gmux -v

brew: mac
	./bin/brewify

