#! /usr/bin/make
.PHONY: all build build-linux build-windows test clean package

# 定义函数生成二进制名字
define binary_name
easy-check-$(1)-$(2)
endef

define ui_binary_name
easy-check-ui-$(1)-$(2)
endef

# 获取 Git 版本号
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
GIT_VERSION := $(if $(GIT_TAG),$(GIT_TAG),$(shell git describe --tags --always --dirty))

# 编译参数
LDFLAGS ?= -ldflags "-X main.version=$(GIT_VERSION)"

# 默认目标
all: build

# 编译目标
build: build-cmd build-ui

# 编译 cmd 可执行文件
build-cmd: build-linux-cmd build-windows-cmd

# 编译 Wails 应用
build-ui: build-linux-ui build-windows-ui

build-linux-ui:
	GOOS=linux GOARCH=amd64 wails build -ldflags "-X main.version=$(GIT_VERSION)" -platform linux/amd64 -o $(call ui_binary_name,linux,amd64)

build-windows-ui:
	GOOS=windows GOARCH=amd64 wails build -ldflags "-X main.version=$(GIT_VERSION)" -platform windows/amd64 -o $(call ui_binary_name,windows,amd64).exe -skipbindings
# GOOS=windows GOARCH=amd64 wails build -ldflags "-X main.version=$(GIT_VERSION)" -platform windows/amd64 -o $(call ui_binary_name,windows,amd64).exe

# 编译 Linux 可执行文件
build-linux-cmd:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(call binary_name,linux,amd64) cmd/main.go

# 编译 Windows 可执行文件
build-windows-cmd:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(call binary_name,windows,amd64).exe cmd/main.go

# 测试目标
test:
	go test ./...

# 清理目标
clean:
	rm -rf bin
	rm -f easy-check.zip

# 通用打包函数
define package_binary
	mkdir -p $(1)/bin
	if [ -f bin/$(2) ]; then \
		cp bin/$(2) $(1)/bin/; \
	elif [ -f build/bin/$(2) ]; then \
		cp build/bin/$(2) $(1)/bin/; \
	fi
	cp -r configs scripts README.md $(1)/
	zip -r $(1).zip $(1)
	rm -rf $(1)
endef

# 打包目标
package: package-linux-cmd package-windows-cmd package-linux-ui package-windows-ui

package-linux-cmd: build-linux-cmd
	$(call package_binary,easy-check-cmd-linux,$(call binary_name,linux,amd64))

package-windows-cmd: build-windows-cmd
	$(call package_binary,easy-check-cmd-windows,$(call binary_name,windows,amd64).exe)

package-linux-ui: build-linux-ui
	$(call package_binary,easy-check-ui-linux,$(call ui_binary_name,linux,amd64))

package-windows-ui: build-windows-ui
	$(call package_binary,easy-check-ui-windows,$(call ui_binary_name,windows,amd64).exe)


format:
	goimports -w .

wire:
	wire ./internal/wire

goreleaser:
	goreleaser release "--clean" "--snapshot"
