MODULE      := github.com/nawodyaishan/ctg
VERSION     := 0.1.0
DIST        := dist
BUILDROOT   := $(DIST)/root
BIN_DIR     := $(BUILDROOT)/Library/Application Support/ColdTurkeyGuardian/bin
CLI_DIR     := $(BUILDROOT)/usr/local/bin
DAEMON_DIR  := $(BUILDROOT)/Library/LaunchDaemons

.PHONY: all build vet test tidy fmt pkg-unsigned clean

all: build

build:
	go build -o $(DIST)/guardiand ./cmd/guardiand
	go build -o $(DIST)/guardian ./cmd/guardian

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -l .

## pkg-unsigned: builds an unsigned .pkg for fast dev-loop installs (T03).
## Verify manually: sudo installer -pkg dist/guardian-dev.pkg -target /
##                  launchctl print system/com.coldturkeyguardian.primary
pkg-unsigned: export COPYFILE_DISABLE=1
pkg-unsigned: build
	rm -rf $(BUILDROOT)
	mkdir -p "$(BIN_DIR)" "$(CLI_DIR)" "$(DAEMON_DIR)"
	cp $(DIST)/guardiand "$(BIN_DIR)/guardiand"
	cp $(DIST)/guardian "$(CLI_DIR)/guardian"
	cp packaging/com.coldturkeyguardian.primary.plist "$(DAEMON_DIR)/"
	cp packaging/com.coldturkeyguardian.twin.plist "$(DAEMON_DIR)/"
	pkgbuild \
		--root "$(BUILDROOT)" \
		--scripts packaging/scripts \
		--identifier com.coldturkeyguardian.pkg \
		--version $(VERSION) \
		--install-location / \
		$(DIST)/guardian-dev-component.pkg
	productbuild \
		--distribution packaging/distribution.xml \
		--package-path $(DIST) \
		$(DIST)/guardian-dev.pkg

clean:
	rm -rf $(DIST)
