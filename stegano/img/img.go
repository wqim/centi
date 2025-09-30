package img
import (
	"fmt"
)

func Hide( decoy, data []byte ) ([]byte, error) {
	if decoy[0] == 0x47 && decoy[1] == 0x49 && decoy[2] == 0x46 {
		// a gif image
		//fmt.Println("GIF image")
		return HideInGif( decoy, data )
	}
	if decoy[0] == 0x89 && decoy[1] == 0x50 && decoy[2] == 0x4e &&
		decoy[3] == 0x47 && decoy[4] == 0x0d && decoy[5] == 0x0a &&
		decoy[6] == 0x1a && decoy[7] == 0x0a {
		// a png image
		//fmt.Println("PNG image")
		return EncodeWithLSB( RMode | GMode | BMode, data, decoy )
	}

	if decoy[0] == 0xff && decoy[1] == 0xd8 && decoy[2] == 0xff {
		// a jpeg image
		//fmt.Println("JPEG image")
		return HideInJpeg( decoy, data )
	}

	if decoy[0] == 0x42 && decoy[1] == 0x4d {
		// bmp image
		return HideInBMP( decoy, data )
	}
	return nil, fmt.Errorf("Unsupported image format.")
}

func Reveal( decoy []byte ) ([]byte, error) {
	if decoy[0] == 0x47 && decoy[1] == 0x49 && decoy[2] == 0x46 {
		// a gif image
		//fmt.Println("GIF image")
		return RevealFromGif( decoy )
	}
	if decoy[0] == 0x89 && decoy[1] == 0x50 && decoy[2] == 0x4e &&
		decoy[3] == 0x47 && decoy[4] == 0x0d && decoy[5] == 0x0a &&
		decoy[6] == 0x1a && decoy[7] == 0x0a {
		// a png image
		//fmt.Println("PNG image")
		return DecodeFromLSB( RMode | GMode | BMode, decoy )
	}

	if decoy[0] == 0xff && decoy[1] == 0xd8 && decoy[2] == 0xff {
		// a jpeg image
		//fmt.Println("JPEG image")
		return RevealFromJpeg( decoy )
	}
	if decoy[0] == 0x42 && decoy[1] == 0x4d {
		// bmp image
		return RevealFromBMP( decoy )
	}
	return nil, fmt.Errorf("Unsupported image format.")
}
