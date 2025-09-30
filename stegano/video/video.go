package video
import (
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"
	"gocv.io/x/gocv"

	"centi/stegano/util"
)

var (
	ShredingCount = 7
)

func EncodeWithLSB( data []byte, vid []byte ) ([]byte, error) {
	// unfortunately, gocv does not contain any function for creating video object from
	// memory so we had to bypass it via memfd, pipes and other things.
	/* video, err := gocv.VideoCaptureFile( videoPath )
	 */

	filename, err := util.CreateTempfile( vid )
	if err != nil {
		return nil, err
	}

	defer util.ShredFile( filename )

	video, err := gocv.VideoCaptureFile( filename )
	if err != nil {
		return nil, err
	}

	wfile := ""
	codec := "MJPG"
	// todo: fix this
	width := 1920
	height := 1080

	writer, err := gocv.VideoWriterFile( wfile, codec, 25, width, height, true )
	if err != nil {
		return nil, err
	}

	binData, err := util.EncodeToBinary( data )
	if err != nil {
		return nil, err
	}

	messageEmbedded := false
	bitIndex := 0
	totalBits := len(binData)
	for {
		frame := gocv.NewMat()
		if ok := video.Read( &frame ); !ok || frame.Empty() {
			break
		}

		if !messageEmbedded {
			for y := 0; y < frame.Rows(); y++ {
				for x := 0; x < frame.Cols(); x++ {
					if bitIndex >= totalBits {
						messageEmbedded = true
						break
					}

					pixel := frame.GetVecbAt( x, y )
					// embed in the blue channel (index 0)
					//pixel[0] = setLSB(pixel[0], binData[bitIndex])
					pixel[0] = ( pixel[0] & 0xfe ) | binData[bitIndex]
					frame.SetVecbAt( y, x, pixel )
				}
				if messageEmbedded {
					break
				}
			}
		}

		writer.Write( frame )
		frame.Close()
	}
	if bitIndex < totalBits {
		return nil, fmt.Errorf("Decoy video is too short.")
	}

	// return bytes of the vide
	return writer.Bytes(), nil
}

func DecodeFromLSB(vid []byte ) ([]byte, error) {

	videoPath, err := util.CreateTempfile( vid )
	if err != nil {
		return nil, err
	}
	defer util.ShredFile( videoPath )

	var res []byte
	video, err := gocv.VideoCaptureFile( videoPath )
	if err != nil {
		return nil, err
	}

	for {
		frame := gocv.NewMat()
		if ok := video.Read( &frame ); !ok || frame.Empty()  {
			break
		}

		for y := 0; y < frame.Rows(); y++ {
			for x := 0; x < frame.Cols(); x++ {
				pixel := frame.GecVecbAt( y, x )
				res = append( res, byte(pixel[0] & 0x1) )
			}
		}
		frame.Close()
	}

	return util.DecodeFromBinary( res )
}
