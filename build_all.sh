
go get github.com/mitchellh/gox

export PATH=$PATH:$(go env GOPATH)/bin
gox -osarch="windows/amd64 darwin/amd64 linux/amd64" -output="_builds/{{.OS}}/{{.Dir}}"
