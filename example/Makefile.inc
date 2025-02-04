# Copyright 2021-2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be found
# in the LICENSE file.


#
# NOTE: That section has per project variables that have to be configured by the
# projects importing this Makefile.inc
#
# INFO: One project, multiple programs
project  ?=
# INFO: Program list for this project
programs ?=


#
# Start of general Makefile.inc (No modifications generally needed past this)
#
MAKEFLAGS += --no-print-directory
V         ?= @
PREFIX    ?= /usr/local
GO        ?= go
GOLINT    ?= golangci-lint

module    ?= $(shell go list -m)
os        ?= linux
arch      ?= amd64
testout   ?= tests.out
version   ?= \
	$(shell \
		(\
			git describe --tags --match 'v[0-9]*' --dirty --always --abbrev=4 2>/dev/null || \
			cat version 2>/dev/null || \
			echo "1.0.0" \
		) | sed -e "s/^v//;s/-/_/g;s/_/-/;s/_/./g"\
	)

all:: devel lint test

clean::
	@printf '\e[38;5;202m  CLEAN\e[m\n'
	$(V)rm -rf $(programs) $(testout) $(project).info *.$(version) *-debug.$(version)


dep-upgrade::
	@printf '\e[38;5;202m  GET\e[m       %s\n' "$(project)"
	$(V)$(GO) get -u ./...
	@printf '\e[38;5;202m  TIDY\e[m      %s\n' "$(project)"
	$(V)$(GO) mod tidy
	@printf '\e[38;5;202m  VENDOR\e[m    %s\n' "$(project)"
	$(V)GOWORK=off $(GO) mod vendor

devel::
	$(V)if test "$(shell $(GO) env GOWORK)" != "off"; then printf '\e[38;5;192m%s\e[m\n' "*** GO WORKSPACE ENABLED: release builds may fail (or vice versa)"; fi
	$(V)$(foreach p,$(programs), \
		printf '\e[38;5;202m  GO\e[m        %s\n' "$(p) -> $(p).$(version)" && \
		GOOS=$(os) GOARCH=$(arch) $(GO) build -tags osusergo,netgo -ldflags "-s -X $(module)/internal.ProgramVersion=$(version)" -o $(p).$(version) ./cmd/$(p); \
	)
	$(V)$(foreach p,$(programs), \
		printf '\e[38;5;202m  GO (DBG)\e[m  %s\n' "$(p) -> $(p)-debug.$(version)" && \
		GOOS=$(os) GOARCH=$(arch) $(GO) build -tags osusergo,netgo -ldflags "-X $(module)/internal.ProgramVersion=$(version)" -race -gcflags='all=-N -l' -o $(p)-debug.$(version) ./cmd/$(p); \
	)

install::
	$(V)sudo install -d $(PREFIX)/bin
	$(V)$(foreach p,$(programs), \
		printf '\e[38;5;202m  INSTALL\e[m   %s\n' "$(p) -> $(PREFIX)/bin" && \
		sudo install -m 755 $(p) $(PREFIX)/bin; \
	)

lint::
	@printf '\e[38;5;202m  VET\e[m       %s\n' "$(project)"
	$(V)$(GO) vet ./...
	@printf '\e[38;5;202m  LINT\e[m      %s\n' "$(project)"
	$(V)$(GOLINT) run --exclude-dirs grpcssh ./...

release::
	$(V)$(foreach p,$(programs), \
		printf '\e[38;5;202m  RELEASE\e[m   %s\n' "$(p)" && \
		GOOS=$(os) GOARCH=$(arch) GOWORK=off $(GO) build -tags release,osusergo,netgo -ldflags "-s -X $(module)/internal.ProgramVersion=$(version)" -o $(p) ./cmd/$(p); \
	)
	@printf '\e[38;5;202m  VET\e[m       %s\n' "$(project)"
	$(V)GOWORK=off $(GO) vet ./...
	@printf '\e[38;5;202m  LINT\e[m      %s\n' "$(project)"
	$(V)GOWORK=off $(GOLINT) run --exclude-dirs grpcssh ./...
	@printf '\e[38;5;202m  TEST\e[m      %s\n' "$(project)"
	$(V)GOOS=$(os) GOARCH=$(arch) GOWORK=off $(GO) test -tags release,osusergo,netgo -ldflags "-s -X $(module)/internal.ProgramVersion=$(version)" ./... >$(testout) || \
		(grep -v -e "^\?" -e "^ok" -e "^FAIL" $(testout) && exit 2)
	@printf '\e[38;5;202m  INFO\e[m      %s\n' "$(project)"
	@:> $(project).info
	$(V)$(foreach p,$(programs), \
		$(GO) version -m $(p) >> $(project).info; \
	)
	@if test -f scripts/code-stats; then scripts/code-stats >> $(project).info; fi

scrub::
	@printf '\e[38;5;202m  SCRUB\e[m\n'
	$(V)git clean -d -f -x

test::
	@printf '\e[38;5;202m  TEST\e[m      %s\n' "$(project)"
	$(V)GOOS=$(os) GOARCH=$(arch) $(GO) test -tags osusergo,netgo -ldflags "-s -X $(module)/internal.ProgramVersion=$(version)" ./... >$(testout) || \
		(grep -v -e "^\?" -e "^ok" -e "^FAIL" $(testout) && exit 2)

version::
	@printf '\e[38;5;202m  GENERATED VERSION\e[m      %s\n' "$(version)"

.PHONY: all clean dep-upgrade devel install lint release scrub test version
