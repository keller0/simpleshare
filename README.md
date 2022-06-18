# simpleshare

## Share texts and files in local network

 **⚠️ THERE IS NO SECURITY ASSURANCE, PLEASE USE WITH CAUTION.**

### Build and run

```
go build -mod vendor

./simpleshare -h

  A Simple http service for share texts and files in local network
built in Go.
  source: https://github.com/keller0/simpleshare

Usage:
  simpleshare [flags]

Examples:
./simpleshare  -a 127.0.0.1 -p 7777 -f tmpFile

Flags:
  -a, --address string   listen address (default "127.0.0.1")
  -f, --folder string    tmp file folder (default "tmpFile")
  -h, --help             help for simpleshare
  -p, --port string      listen port (default "7777")
  -v, --version          version for simpleshare
```

### Web UI

![ui](https://user-images.githubusercontent.com/19371712/170302601-6d591f06-1ae8-4578-8886-8059401b5715.png)

### Demonstration

https://user-images.githubusercontent.com/19371712/165465190-d917bf54-8dc1-41de-b855-53e834561494.mp4
