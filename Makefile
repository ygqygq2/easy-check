.PHONY: all build build-linux build-windows test clean package

# 获取 Git 版本号
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
GIT_VERSION := $(if $(GIT_TAG),$(GIT_TAG),$(shell git describe --tags --always --dirty))

# 默认目标
all: build

# 编译目标
build: build-cmd build-ui

# 编译 cmd 可执行文件
build-cmd: build-linux-cmd build-windows-cmd

# 编译 Wails 应用
build-ui:
	wails build -ldflags "-X main.version=$(GIT_VERSION)" -platform linux/amd64 -platform windows/amd64

# 编译 Linux 可执行文件
build-linux-cmd:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(GIT_VERSION)" -o bin/easy-check cmd/main.go

# 编译 Windows 可执行文件
build-windows-cmd:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(GIT_VERSION)" -o bin/easy-check.exe cmd/main.go



# 测试目标
test:
	go test ./...

# 清理目标
clean:
	rm -rf bin
	rm -f easy-check.zip

# 打包目标
package: build
	mkdir -p easy-check
	cp -r bin easy-check/
	cp -r configs easy-check/
	cp -r scripts easy-check/
	cp README.md easy-check/
	zip -r easy-check.zip easy-check
	rm -rf easy-check

format:
	goimports -w .

wire:
	wire ./internal/wire
