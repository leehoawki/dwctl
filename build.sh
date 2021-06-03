export CGO_ENABLED=0
export GOARCH=amd64
export GOOS=linux
go build dwctl.go
chmod +x dwctl