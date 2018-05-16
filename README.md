# FileColly

![](https://github.com/smileboywtu/FileColly/blob/master/screen/arch.png)

collect local file and send compress content to redis

# Practice

- msgpack support
- compress binary
- high performance
- configurable by yaml
- command line flags
- cache file support
 
# How to start

1. clone repo:
``` shell
git clone https://github.com/smileboywtu/FileColly.git
```
2. install golang official `dep` and run inside project directory:
``` shell
go get -u github.com/golang/dep/cmd/dep

dep ensure
```
3. build binary, then config:
``` shell
./build.sh
```