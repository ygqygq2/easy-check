.PHONY: all build test clean package

# 默认目标
all: build

# 编译目标
build:
	go build -o bin/easy-check cmd/main.go

# 测试目标
test:
	go test ./...

# 清理目标
clean:
	rm -rf bin
	rm -f easy-check.zip

# 打包目标
package: build
	zip -r easy-check.zip bin/easy-check configs/config.yaml
