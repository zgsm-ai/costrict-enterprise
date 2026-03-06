# 如果版本发生变化需要修改这里的版本号，以及build.py中的版本号
VER    := 1.0.250803
OS     := $(shell go env GOOS)
ARCH   := $(shell go env GOARCH)
EXEEXT ?= 
ifeq (windows,$(OS))
EXEEXT := .exe
endif
APP    := smc$(EXEEXT)

build:
	python ./build.py --software $(VER) --os $(OS) --arch $(ARCH)

install:
	python ./build.py --software $(VER) --install --os $(OS) --arch $(ARCH)

.PHONY: build install
