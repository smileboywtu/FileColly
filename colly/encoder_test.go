// Benchmark test for file content encoding

//  timestamp: 1526523292
//	go test -v -bench=. ./colly -run=BenchmarkFileContent
//	goos: linux
//	goarch: amd64
//	pkg: github.com/smileboywtu/FileColly/colly
//	BenchmarkFileContentEncoder_Encode10-8   	   10000	    204298 ns/op
//	BenchmarkFileContentEncoder_Encode20-8   	   10000	    197779 ns/op
//	BenchmarkFileContentEncoder_Encode30-8   	   10000	    205312 ns/op
//	BenchmarkFileContentEncoder_Encode40-8   	   10000	    204323 ns/op
//	BenchmarkFileContentEncoder_Encode50-8   	    5000	    203801 ns/op
//	PASS
//	ok  	github.com/smileboywtu/FileColly/colly	9.252s


package colly

import "testing"
import "github.com/icrowley/fake"

func testFileContentEncoder_Encode(data []byte, b *testing.B) {

	encoder := FileContentEncoder{
		FilePath:    "/tmp/a.txt",
		FileContent: data,
	}
	for i := 0; i < b.N; i++ {
		encoder.Encode()
	}
}

func BenchmarkFileContentEncoder_Encode10(b *testing.B) {
	testFileContentEncoder_Encode([]byte(fake.CharactersN(10)), b)
}

func BenchmarkFileContentEncoder_Encode20(b *testing.B) {
	testFileContentEncoder_Encode([]byte(fake.CharactersN(20)), b)
}

func BenchmarkFileContentEncoder_Encode30(b *testing.B) {
	testFileContentEncoder_Encode([]byte(fake.CharactersN(30)), b)
}
func BenchmarkFileContentEncoder_Encode40(b *testing.B) {
	testFileContentEncoder_Encode([]byte(fake.CharactersN(40)), b)
}
func BenchmarkFileContentEncoder_Encode50(b *testing.B) {
	testFileContentEncoder_Encode([]byte(fake.CharactersN(50)), b)
}
