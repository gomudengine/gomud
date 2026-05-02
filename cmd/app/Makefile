.PHONY: install start build-mac build-win build-linux build clean vendor-gomud-files clean-gomud-vendored

HOST ?= gomud.net
WS_URL ?= wss://$(HOST)/ws

install: vendor-gomud
	npm install

start: vendor-gomud
	npm start

build-mac: install
	npm run build:mac

build-win: install
	npm run build:win

build-linux: install
	npm run build:linux

build: install
	npm run build:mac
	npm run build:win
	npm run build:linux

clean: clean-gomud-vendored
	@echo "Clearing dist/ and node_modules/ ..."
	@rm -rf dist node_modules

clean-gomud-vendored:
	@echo "Clearing build_cache..."
	@rm -rf build_cache/*

vendor-gomud: clean-gomud-vendored
	@echo "Vendoring files from GoMud html..."
	@cp -rp ../../_datafiles/html/public/webclient-pure.html build_cache/
	@mkdir -p build_cache/static
	@cp -rp ../../_datafiles/html/public/static/js build_cache/static/
	@cp -rp ../../_datafiles/html/public/static/css build_cache/static/
	@echo "Replacing template vars with new values..."
	@sed -i '' 's/{{ .ASSET_BASE_URL }}//g' build_cache/webclient-pure.html
	@sed -i '' 's/{{ .CONFIG.Server.MudName }}//g' build_cache/webclient-pure.html
	@sed -i '' 's|="/static/|="static/|g' build_cache/webclient-pure.html
	@sed -i '' 's|src="/static/|src="static/|g' build_cache/webclient-pure.html
	@sed -i '' 's|WS_URL_OVERRIDE = "";|WS_URL_OVERRIDE = "$(WS_URL)";|g' build_cache/webclient-pure.html
