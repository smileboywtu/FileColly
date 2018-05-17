package colly

import "testing"

func BenchmarkFileContentEncoder_Encode(b *testing.B) {

	encoder := FileContentEncoder{
		FilePath: "/tmp/a.txt",
		FileContent:[]byte{'a', 'b', 'c'},
	}
	for i:=0 ; i<b.N; i++ {
		encoder.Encode()
	}
}