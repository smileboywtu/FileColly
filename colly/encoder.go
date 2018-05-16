// Compress and pack data
package colly

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"github.com/vmihailenco/msgpack"
)

type FileContentEncoder struct {
	FilePath    string
	FileContent []byte
}

type EncodeResult struct {
	Path          string
	EncodeContent string
	Err           error
}

// Encode encode data in base64 format and
// compress use msgpack
func (c *FileContentEncoder) Encode() (string, error) {

	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, 6)

	if err != nil {
		return "", err
	}
	writer.Write(c.FileContent)
	writer.Close()

	b64path := base64.StdEncoding.EncodeToString([]byte(c.FilePath))

	ctx := map[string]string{}
	ctx["path"] = b64path
	ctx["content"] = buf.String()

	packBytes, err := msgpack.Marshal(ctx)
	if err != nil {
		return "", err
	}

	return string(packBytes[:]), nil
}
