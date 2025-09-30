package protocol
import (
	"bytes"
	//"io/ioutil"
	"io"
	"compress/gzip"
)

func Compress( data []byte ) (uint8, []byte, error) {
	if data == nil || len(data) == 0 {
		return 0, data, nil
	}

	compressed, err := compress( data )
	if err != nil {
		return 0, nil, err
	}
	// check if we are able to decrease the total
	// size of data
	if len(compressed) >= len(data) {
		return 0, data, nil
	} else {
		return 1, compressed, nil
	}
}

func compress( data []byte ) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decompress( data []byte ) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return data, nil
	}
	buf := bytes.NewReader(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var out bytes.Buffer
	if _, err := io.Copy(&out, gz); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
