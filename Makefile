# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gero android ios gero-cross swarm evm all test clean
.PHONY: gero-linux gero-linux-386 gero-linux-amd64 gero-linux-mips64 gero-linux-mips64le
.PHONY: gero-linux-arm gero-linux-arm-5 gero-linux-arm-6 gero-linux-arm-7 gero-linux-arm64
.PHONY: gero-darwin gero-darwin-386 gero-darwin-amd64
.PHONY: gero-windows gero-windows-386 gero-windows-amd64

GOBIN = $(shell pwd)/build/bin
root=$(shell pwd)
GO ?= latest
GORUN = env GO111MODULE=on go run

PKG = ./...

gece:
	build/env.sh go run build/ci.go install ./cmd/gece
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gece\" to launch gece."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint $(PKG)

clean:
	./build/clean_go_build_cache.sh
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gece-cross: gece-linux gece-darwin gece-windows
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gece-*

gece-linux: gece-linux-amd640-v3 gece-linux-amd64-v4
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gece-linux-*

gece-linux-amd64-v3:
	build/env.sh linux-v3 go run build/ci.go xgo -- --go=$(GO) --out=gece-v3 --targets=linux/amd64 -v ./cmd/gece
	#build/env.sh linux-v3 go run build/ci.go xgo -- --go=$(GO) --out=bootnode-v3 --targets=linux/amd64 -v ./cmd/bootnode
	@echo "Linux centos amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gece-v3-linux-* | grep amd64

gece-linux-amd64-v4:
	build/env.sh linux-v4 go run build/ci.go xgo -- --go=$(GO) --out=gece-v4 --targets=linux/amd64 -v ./cmd/gece
	#build/env.sh linux-v3 go run build/ci.go xgo -- --go=$(GO) --out=bootnode-v4 --targets=linux/amd64 -v ./cmd/bootnode
	@echo "Linux  ubuntu amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gece-v4-linux-* | grep amd64

gece-darwin: gece-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gece-darwin-*


gece-darwin-amd64:
	build/env.sh darwin-amd64 go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gece
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gece-darwin-* | grep amd64

gece-windows: gece-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gece-windows-*

gece-windows-amd64:
	build/env.sh windows-amd64 go run build/ci.go xgo -- --go=$(GO)  --targets=windows/amd64 -v ./cmd/gece
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gece-windows-* | grep amd64

gecetx-darwin-amd64:
	build/env.sh darwin-amd64 go run build/ci.go xgo -- --go=$(GO) --out=gecetx  --targets=darwin/amd64 -v ./cmd/tx
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gecetx-darwin-* | grep amd64

gecetx-linux-amd64-v3:
	build/env.sh linux-v3 go run build/ci.go xgo -- --go=$(GO) --out=gecetx-v3 --targets=linux/amd64 -v ./cmd/tx
	@echo "Linux centos amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gecetx-v3-linux-* | grep amd64

gecetx-linux-amd64-v4:
	build/env.sh linux-v4 go run build/ci.go xgo -- --go=$(GO) --out=gecetx-v4 --targets=linux/amd64 -v ./cmd/tx
	@echo "Linux  ubuntu amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gecetx-v4-linux-* | grep amd64

gecetx-windows-amd64:
	build/env.sh windows-amd64 go run build/ci.go xgo -- --go=$(GO) --out=gecetx --targets=windows/amd64 -v ./cmd/tx
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gecetx-windows-* | grep amd64