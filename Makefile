VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
WAILS := $(shell which wails 2>/dev/null || echo $(HOME)/go/bin/wails)

# ── CLI ──────────────────────────────────────────────────────────────

.PHONY: build install clean uninstall

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o ck ./cmd/claude-kit/

install: build
	cp ck /usr/local/bin/ck 2>/dev/null || cp ck $(HOME)/bin/ck
	cp ck /usr/local/bin/claude-kit 2>/dev/null || cp ck $(HOME)/bin/claude-kit
	mkdir -p $(HOME)/.bmad/templates
	rsync -a --delete project-template/.claude/ $(HOME)/.bmad/templates/
	@echo "✓ CLI $(VERSION) installed"

clean:
	rm -f ck claude-kit-ui
	rm -rf cmd/claude-kit-ui/build/bin/
	rm -rf cmd/claude-kit-ui/frontend/dist/

uninstall:
	rm -f /usr/local/bin/ck /usr/local/bin/claude-kit $(HOME)/bin/ck $(HOME)/bin/claude-kit
	rm -rf $(HOME)/.bmad/templates
	@echo "✓ Uninstalled"

# ── Desktop App ──────────────────────────────────────────────────────

.PHONY: app app-install app-clean

app:
	cd cmd/claude-kit-ui/frontend && npm install --silent
	cd cmd/claude-kit-ui && $(WAILS) build
	@echo "✓ App built: cmd/claude-kit-ui/build/bin/Claude Kit.app"

app-install: app
	rm -rf /Applications/Claude\ Kit.app
	cp -R "cmd/claude-kit-ui/build/bin/Claude Kit.app" /Applications/
	@echo "✓ App installed to /Applications/Claude Kit.app"

app-clean:
	rm -rf cmd/claude-kit-ui/build/bin/
	rm -rf cmd/claude-kit-ui/frontend/dist/
	rm -rf cmd/claude-kit-ui/frontend/node_modules/

# ── Release artifact (zipped .app for GitHub releases) ───────────────

.PHONY: app-zip

app-zip: app
	cd cmd/claude-kit-ui/build/bin && zip -r ../../../../Claude-Kit-macOS-arm64.zip "Claude Kit.app"
	@echo "✓ Claude-Kit-macOS-arm64.zip ready for release"

# ── All ──────────────────────────────────────────────────────────────

.PHONY: all

all: build app
	@echo "✓ CLI + App built"

# ── Test ─────────────────────────────────────────────────────────────

.PHONY: test

test:
	go test ./internal/... ./cmd/claude-kit/...
