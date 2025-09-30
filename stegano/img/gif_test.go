package img
import (
	"os"
	"bytes"
	"testing"
)

func TestGIF( t *testing.T ) {
	files := []string{
		"../tests/two-time-forsaken.gif",
	}
	tests := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello World!"),
		bytes.Repeat([]byte("a"), 4096),
		bytes.Repeat([]byte("A"), 10000),
	}

	for _, filename := range files {
		for _, data := range tests {
			img, err := os.ReadFile( filename )
			if err != nil {
				t.Fatalf("Failed to read file %s: %s", filename, err.Error())
			} else {
				// run tests...
				enc, err := HideInGif( img, data )
				if err != nil {
					t.Errorf("Failed to encode data in jpeg: %s", err.Error())
				} else {
					//os.WriteFile( "../tests/test_encoded.gif", enc, 0600 )
					dec, err := RevealFromGif( enc )
					if err != nil {
						t.Errorf("Failed to extract data from gif: %s", err.Error() )
					} else if bytes.Equal( data, dec ) == false {
						t.Errorf("GIF steganography spoiled data: %v != %v", data, dec)
					}
				}
			}
		}
	}
}
