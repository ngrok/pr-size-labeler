VERSION := 65c7ebc607d90fd62527fe82a0659eba86061b3c

binaries:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags="-s -w" -trimpath -o ./bin/pr-size-labeler-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -buildvcs=false -ldflags="-s -w" -trimpath -o ./bin/pr-size-labeler-linux-arm64
	ls -alth ./bin

test:
	go test -v -race ./... -coverprofile cover.out
