RELEASES=bin/hello-darwin-amd64 \
	 bin/hello-linux-amd64 \
	 bin/hello-linux-386 \
	 bin/hello-linux-arm \
	 bin/hello-windows-amd64.exe \
	 bin/hello-windows-386.exe \
	 bin/hello-solaris-amd64 

all: $(RELEASES)

bin/hello-%: GOOS=$(firstword $(subst -, ,$*))
bin/hello-%: GOARCH=$(subst .exe,,$(word 2,$(subst -, ,$*)))
bin/hello-%: $(wildcard *.go)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build \
	     -ldflags "-X main.osarch=$(GOOS)/$(GOARCH) -s -w" \
	     -buildmode=exe \
	     -tags release \
	     -o $@

clean:
	rm -rf bin