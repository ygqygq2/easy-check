.PHONY: all build build-linux build-windows test clean package

# 获取 Git 版本号
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
GIT_VERSION := $(if $(GIT_TAG),$(GIT_TAG),$(shell git describe --tags --always --dirty))

# 默认目标
all: build

# 编译目标
build: build-linux build-windows

# 编译 Linux 可执行文件
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(GIT_VERSION)" -o bin/easy-check cmd/main.go

# 编译 Windows 可执行文件
build-windows:
#GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -X main.version=$(GIT_VERSION)" -o bin/easy-check.exe cmd/main.go
	GOOS=windows GOARCH=amd64 gogio -ldflags "-H windowsgui -X main.version=$(GIT_VERSION)" \
  -buildmode=exe -icon=internal/assets/images/logo.png -target=windows -o bin/easy-check.exe ./cmd


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
