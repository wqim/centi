package text
import (
)

const (
	ZeroWidthJoiner = '\u200d'	// zero-width joiner
	ZeroWidthNonJoiner = '\u200c'	// zero-width non-joiner
)

func Hide( decoy, data []byte ) ([]byte, error) {
	str, err := EncodeWithUnprintable( ZeroWidthJoiner, ZeroWidthNonJoiner, EmbedMode, data, string(decoy) )
	return []byte(str), err
}


func Reveal( decoy []byte ) ([]byte, error) {
	return DecodeFromUnprintable( ZeroWidthJoiner, ZeroWidthNonJoiner, string(decoy) )
}
