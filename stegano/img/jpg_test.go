package img
import (
	"os"
	"bytes"
	"testing"
)

func TestJPEG( t *testing.T ) {
	files := []string{
		"../tests/1.jpg",
		"../tests/2.jpg",
		"../tests/3.jpg",
		"../tests/4.jpg",
		"../tests/5.jpg",
	}
	tests := [][]byte{
		//nil,
		//[]byte{},
		[]byte("Hello World!"),
		bytes.Repeat([]byte("a"), 4096),
	}

	for _, filename := range files {
		for _, data := range tests {
			img, err := os.ReadFile( filename )
			if err != nil {
				t.Fatalf("Failed to read file %s: %s", filename, err.Error())
			} else {
				// run tests...
				enc, err := HideInJpeg( img, data )
				if err != nil {
					t.Errorf("Failed to encode data in jpeg (%s): %s", filename, err.Error())
				} else {
					dec, err := RevealFromJpeg( enc )
					if err != nil {
						t.Errorf("Failed to extract data from jpeg: %s", err.Error() )
					} else if bytes.Equal( data, dec ) == false {
						t.Errorf("JPEG steganography spoiled data: %v != %v", data, dec)
					}
				}
			}
		}
	}
}
