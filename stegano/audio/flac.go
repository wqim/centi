package audio
import (
	"io"
	"fmt"
	"bytes"
	"github.com/mewkiz/flac"
	//"github.com/mewkiz/flac/meta"
	"centi/stegano/util"
)

func HideInFlac( decoy, data []byte ) ([]byte, error) {

	if data == nil || len(data) == 0 {
		return data, nil
	}

	bits, err := util.EncodeToBinary( data )
	if err != nil {
		return nil, err
	}

	stream, err := flac.Parse( bytes.NewReader(decoy) )
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	
	// just encode data in the lsb of each sample.
	idx := 0
	output := bytes.NewBuffer([]byte{})

	encoder, err := flac.NewEncoder( output, stream.Info, stream.Blocks... )
	if err != nil {
		return nil, err
	}

	defer encoder.Close()

	for {
		frame, err := stream.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if err = frame.Parse(); err != nil {
			return nil, err
		}

		//fmt.Println("Frame:", frame)
		//fmt.Println("Amount of subframes: ", len(frame.Subframes))
		for _, subframe := range frame.Subframes {
			if idx >= len(bits) {
				break
			}

			for i, sample := range subframe.Samples {
				if idx >= len(bits) {
					break
				}
				subframe.Samples[i] = ( (sample >> 1) << 1 ) | int32( bits[idx] )
				idx++
			}
		}
		if err = encoder.WriteFrame( frame ); err != nil {
			return nil, err
		}
	}
	if idx < len(bits) {
		return nil, fmt.Errorf("size of flac file is too small.")
	}

	return output.Bytes(), nil
}

func RevealFromFlac( decoy []byte ) ([]byte, error) {
	
	if decoy == nil || len(decoy) == 0 {
		return decoy, nil
	}

	stream, err := flac.Parse( bytes.NewReader(decoy) )
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	// yeah, lsb again because i didn't invent anything more creative and smart
	result := []uint8{}
	for {
		frame, err := stream.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if err = frame.Parse(); err != nil {
			return nil, err
		}
		for _, subframe := range frame.Subframes {
			for _, sample := range subframe.Samples {
				result = append( result, uint8( sample & 0x1 ) )
			}
		}
	}

	return util.DecodeFromBinary( result )
}
