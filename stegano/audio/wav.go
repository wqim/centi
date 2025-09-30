package audio
import (
	//"fmt"
	"math"
	"bytes"
	"centi/stegano/goaudio/wave"
	"centi/stegano/util"
)

var (
	ScaleFactor = 1000.0
)

func HideInWav( data []byte, audio []byte ) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return data, nil
	}

	wv, err := wave.ReadWaveFromReader( bytes.NewBuffer( audio ) )
	if err != nil {
		return nil, err
	}
	data, err = util.EncodeToBinary( data )
	if err != nil {
		return nil, err
	}

	frames := encodeWithLSB( data, wv.Frames )
	newBuf := bytes.NewBuffer( []byte{} )
	if err = wave.WriteWaveToWriter( frames, wv.WaveFmt, newBuf ); err != nil {
		return nil, err
	}
	return newBuf.Bytes(), nil
}

func RevealFromWav( audio []byte ) ([]byte, error) {
	if audio == nil || len(audio) == 0 {
		return audio, nil
	}

	wv, err := wave.ReadWaveFromReader( bytes.NewBuffer( audio ) )
	if err != nil {
		return nil, err
	}
	return decodeFromLSB( wv.Frames )
}


func encodeWithLSB( data []byte, samples []wave.Frame ) []wave.Frame {
	for i, s := range samples {
		if i < len(data) {
			// is := ( uint64(float64(s)) & 0xffffffff_fffffffe ) | uint64( data[i] )
//is := ( F2i( float64(s) ) & 0xffffffff_fffffffe ) | uint64( data[i] )
			//is := ( F2i( float64(s) ) & 0xffffffff_fffffffe ) | uint64( data[i] )
			is := float64( uint64( float64(s) ) ) + float64( data[i] ) / ScaleFactor
			//samples[i] = wave.Frame( I2f(is) )
			samples[i] = wave.Frame( is )
			//fmt.Println( s, samples[i] )
		} else {
			break
		}
	}
	return samples
}

func decodeFromLSB( samples []wave.Frame ) ([]byte, error) {
	res := []byte{}
	for _, s := range samples {
		//is := uint64( float64(s) )
		//is := F2i( float64(s) ) & 0x1
		is := float64(s) - float64( uint64(float64(s)) )
		if is * ScaleFactor > 0.0 {
			res = append( res, 1 )
		} else {
			res = append( res, 0 )
		}
		/*if is != 0 {
			fmt.Println( is, uint8( is & 0x1) )
		}*/
		
	}
	return util.DecodeFromBinary( res )
}

func F2i( f float64 ) uint64 {
	/*f64 := make( []byte, 8 )
	binary.LittleEndian.PutFloat64( f64, f )
	res := binary.LittleEndian.Uint64( f64 ) */
	return math.Float64bits( f )
}

func I2f( i uint64 ) float64 {
	/*i64 := make( []byte, 8 )
	binary.LittleEndian.PutUint64( i64, i )
	res := binary.LittleEndian.Float64( i64 )*/
	return math.Float64frombits( i )
}
