package text
import (
	"fmt"
	//"strings"
	//"strconv"
	"centi/stegano/util"
)


/*
 * The same basic logic as in EncodeWithTypos, but hides zeros and ones with
 * another characters. Also has some modes.
 */
func EncodeWithUnprintable( unprintable0, unprintable1 rune, mode uint8, data []byte, s string ) (string, error) {
	
	encodedTmp, err := util.EncodeToBinary( data )
	if err != nil {
		return "", err
	}

	// todo: handle the case where unprintable0 = 1 and unprintable1 = 0
	encoded := ""
	for _, e := range encodedTmp {
		if e == 0 {
			encoded += string( unprintable0 )
		} else {
			encoded += string( unprintable1 )
		}
	}

	switch mode {
	case PrefixMode:
		return string(encoded) + s, nil
	case SuffixMode:
		return s + string(encoded), nil
	case EmbedMode:
		// the most complex one...
		result := ""
		eIdx := 0
		for _, l := range s {

			if eIdx < len(encoded) {
				result += string( encoded[eIdx] )
				eIdx++
			}
			result += string( l )
		}
		if eIdx < len(encoded) {
			result += string( encoded[eIdx:] )
		}
		return result, nil
	default:
		return "", fmt.Errorf("Invalid encoding mode")
	}
}

func DecodeFromUnprintable( unprintable0, unprintable1 rune, s string ) ([]byte, error) {
	result := []uint8{}
	for _, run := range s {
		if run == unprintable0 {
			result = append( result, 0 )
		} else if run == unprintable1 {
			result = append( result, 1 )
		}
	}
	decoded, err := util.DecodeFromBinary( result )
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
