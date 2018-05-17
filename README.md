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

# About Benchmark


Machine Detail:

- 16G
- 8C

Speed:

``` shell
dd if=/dev/zero of=/tmp/output bs=8k count=10k; rm -f /tmp/output
10240+0 records in
10240+0 records out
83886080 bytes (84 MB, 80 MiB) copied, 0.026572 s, 3.2 GB/s
```

## encoder benchmark

``` shell
timestamp: 1526523292
go test -v -bench=. ./colly -run=BenchmarkFileContent
goos: linux
goarch: amd64
pkg: github.com/smileboywtu/FileColly/colly
BenchmarkFileContentEncoder_Encode10-8   	   10000	    204298 ns/op
BenchmarkFileContentEncoder_Encode20-8   	   10000	    197779 ns/op
BenchmarkFileContentEncoder_Encode30-8   	   10000	    205312 ns/op
BenchmarkFileContentEncoder_Encode40-8   	   10000	    204323 ns/op
BenchmarkFileContentEncoder_Encode50-8   	    5000	    203801 ns/op
PASS
ok  	github.com/smileboywtu/FileColly/colly	9.252s
```

## collector benchmark

if you want to run test on your machine, first `run build_testdata.sh` inside `hack` directory

``` shell
go test -v -bench=. ./colly -run=BenchmarkCollector_Start
goos: linux
goarch: amd64
pkg: github.com/smileboywtu/FileColly/colly
BenchmarkCollector_Start-8               	       1	239480616423 ns/op
```

## summary

file count: 500000  2092 f/s

filesize: 20 Bytes