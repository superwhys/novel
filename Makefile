SHELL := /bin/sh

APP_NAME ?= novel
VERSION ?= dev
GO ?= go
PNPM ?= pnpm
CGO_ENABLED ?= 0

FRONTEND_DIR := frontend
BUILD_DIR := build
DIST_DIR := dist
BACKEND_BINARY := $(BUILD_DIR)/$(APP_NAME)
RELEASE_NAME := $(APP_NAME)-$(VERSION)-linux-amd64
RELEASE_DIR := $(DIST_DIR)/$(RELEASE_NAME)
RELEASE_ARCHIVE := $(DIST_DIR)/$(RELEASE_NAME).tar.gz

.PHONY: all build frontend frontend-deps backend package deploy clean help

all: build

build: frontend backend ## 构建前端和后端

frontend-deps: ## 按 pnpm-lock.yaml 安装前端依赖
	cd "$(FRONTEND_DIR)" && $(PNPM) install --frozen-lockfile

frontend: frontend-deps ## 构建 Vue 前端，产物位于 frontend/dist
	cd "$(FRONTEND_DIR)" && $(PNPM) run build

backend: ## 构建 Go 后端，产物位于 build/novel
	mkdir -p "$(BUILD_DIR)"
	CGO_ENABLED="$(CGO_ENABLED)" GOOS="linux" GOARCH="amd64" \
		$(GO) build -trimpath -o "$(BACKEND_BINARY)" .

package: build ## 构建并打包前后端，生成 dist/*.tar.gz
	rm -rf "$(RELEASE_DIR)" "$(RELEASE_ARCHIVE)"
	mkdir -p "$(RELEASE_DIR)/frontend/dist" "$(RELEASE_DIR)/docs/story"
	cp "$(BACKEND_BINARY)" "$(RELEASE_DIR)/$(APP_NAME)"
	cp "$(FRONTEND_DIR)/dist/index.html" "$(RELEASE_DIR)/frontend/dist/index.html"
	cp -R "$(FRONTEND_DIR)/dist/assets" "$(RELEASE_DIR)/frontend/dist/assets"
	cp -R docs/story/content "$(RELEASE_DIR)/docs/story/content"
	COPYFILE_DISABLE=1 tar --no-xattrs -C "$(DIST_DIR)" -czf "$(RELEASE_ARCHIVE)" "$(RELEASE_NAME)"
	rm -rf "$(RELEASE_DIR)"
	@echo "发布包已生成：$(RELEASE_ARCHIVE)"

deploy: ## 部署后端和 Nginx 前端到 ali-prod
	bash scripts/deploy.sh "$(VERSION)"

clean: ## 清理所有构建产物
	rm -rf "$(BUILD_DIR)" "$(DIST_DIR)" "$(FRONTEND_DIR)/dist"

help: ## 查看可用命令
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
