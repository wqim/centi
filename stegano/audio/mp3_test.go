package audio
import (
	"os"
	"bytes"
	"strconv"
	"testing"
)

func TestMP3(t *testing.T) {
	files := []string{
		"../tests/test1.mp3",
		"../tests/test2.mp3",
	}
	tests := [][]byte{
		nil,
		[]byte{},
		[]byte("HELLO WORLD"),
		bytes.Repeat( []byte("a"), 4096 ),
		bytes.Repeat( []byte("A"), 4096 * 8 ),
	}

	description := "test-comment-section"

	for idx, filename := range files {
		for _, data := range tests {
			content, err := os.ReadFile( filename )
			if err == nil {
				enc, err := HideInMP3( description, content, data )
				if err != nil {
					t.Fatalf("Failed to hide data in mp3 file %s: %s", filename, err.Error())
				} else {
					os.WriteFile("../tests/test-encoded-" + strconv.Itoa( idx ) + ".mp3", enc, 0600 )
					dec, err := RevealFromMP3( description, enc )
					if err != nil {
						t.Fatalf("Failed to reveal hidden data: %s", err.Error())
					} else if bytes.Equal( dec, data ) == false {
						// fix not to trash out console...
						if len(dec) > 100 {
							dec = dec[:100]
						}
						if len(data) > 100 {
							data = data[:100]
						}
						t.Fatalf("Hidden != revealed: %v != %v", data, dec )
					}
				}
			}
		}
	}
}
