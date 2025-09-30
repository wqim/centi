package text
import (
	"fmt"
	"strings"
	"centi/stegano/util"
)


/*
 * encodes data using typos undetectable to human eye.
 * for example, replaces 'a' with russian 'a' in certain conditions, or, maybe, even replaces the letter from one encoding to another.
 *
 * [in]:
 *	alph1 - alphabet from letters located in the 's' parameter (for example, latin letters).
 *	alph2 - alphabet from letters to replace source ones.
 *	data - data to encode.
 *	s - decoy message.
 * [out]:
 * 	out - encoded message
 * 	err - error if size of s is too small to encode data in it.
 *
 */
func EncodeWithTypos( alph1, alph2 string, data []byte, s string ) (string, error) {
	totalLetters := 0
	for _, l := range s {
		if strings.ContainsRune( alph1, l ) {
			totalLetters += 1
		}
	}
	binaryData, err := util.EncodeToBinary( data )
	if err != nil {
		return "", err
	}

	if totalLetters < len(binaryData) {
		return "", fmt.Errorf("String is too short to encode the data.")
	}


	idx := 0 // index of bit in data
	out := []rune{}
	//runes := []rune(s)
	//a1 := []rune( alph1 )
	a2 := []rune( alph2 )

	for _, run := range s {
		// replace letter from alph1 to letter from alph2
		// in case if binaryData[idx] == 1
		if strings.ContainsRune( alph1, run ) {
			if idx < len(binaryData) {
				if binaryData[idx] == 1 {
					out = append( out, a2[ strings.IndexRune(alph1, run )  ] )
				} else {
					out = append( out, run )
				}
			}
			idx++
		} else {
			out = append( out, run )
		}
	}
	return string(out), nil
}

// inverse function
func DecodeFromTypos( alph1, alph2, s string ) ([]byte, error) {

	tmpresult := []uint8{}
	for _, l := range s {
		if strings.ContainsRune( alph1, l ) {
			tmpresult = append( tmpresult, uint8(0) )
		} else if strings.ContainsRune( alph2, l ) {
			tmpresult = append( tmpresult, uint8(1) )
		}
	}
	decoded, err := util.DecodeFromBinary( tmpresult )
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
