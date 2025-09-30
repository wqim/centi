package util
import (
	"os"
	"os/exec"
	"errors"
	"crypto/rand"
	"golang.org/x/text/unicode/norm"
)

const (
	ShredingCount = 10
)
func ReplaceBytes( src []uint8, from uint8, to uint8 ) []uint8 {
	for i, b := range src {
		if b == from {
			src[i] = to
		}
	}
	return src
}

func Abs( x int ) int {
	if x < 0 {
		return -x
	}
	return x
}

func FixUnicode( in string ) string {
	return norm.NFC.String( in )
}

func ShredFile( filename string ) error {
	
	fileInfo, err := os.Stat( filename )
	if err != nil {
		return err
	}

	buf := make( []byte, fileInfo.Size() )

	for i := 0; i < ShredingCount; i++ {

		// just generate random bytes and write them as file content.
		// todo: optimize this function for working with large files.
		if _, err := rand.Read( buf ); err != nil {
			return err
		}
		if err = os.WriteFile( filename, buf, 0660 ); err != nil {
			return err
		}
	}
	return nil
}


func CreateTempfile( data []byte ) (string, error) {
	f, err := os.CreateTemp( "", "tmpfile-" )
	if err != nil {
		return "", err
	}
	defer f.Close()
	if data != nil {
		if _, err := f.Write(data); err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}

func PathToProgram( prog string ) (string, error) {
	path, err := exec.LookPath( prog )
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	return path, err
}
