package util
import (
	"strings"
	/*"strconv"
	"crypto/sha512"
	"centi/cryptography" */
)

func PrepareFilename( filename string ) string {
	parts := strings.Split( filename, "/" )
	if len(parts) == 1 {
		parts = strings.Split( filename, "\\" )
	}
	part := parts[ len(parts) - 1 ]
	parts = strings.Split( part, "." )
	if len(parts) == 2 {
		return GenFilename( parts[0], parts[1] )
	}
	return part
}

func MapContains( mp map[string]string, key string ) bool {
	for k, _ := range mp {
		if k == key {
			return true
		}
	}
	return false
}

func CalculateDataSize( data []byte, packetSize uint ) uint {
	// calculate the correct size of data which can be packed...
	/*plaintextSize := packetSize - cryptography.TagSize - cryptography.NonceSize
	length := strconv.Itoa( len(data) )
	encoded := cryptography.EncodeData( data )
	hashSize := sha512.Size

	// todo: calculate this value based on what we have
	jsonedDelta := 716	// with slight overflow
	return packetSize - plaintextSize - length - len(encoded) - hashSize - jsonedDelta */
	return packetSize / 2
}
