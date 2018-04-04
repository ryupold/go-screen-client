# go-screen-client
client app to stream mjpeg streams and show them in the browser on a remote host

## build
if you want to compile the appliocation by your own you must:
1. install [go compiler](https://golang.org/)
2. run
```bash
go generate
```
  - on windows
```
go build -ldflags -H=windowsgui
```
  - on macOS/linux
```
go build 
```