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
