# ── Pinned asset versions ─────────────────────────────────────────────────────
# Bump these lines to upgrade. DaisyUI and Tailwind versions are also in
# package.json ("^4" / "^5" tracks the latest patch within that major).
HTMX_VERSION := 2.0.4
HTMX_EXT_RT_VERSION := 2.0.10

# ── Derived paths ─────────────────────────────────────────────────────────────
HTMX_JS        := static/htmx.min.js
APP_CSS        := static/app.css
HTMX_EXT_RT_JS := static/htmx-ext-response-targets.js
BINARY         := send2ereader

.PHONY: setup deps htmx css css-watch templ-gen generate build run dev

# First-time setup: install npm packages + templ CLI, then download all assets.
setup: deps htmx
	go install github.com/a-h/templ/cmd/templ@latest

# Install / update DaisyUI and Tailwind CLI from package.json.
deps:
	npm install

# Download a pinned HTMX release. Re-run after changing HTMX_VERSION.
htmx:
	mkdir -p static
	curl -fsSL "https://unpkg.com/htmx.org@$(HTMX_VERSION)/dist/htmx.min.js" \
		-o $(HTMX_JS)
	curl -fsSL "https://unpkg.com/htmx.org@$(HTMX_EXT_RT_VERSION)/dist/ext/response-targets.js" \
		-o $(HTMX_EXT_RT_JS)

# Build app.css by scanning templates/ for class names.
# Re-run whenever you add new Tailwind/DaisyUI classes to .templ files.
css:
	mkdir -p static
	npx @tailwindcss/cli -i ./input.css -o $(APP_CSS) --minify

# Watch mode: auto-rebuild CSS on template changes (useful during development).
css-watch:
	npx @tailwindcss/cli -i ./input.css -o $(APP_CSS) --watch

# Generate Go code from .templ files.
templ-gen:
	templ generate

# Full code-generation step (CSS + templ).
generate: css templ-gen

# Compile the binary (requires generate + htmx to have been run at least once).
build: generate
	go build -o $(BINARY) .

# Run without producing a binary (still regenerates CSS + templates each time).
run: generate
	go run .

# Development: run CSS watcher and the server in parallel.
# Requires air (go install github.com/air-verse/air@latest) for hot reload.
dev:
	$(MAKE) templ-gen
	$(MAKE) -j2 css-watch dev-server

dev-server:
	air
