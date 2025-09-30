package img
import (
	"fmt"
	"bytes"
	"image/gif"
	"centi/stegano/util"
)

func HideInGif( gifbytes []byte, data []byte ) ([]byte, error) {
	g, err := gif.DecodeAll( bytes.NewReader( gifbytes ) )
	if err != nil {
		return nil, err
	}
	// flatten data bits
	bits, err := util.EncodeToBinary( data )
	if err != nil {
		return nil, err
	}
	// embed bits into pixel indicies
	bitIdx := 0
	for frameIdx, frame := range g.Image {
		for i := range frame.Pix {
			if bitIdx >= len(bits) {
				break
			}
			// modify least significant bit
			frame.Pix[i] = (frame.Pix[i] & 0xfe ) | bits[bitIdx]
			bitIdx++
		}
		g.Image[frameIdx] = frame
		if bitIdx >= len(bits) {
			break
		}
	}
	if bitIdx < len(bits) {
		return nil, fmt.Errorf("GIF file is too small")
	}

	outbuf := bytes.NewBuffer( []byte{} )
	err = gif.EncodeAll( outbuf, g )
	if err != nil {
		return nil, err
	}
	return outbuf.Bytes(), nil
}

func RevealFromGif( gifbytes []byte ) ([]byte, error) {
	g, err := gif.DecodeAll( bytes.NewReader( gifbytes ) )
	if err != nil {
		return nil, err
	}
	bits := []uint8{}
	for _, frame := range g.Image {
		for _, pix := range frame.Pix {
			bits = append( bits, uint8(pix & 1) )
		}
	}

	return util.DecodeFromBinary( bits )
}
