VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# ── CLI ──────────────────────────────────────────────────────────────

.PHONY: build install clean uninstall test

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o ck ./cmd/claude-kit/

install: build
	ln -sf $(CURDIR)/ck /usr/local/bin/ck 2>/dev/null || ln -sf $(CURDIR)/ck $(HOME)/bin/ck
	ln -sf $(CURDIR)/ck /usr/local/bin/claude-kit 2>/dev/null || ln -sf $(CURDIR)/ck $(HOME)/bin/claude-kit
	mkdir -p $(HOME)/.bmad/templates
	rsync -a --delete project-template/.claude/ $(HOME)/.bmad/templates/
	@echo "✓ CLI $(VERSION) installed (symlinked from $(CURDIR)/ck)"

clean:
	rm -f ck

uninstall:
	rm -f /usr/local/bin/ck /usr/local/bin/claude-kit $(HOME)/bin/ck $(HOME)/bin/claude-kit
	rm -rf $(HOME)/.bmad/templates
	@echo "✓ Uninstalled"

# ── Test ─────────────────────────────────────────────────────────────

test:
	go test ./internal/... ./cmd/claude-kit/...
