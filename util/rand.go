package util
import (
	"math/big"
	"strconv"
	"encoding/base64"
	"crypto/rand"
	"centi/cryptography"
)

const (
	IDLength = 32
)

var (
	lastIDFailed = 0
)

func GenFilename( prefix string, ext string ) string {
	filename := prefix + strconv.Itoa( RandInt(100000) ) + "." + ext
	return filename
}

func RandInt( max int ) int {
	limit := big.NewInt( int64(max) )
	integer, err := rand.Int( rand.Reader, limit )
	if err != nil {
		return 0
	}
	return int(integer.Int64())
}

func GenID() string {
	//buffer := make([]byte, length)
	buffer, err := cryptography.GenRandom( uint(IDLength) )
	if err != nil {
		lastIDFailed++
		return "gen-id-failed-" + strconv.Itoa( lastIDFailed )
	}
	return base64.StdEncoding.EncodeToString( buffer )
}
