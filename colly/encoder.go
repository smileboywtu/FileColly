// Compress and pack data
package colly

import (
	"bytes"
	"compress/zlib"
	"github.com/vmihailenco/msgpack"
	"encoding/base64"
)

type FileEncoder struct {
	FilePath    string
	FileContent []byte
}

func (c *FileEncoder) Encode() (string, error) {

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
