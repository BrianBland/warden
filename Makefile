MAKEFILE_DIR=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# .PHONY: all build clean

print-%: ; @echo $*=$($*)

.PHONY: clean all build
.DEFAULT: default
all: build

build:
	docker run --rm -v $(MAKEFILE_DIR):/go/src/github.com/BrianBland/warden -e "GOPATH=/go/src/github.com/BrianBland/warden/Godeps/_workspace:/go" golang:1.4.2 go build -o /go/src/github.com/BrianBland/warden/warden /go/src/github.com/BrianBland/warden/cmd/warden/warden.go

clean:
	rm $(MAKEFILE_DIR)/warden
