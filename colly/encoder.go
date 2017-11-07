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

// Encode files use zlib and msgpack lib tool
func (c *FileEncoder) Encode() (string, error) {

	var buf bytes.Buffer
	// level same to python default
	writer, err := zlib.NewWriterLevel(&buf, 6)
	defer func() {
		writer.Close()
	}()

	if err != nil {
		return "", err
	}
	writer.Write(c.FileContent)

	b64path := base64.StdEncoding.EncodeToString([]byte(c.FilePath))

	// file context
	ctx := map[string]string{}
	ctx["path"] = b64path
	ctx["content"] = buf.String()

	packBytes, err := msgpack.Marshal(ctx)
	if err != nil {
		return "", err
	}

	return string(packBytes[:]), nil
}
