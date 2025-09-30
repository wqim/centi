package util
import (
	"fmt"
	"encoding/binary"
)

/*
 * transform data from/to binary form
 */
func ToBin( x byte ) []byte {
	result := []byte{}
	for {
		result = append( result, byte(x % 2) )
		x /= 2
		if x == 0 {
			break
		}
	}
	for {
		if len(result) == 8 {
			return result
		}
		result = append( result, 0 )
	}
}

func FromBin( x []byte ) byte {
	result := byte(0)
	for i := 0; i < 8; i++ {
		result *= 2
		result += x[ 8 - i - 1 ]
	}
	return result
}


func EncodeToBinary( data []byte ) ([]byte, error) {
	result := make([]uint8, 8)
	binary.LittleEndian.PutUint64( result, uint64(len(data)) )

	//fmt.Println("[debug 1]: length = ", len(data) )
	for _, b := range data {
		result = append( result, b )
	}

	res := []byte{}
	for _, b := range result {
		res = append( res, ToBin( b )... )
	}

	if len(res) < 100 {
		//fmt.Println( "[encodeWIthBinary]: ", res )
	}
	return res, nil
}

func DecodeFromBinary( data []uint8 ) ([]byte, error) {
	result := []byte{}
	if len(data) < 100 {
		//fmt.Println("[decodeFromBinary]: ", data)
	}
	for i := 0; i < len(data); i += 8 {
		if len(data) >= i + 8 {
			result = append( result, FromBin( data[i:i+8] ) )
		}
	}

	//fmt.Println("[decodeFromBinary] result = ", result[:8])

	if len(result) < 8 {
		return nil, fmt.Errorf("There is no encoded data")
	}
	var length uint64
	//fmt.Println( "[decode]: ", result[:16] )
	length = binary.LittleEndian.Uint64( result[:8] )
	//fmt.Println("[debug 2]: length = ", length)
	//fmt.Println("[debug 3]: length(result) = ", len(result))
	if uint64(len(result)) < uint64(8 + length) {
		return nil, fmt.Errorf("Invalid data length")
	}
	return result[8:8+length], nil
}
