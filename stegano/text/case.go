package text
import (
	"fmt"
	"strings"
	"centi/stegano/util"
)

// encodes data in the case of letters.
// if GreedyMode is used, encodes data in the start of sentences in the text
// if not, changes the case of every word's first letter.
// todo: add support for other alphabets (russian, french, etc.)
// todo: add mode in which all the letters are encoded.
func EncodeWithCase( mode uint8, data []byte, s string ) (string, error) {
	
	delimeter := getDelimeter( mode )
	
	//fmt.Printf("delimeter: '%s'\n", delimeter)

	parts := strings.Split(s, delimeter )
	encoded, err := util.EncodeToBinary( data )
	if err != nil {
		return "", err
	}

	if len(encoded) > len(parts) {
		return "", fmt.Errorf("Unable to encode: decoy message is too small ")
	}

	to1Case := strings.ToUpper
	to0Case := strings.ToLower
	if mode & InverseMode == InverseMode {
		to1Case = strings.ToLower
		to0Case = strings.ToUpper
	}
	idx := 0
	for i, _ := range parts{
		if idx < len( encoded ) {
			if strings.Contains( IgnoredChars, string( []rune(parts[i])[0] ) ) == false {	// handle non-letter characters.
				if encoded[idx] == 1 {
					parts[i] = to1Case( string( []rune(parts[i])[0]) ) + parts[i][1:]
				} else {
					parts[i] = to0Case( string( []rune(parts[i])[0]) ) + parts[i][1:]
				}
				idx++
			}
		}
	}
	return strings.Join( parts, delimeter ), nil
}

// inverse function
func DecodeFromCase( mode uint8, s string ) ([]byte, error) {
	delimeter := getDelimeter( mode )
	parts := strings.Split( s, delimeter )
	encoded := []uint8{}

	for _, part := range parts {

		if strings.Contains( IgnoredChars, string( []rune(part)[0]) ) == false {
			if mode & InverseMode == InverseMode {
				// this is zero
				if strings.ToUpper( string( []rune(part)[0]) ) == string([]rune(part)[0]) {
					encoded = append( encoded, 0 )
				} else {
					// this is one
					encoded = append( encoded, 1 )
				}
			} else {
				// this is one
				if strings.ToUpper( string([]rune(part)[0]) ) == string([]rune(part)[0]) {
					encoded = append( encoded, 1 )
					//fmt.Println("uppercase word:", string(part))
				} else {
					// this is zero
					encoded = append( encoded, 0 )
				}
			}
		}
	}

	decoded, err := util.DecodeFromBinary( encoded )
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func getDelimeter( mode uint8 ) string {
	delimeter := " "
	if mode & GreedyMode == GreedyMode {
		delimeter = "."
	}
	return delimeter
}
