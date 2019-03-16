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

ci: test integration

test:
	go test ./... -coverprofile=coverage.txt -covermode=count

integration: build
	@cat testdata/test.yml.template | sed -e "s|__PWD__|${PWD}|" > testdata/test.yml
	@./hrv fetch -c testdata/test.yml -o test.db --start-time='2019-01-01 00:00:00'
	test `./hrv cat test.db | grep -c ''` -gt 0 || exit 1
	@rm test.db

build:
	go build -ldflags="$(BUILD_LDFLAGS)" ./cmd/hrv

depsdev:
	go get golang.org/x/tools/cmd/cover
	go get golang.org/x/lint/golint
	go get github.com/linyows/git-semv/cmd/git-semv
	go get github.com/Songmu/ghch/cmd/ghch

prerelease:
	ghch -w -N ${VER}
	git add CHANGELOG.md
	git commit -m'Bump up version number'
	git tag ${VER}

release:
	goreleaser

.PHONY: default test
