package audio
import (
	"os"
	"bytes"
	"testing"
)

func TestFlac( t *testing.T ) {
	files := []string{
		"../tests/test1.flac",
		"../tests/test2.flac",
	}
	tests := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello world"),
		bytes.Repeat([]byte("a"), 4096),
	}

	for _, filename := range files {
		for _, data := range tests {
			content, _ := os.ReadFile( filename )
			enc, err := HideInFlac( content, data )
			if err != nil {
				t.Fatalf("Failed to hide data in flac file %s: %s", filename, err.Error())
			} else {
				os.WriteFile("../tests/test-encoded.flac", enc, 0660)
				dec, err := RevealFromFlac( enc )
				if err != nil {
					t.Fatalf("Failed to reveal data from flac file: %s", err.Error())
				} else if bytes.Equal(dec, data) == false {
					if len(dec) > 32 {
						dec = dec[:32]
					}
					if len(data) > 32 {
						data = data[:32]
					}
					t.Fatalf("Steganography spoiled data: %v != %v", data, dec)
				}
			}
		}
	}
}
