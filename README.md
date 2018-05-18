# FileColly

![](https://github.com/smileboywtu/FileColly/blob/master/screen/arch.png)

collect local file and send compress content to redis

# Practice

- msgpack support
- compress binary
- high performance
- configurable by yaml
- command line flags
- cache file timeout support
- collect file timeout support
- queue limit support and checks
 
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

# Internal

you should adjust the `read wait time` config to avoid uncomplete files.
the collector will check if destination queue size true turns as need

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
BenchmarkCollector_Start100-8            	       1	2504965409 ns/op
BenchmarkCollector_Start300-8            	       1	3111993207 ns/op
BenchmarkCollector_Start500-8            	       1	3622320302 ns/op
BenchmarkCollector_Start1000-8           	       1	3786183160 ns/op
```

you should use benchmark test suit to find the suitable worker numbers.

## summary

filesize: 1024 Bytes

file count: 10000  3300 f/s

