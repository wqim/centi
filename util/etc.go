package util
import (
	/*"strconv"
	"crypto/sha512"
	"centi/cryptography" */
)


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
