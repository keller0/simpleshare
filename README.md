# simpleshare

## Share texts and files in local network

 **⚠️ THERE IS NO SECURITY ASSURANCE, PLEASE USE WITH CAUTION.**

### Build and run

```shell
go build -mod vendor

./simpleshare -h

  A Simple http service for share texts and files in local network
built in Go.
  source:0 https://github.com/keller0/simpleshare

Examples:
./simpleshare  -a 127.0.0.1 -p 7777 -f tmpFile

Flags:
  -a, --address string   listen address (default "127.0.0.1")
  -f, --folder string    tmp file folder (default "tmpFile")
  -h, --help             help for simpleshare
  -p, --port string      listen port (default "7777")

Use "simpleshare [command] --help" for more information about a command.
```

### Demonstration

https://user-images.githubusercontent.com/19371712/165465190-d917bf54-8dc1-41de-b855-53e834561494.mp4
