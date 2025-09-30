package document
import (
	"os"
	"log"
	"bytes"
	"strconv"
	"testing"
)

func TestPdf(t *testing.T) {
	files := []string{
		"../tests/document.pdf",
		"../tests/test2.pdf",
		"../tests/test.pdf",
	}

	tests := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello worlf"),
		bytes.Repeat( []byte("a"), 512 ),
	}

	modes := []uint8{ AfterEOF, CRNL, OperatorMode }

	for _, filename := range files {
		
		pdf, err := os.ReadFile( filename )
		if err != nil {
			t.Fatalf("Failed to read file %s: %s", filename, err.Error())
			continue
		}

		for _, mode := range modes {
			for _, data := range tests {
				log.Printf("Embedding data(%d bytes) with mode %d in %s...\n", len(data), int(mode), filename)
				newPdf, err := HideInPdf( mode, pdf, data )
				if err != nil {
					t.Fatalf("Failed to hide data in pdf (mode %d): %s", mode, err.Error())
				} else {
					os.WriteFile( "../tests/test-encoded-" + strconv.Itoa( int(mode) ) + ".pdf", newPdf, 0600 )
					decoded, err := RevealFromPdf( mode, newPdf )
					if err != nil {
						t.Errorf("Failed to reveal data from pdf (%d): %s", mode, err.Error())
					} else if bytes.Equal( decoded, data ) == false {
						t.Errorf("Mode %d sploiled the data:\n%v != %v", mode, data, decoded)
					}
				}
			}
		}
	}
}
