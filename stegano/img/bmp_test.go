package img
import (
	"os"
	"bytes"
	"testing"
)

func TestBMP( t *testing.T ) {
	images := []string{
		"../tests/test.bmp",
	}

	tests := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello world!"),
		bytes.Repeat([]byte("a"), 4096),
		bytes.Repeat([]byte("A"), 10000),
	}

	for _, data := range tests {
		for _, filename := range images {
			img, _ := os.ReadFile( filename )
			enc, err := HideInBMP( img, data )
			if err != nil {
				t.Errorf("Failed to encode data: %v", err)
			} else {
				//os.WriteFile("../tests/test_encoded.bmp", enc, 0600)
				dec, err := RevealFromBMP( enc )
				if err != nil {
					t.Errorf("Failed to extract data: %v", err)
				} else if bytes.Equal( data, dec ) == false {
					t.Errorf("Steganography spoiled the data. %v != %v",
						data, dec)
				}
			}
		}
	}
}
