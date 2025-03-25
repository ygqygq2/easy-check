.PHONY: all build build-linux build-windows test clean package

# 默认目标
all: build

# 编译目标
build: build-linux build-windows

# 编译 Linux 可执行文件
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/easy-check cmd/main.go

# 编译 Windows 可执行文件
build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/easy-check.exe cmd/main.go

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
