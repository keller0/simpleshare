# simpleshare

## Share texts and files in local network

 **⚠️ THERE IS NO SECURITY ASSURANCE, PLEASE USE WITH CAUTION.**

### Web UI

![ui](https://github.com/user-attachments/assets/d5c38118-e255-4d5d-a8eb-8dcd6bd3ead0)

### Build and run

```
go build

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
