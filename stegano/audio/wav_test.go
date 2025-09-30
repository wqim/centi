package audio
import (
	"os"
	"fmt"
	"bytes"
	"testing"
)

func TestWAV( t *testing.T ) {
	files := []string{
		"../tests/cave_theme_1.wav",
	}
	tests := [][]byte{
		nil,
		[]byte{},
		[]byte(`
HELLO WORLD
HALOWEEN
BINARY WORLD: a lot of strange digits.
NON-BINARY WORLD: a lot of strange people.
		`),
		bytes.Repeat( []byte("a"), 1024 ),
		bytes.Repeat( []byte("A"), 4096 ),
	}

	for i, data := range tests {
		for j, filename := range files {
			fmt.Println("START", i, j)
			wv, err := os.ReadFile( filename )
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			} else {
				encoded, err := HideInWav( data, wv )
				fmt.Println("Encoded")
				if err != nil {
					t.Errorf("Failed to encode data in wav file: %s", err.Error())
				} else {
					decoded, err := RevealFromWav( encoded )
					fmt.Println("Decoded")
					if err != nil {
						t.Errorf("Failed to decode wav file: %s", err.Error())
					} else if bytes.Equal(decoded, data) == false{
						t.Errorf("Steganography method spoiled the data. %v != %v",
						data, decoded )
					}
				}
			}
		}
	}
}
