export CGO_ENABLED=0
export GOARCH=amd64
export GOOS=linux
go build dwctl.go
chmod +x dwctl
scp dwctl root@10.141.62.83:/usr/local/bin/
scp dwctl root@10.141.48.29:/usr/local/bin/