package img
import (
	"fmt"
	"bytes"
	"image/jpeg"
	"encoding/binary"
	"lukechampine.com/jsteg"
)

func HideInJpeg( jpgBytes []byte, data []byte ) ([]byte, error) {
	
	img, err := jpeg.Decode( bytes.NewBuffer( jpgBytes ) )
	if err != nil {
		return nil, err
	}
	cap := jsteg.Capacity( img, nil )
	if cap < len(data) + 8 {
		return nil, fmt.Errorf("Not enough space to embed data ( %d < %d )", cap, len(data) + 8 )
	}

	newData := make([]byte, len(data) + 8)
	binary.LittleEndian.PutUint64( newData, uint64(len(data)) )
	copy( newData[8:], data )

	outbuf := bytes.NewBuffer( []byte{} )
	jsteg.Hide( outbuf, img, newData, nil )
	return outbuf.Bytes(), err
}

func RevealFromJpeg( jpgBytes []byte ) ([]byte, error) {
	if jpgBytes == nil || len(jpgBytes) == 0 {
		return jpgBytes, nil
	}
	hidden, err := jsteg.Reveal( bytes.NewBuffer(jpgBytes) )
	if err != nil {
		return nil, err
	}

	size := binary.LittleEndian.Uint64( hidden[:8] )
	if uint64(len( hidden ) - 8) < size {
		return nil, fmt.Errorf("JPEG: Invalid length encoding")
	}
	return hidden[8:8+size], nil
}
