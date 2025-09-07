include-dir := $(CURDIR)/mk

include $(include-dir)/all.mk

sources := $(addprefix $(CURDIR)/,$(wildcard *.go))

GOPATH ?= $(HOME)/go
GOBIN ?= $(GOPATH)/bin

MAKEFLAGS += -j
goenv := GOEXPERIMENT=greenteagc GOEXPERIMENT=jsonv2
go := $(goenv) go
build-dir := $(CURDIR)/build
local-clean :=

cli-dir := cli
cli-bin := fastmask

binary-tags := cli
phony-targets := clean all

tool-types := go

quality-checkers := govulncheck scc staticcheck errcheck

# `make deepcheck=t` to run expensive static analysis tools durin CI
$(if $(strip $(deepcheck)), \
  $(eval quality-checkers += deadcode nilaway) \
 )

go-tools := $(quality-checkers) gofumpt wgo
gofumpt-url := mvdan.cc/gofumpt@latest
nilaway-url := go.uber.org/nilaway/cmd/nilaway@latest
govulncheck-url := golang.org/x/vuln/cmd/govulncheck@latest
scc-url := github.com/boyter/scc/v3@latest
wgo-url := github.com/bokwoon95/wgo@latest
deadcode-url := golang.org/x/tools/cmd/deadcode@latest
staticcheck-url := honnef.co/go/tools/cmd/staticcheck@latest
errcheck-url := github.com/kisielk/errcheck@latest

nilaway-artifact := $(build-dir)/nilaway.out
nilaway-flags := ./...

govulncheck-artifact := $(build-dir)/govulncheck.out
govulncheck-flags :=

staticcheck-artifact := $(build-dir)/staticcheck.out
staticcheck-flags := ./...

errcheck-artifact := $(build-dir)/errcheck.out
errcheck-flags := ./...

scc-artifact := $(build-dir)/scc.out
scc-flags := \
  --no-cocomo \
  --sort code \
  -M json \
  -M css \
  -M gitignore \
  --exclude-dir deprecated \
  --exclude-ext xml \
  --exclude-ext toml \
  --exclude-ext md \

deadcode-artifact := $(build-dir)/deadcode.out
deadcode-flags := ./...


go-verbs := test vet
go-test-flags := -failfast -parallel=8 -count=2 -shuffle=on
go-vet-flags :=

go-mod-tidy-artifact := $(build-dir)/go-mod-tidy.out

.DEFAULT_GOAL := all

$(foreach .ph,$(phony-targets), \
  $(eval .PHONY: $(.ph)) \
  $(eval $(.ph):) \
 )

$(foreach .tool-type,$(tool-types), \
  $(foreach .tool,$($(.tool-type)-tools), \
    $(call install-$(.tool-type)-tool,$(.tool)) \
   ) \
 )

all \
: tools $(scc-artifact) $(go-mod-tidy-artifact) \
; @cat $(scc-artifact)

$(foreach tag,$(binary-tags), \
  $(eval .target := $(build-dir)/$($(tag)-bin)) \
  $(eval .source-dir := $(CURDIR)/$($(tag)-dir)) \
  $(eval .sources := $(shell find $(.source-dir) -type f -name "*.go")) \
  $(eval sources += $(.sources)) \
  $(eval \
    $(.target) \
    : $(.sources) $(MAKEFILE_LIST) \
    | $(build-dir) $(go-mod-tidy-artifact) \
    ; @echo BUILDING `basename $(.target)` \
    ; $(go) generate $(.source-dir)/... \
    ; $(gofumpt) -l -w . \
    ; $(go) build -o $(.target) $(.sources) \
   ) \
  $(eval all: $(.target)) \
 )

$(go-mod-tidy-artifact) \
: $(sources) $(MAKEFILE_LIST) \
; @echo RUNNING tidy \
; $(go) mod tidy > $@ 2>&1 \
; exit_code=$$? \
; cat $@ \
; exit $$exit_code

$(foreach .verb,$(go-verbs), \
  $(eval .artifact-name := go-$(.verb)-artifact) \
  $(eval $(.artifact-name) := $(build-dir)/go-$(.verb).out) \
  $(eval .artifact := $($(.artifact-name))) \
  $(eval \
    $(.artifact) \
    : $(sources) $(MAKEFILE_LIST) \
    | $(build-dir) \
    ; @echo RUNNING $(.verb) \
    ; $(go) $(.verb) $(go-$(.verb)-flags) ./... > $(.artifact) 2>&1 \
    ; exit_code=$$$$? \
    ; cat $(.artifact) \
    ; exit $$$$exit_code \
   ) \
  $(eval all: $(.artifact)) \
 )

$(foreach .qc,$(quality-checkers), \
  $(eval .qc-bin := $($(.qc))) \
  $(eval .qc-flags-variable := $(.qc)-flags) \
  $(eval .qc-flags := $($(.qc-flags-variable))) \
  $(eval .artifact-variable := $(.qc)-artifact) \
  $(eval .artifact := $($(.artifact-variable))) \
  $(call assert,$(.artifact),missing $(.artifact-variable)) \
  $(eval \
    $(.artifact) \
    : $(.qc-bin) $(sources) $(MAKEFILE_LIST) \
    | $(build-dir) \
    ; @echo RUNNING $(.qc) \
    ; $(.qc-bin) \
      $(.qc-flags) \
      > $(.artifact) \
      2>&1 \
    ; exit_code=$$$$? \
    ; test $$$$? -ne 0 && cat $(.artifact) \
    ; exit $$$$exit_code \
   ) \
  $(eval all: $(.artifact)) \
 )

$(if $(strip $(build-dir)), \
  $(if $(call neq,/,$(build-dir)), \
    $(eval clean: clean-build-dir) \
    $(eval .PHONY: clean-build-dir) \
    $(eval clean-build-dir:; rm -Rf $(build-dir)) \
   ) \
 )

$(build-dir):; mkdir -p $@
