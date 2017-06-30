.PHONY: dev-setup

dev-setup:
	cp .script/hooks/* .git/hooks/
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/go-playground/overalls
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

.PHONY: deps

deps:
	dep ensure -v

.PHONY: test

test:
	overalls -project=github.com/htdvisser/squatt -ignore=".git,vendor" -- -v

.PHONY: lint-full

lint-full:
	gometalinter --deadline=300s --vendor --disable errcheck --disable gas ./...

.PHONY: cert

cert:
	go run $$(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost
