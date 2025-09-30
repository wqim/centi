package audio
import (
	"bytes"
)

func Hide( description string, decoy, data []byte ) ([]byte, error) {
	if bytes.Equal(decoy[:4], []byte("fLaC")) {
		// flac file
		return HideInFlac( decoy, data )
	} else if bytes.Equal(decoy[:4], []byte("RIFF")) {
		// for now, assume it's a wav file
		// todo: add support for other RIFF formats
		return HideInWav( data, decoy )
	} else {
		// mp3 thing...
		return HideInMP3( description, decoy, data )
	}
}

func Reveal( description string, decoy []byte ) ([]byte, error) {
	if bytes.Equal(decoy[:4], []byte("fLaC")) {
		// flac file
		return RevealFromFlac( decoy )
	} else if bytes.Equal(decoy[:4], []byte("RIFF")) {
		// for now, assume it's a wav file
		// todo: add support for other RIFF formats
		return RevealFromWav( decoy )
	} else {
		// mp3 thing...
		return RevealFromMP3( description, decoy )
	}
}
