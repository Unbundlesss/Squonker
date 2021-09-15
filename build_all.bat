@echo off
go get github.com/mitchellh/gox
 
echo building all versions ...
gox -osarch="windows/amd64 darwin/amd64 linux/amd64" -output="_builds/{{.OS}}/{{.Dir}}"

pause