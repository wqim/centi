package img
import (
	"os"
	"bytes"
	"testing"
)

func TestPNG( t *testing.T ) {
	images := []string{
		"../tests/test.png",
		"../tests/Untitled.png",
	}

	tests := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello world!"),
		bytes.Repeat([]byte("a"), 4096),
		bytes.Repeat([]byte("A"), 10000),
	}

	modes := []uint8{
		RMode,
		GMode,
		BMode,
		RMode | GMode,
		RMode | BMode,
		GMode | BMode,
		RMode | GMode | BMode,
	}

	for _, data := range tests {
		for _, filename := range images {
			img, _ := os.ReadFile( filename )
			for _, mode := range modes {
				enc, err := EncodeWithLSB( mode, data, img )
				if err != nil {
					t.Errorf("Failed to encode data: %v", err)
				} else {
					//os.WriteFile( "test.png", enc, 0660 )
					dec, err := DecodeFromLSB( mode, enc )
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
}
