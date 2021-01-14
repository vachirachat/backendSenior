package encryption

import (
	"bytes"
	"encoding/base64"
)

// B64Encode encode base 64
func B64Encode(data []byte) []byte {
	var result bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &result)
	b64.Write(data)
	b64.Close()

	return result.Bytes()
}

// B64Decode decode base64
func B64Decode(data []byte) ([]byte, error) {
	decodedSize := base64.StdEncoding.DecodedLen(len(data))
	dst := make([]byte, decodedSize)

	_, err := base64.StdEncoding.Decode(dst, data)
	return dst, err
}
