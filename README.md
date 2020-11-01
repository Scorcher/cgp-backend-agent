# Test run

```bash
go run cgpbackendagent -dataDir test-dir
```

# Build

for linux x86_64
```bash
GOOS=linux GOARCH=amd64 go build cgpbackendagent
```

# API

## getfile

```bash
curl -s localhost:8041/getfile/test1.intranet.ru/test.ru/test@test.com/account.dst
```