PKG = github.com/k1LoW/harvest
COMMIT = $$(git describe --tags --always)
OSNAME=${shell uname -s}
ifeq ($(OSNAME),Darwin)
	DATE = $$(gdate --utc '+%Y-%m-%d_%H:%M:%S')
else
	DATE = $$(date --utc '+%Y-%m-%d_%H:%M:%S')
endif

export GO111MODULE=on

BUILD_LDFLAGS = -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)

default: test

ci: depsdev test build integration sec

test:
	go test ./... -coverprofile=coverage.out -covermode=count

sec:
	gosec ./...

integration:
	@cat testdata/test.yml.template | sed -e "s|__PWD__|${PWD}|" > testdata/test.yml
	@./hrv fetch -c testdata/test.yml -o test.db --start-time='2019-01-01 00:00:00' -v
	test `./hrv cat test.db | grep -c ''` -gt 0 || exit 1
	@rm test.db

build:
	go build -ldflags="$(BUILD_LDFLAGS)" ./cmd/hrv

depsdev:
	go install github.com/Songmu/ghch/cmd/ghch@v0.10.2
	go install github.com/Songmu/gocredits/cmd/gocredits@v0.2.0
	go install github.com/securego/gosec/v2/cmd/gosec@v2.8.1

dbdoc: build
	@cat testdata/test.yml.template | sed -e "s|__PWD__|${PWD}|" > testdata/test.yml
	@./hrv fetch -c testdata/test.yml -o harvest.db --start-time='2019-01-01 00:00:00'
	@tbls doc -f
	@rm harvest.db

prerelease:
	ghch -w -N ${VER}
	git add CHANGELOG.md
	git commit -m'Bump up version number'
	git tag ${VER}

release:
	goreleaser --rm-dist

.PHONY: default test
