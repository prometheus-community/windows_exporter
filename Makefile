export GOOS=windows

build:
	promu build -v

test:
	go test -v ./...

lint:
	gometalinter --vendor --config gometalinter.config ./...

fmt:
	gofmt -l -w -s .

crossbuild:
	# The prometheus/golang-builder image for promu crossbuild doesn't exist
	# on Windows, so for now, we'll just build twice
	GOARCH=amd64 promu build --prefix=output/amd64
	GOARCH=386   promu build --prefix=output/386
