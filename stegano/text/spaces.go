package text
import (
	"fmt"
	"strings"
	
	"centi/stegano/util"
)

/*
 * mostly suitable for hiding data inside code snippets.
 * the logic in straight mode is:
 * ' ' * 4 -> 0
 * '\t' -> 1
 * and in inverse mode is:
 * ' ' * 4 -> 1
 * '\t' -> 0
 * And it must work for almost all the languages (except the Python, of course)
 */

func EncodeWithSpaces( mode uint8, data []byte, s string ) (string, error) {

	zeroString, oneString := determineStrings( mode )

	parts := strings.Split( s, "\n" )
	totalSize := 0

	for i, part := range parts {
		totalSize += strings.Count( part, zeroString ) + strings.Count( part, oneString )
		parts[i] = strings.ReplaceAll( parts[i], zeroString, "\n" )
		parts[i] = strings.ReplaceAll( parts[i], oneString, "\n" )
	}
	if totalSize < len(data) * 8 {
		return "", fmt.Errorf("unable to encode")
	}
	
	enc, err := util.EncodeToBinary( data )
	if err != nil {
		return "", err
	}

	bitIndex := 0

	//fmt.Println("Encode from spaces")
	for idx, part := range parts {

		finalPart := part
		for {
			if strings.Contains( finalPart, "\n" ) == false {
				break
			}
			if bitIndex < len(enc) {
				if enc[bitIndex] == 1 {
					finalPart = strings.Replace( finalPart, "\n", oneString, 1 )
				} else {
					finalPart = strings.Replace( finalPart, "\n",  zeroString, 1 )
				}
				bitIndex++
			} else {
				finalPart = strings.ReplaceAll( finalPart, "\n", zeroString )
			}
		}
		parts[idx] = finalPart
	}
	return strings.Join( parts, "\n" ), nil
}

func DecodeFromSpaces( mode uint8, s string ) ([]byte, error) {
	
	zeroString, oneString := determineStrings( mode )
	
	parts := strings.Split( s, "\n" )
	encoded := []byte{}

	//fmt.Println("DecodeFromSpaces")
	for _, part := range parts {
		idx := 0

		for {
			if idx >= len(part) {
				break
			}
			if strings.HasPrefix( string(part[idx:]), zeroString ) {
				idx += len( zeroString )
				encoded = append( encoded, 0 )
			} else if strings.HasPrefix( string(part[idx:]), oneString ) {
				idx += len( oneString )
				encoded = append( encoded, 1 )
			} else {
				idx++
			}
		}
	}
	
	decoded, err := util.DecodeFromBinary( encoded )
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func determineStrings( mode uint8 ) (string, string) {
	zeroString := "    "
	oneString := "\t"

	if mode & InverseMode == InverseMode {
		zeroString = "\t"
		oneString = "    "
	}
	return zeroString, oneString
}
