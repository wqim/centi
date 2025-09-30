package archive
import (
	"os"
	"bytes"
	"testing"
)

func TestParsing( t *testing.T ) {
	filename := "../tests/test.zip"
	data, err := os.ReadFile( filename )
	if err != nil {
		t.Fatalf("Unable to run tests because test zip file not found: %s", err.Error())
		return
	}

	// parse zip file
	zip, err := FromBytes( data )
	if err != nil {
		t.Errorf("Failed to parse zip file: %s", err.Error())
	} else {
		data := []byte("Hello world!!!!")
		if err := zip.Embed( data ); err != nil {
			t.Errorf("Failed to embed data in zip file: %s", err.Error())
		} else {
			dumped, err := zip.Bytes()
			if err != nil {
				t.Errorf("Failed to turn zip file into bytes: %s", err.Error())
			} else {
				//os.WriteFile( "tmp.zip", dumped, 0600 )
				zip2, err := FromBytes( dumped )
				if err != nil {
					t.Errorf("Failed to parse zip file (2): %s", err.Error())
				} else {
					extracted, err := zip2.Extract()
					if err != nil {
						t.Errorf("Failed to extract data from new zip: %s", err.Error())
					} else if bytes.Equal( extracted, data ) == false {
						t.Errorf("Data was changed during steganography. Orig: %v; Extracted: %v",
							data, extracted)
					}
				}
			}
		}
	}
}
