# FileColly

![](https://github.com/smileboywtu/FileColly/blob/master/screen/arch.png)

collect local file and send compress content to redis

# Practice

- Golang
- File pipeline
- readers worker pool
- sender worker pool
- msgpack support
- compress binary
- high performance
- configurable
- dep vendor
 
# How to start

1. clone repo:
``` shell
git clone https://github.com/smileboywtu/FileColly.git
```
2. install golang official `dep` and run inside project directory:
``` shell
dep ensure
```
3. build binary, then config:
``` shell
go build -ldflags "-w -s" main.go
```


  