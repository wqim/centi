package img
import (
	//"fmt"
	"bytes"
	"image"
	"image/color"
	"golang.org/x/image/bmp"
	"centi/stegano/util"
)

func HideInBMP( decoy, data []byte ) ([]byte, error) {
	// basically, the same as with png
	// just have another package imported
	// todo: make the code more clear and reusable.
	img, err := bmp.Decode( bytes.NewReader( decoy ) )
	if err != nil {
		return nil, err
	}
	encoded, err := util.EncodeToBinary( data )
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	
	rgbaImg := image.NewRGBA( bounds )
	// create RGBA image from source one
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rgbaImg.Set( x, y, img.At(x, y ) )
		}
	}
	// embed bits into least significant bit of the channel
	// according to mode
	bitIndex := 0
	for y := 0; y < height && bitIndex < len( encoded ); y++ {
		for x := 0; x < width && bitIndex < len(encoded); x++ {

			r, g, b, a := rgbaImg.At(x, y).RGBA()
			r8 := uint8(r)
			g8 := uint8(g)
			b8 := uint8(b)
			a8 := uint8(a)

			r8 = (r8 & 0xfe) | encoded[ bitIndex ]
			bitIndex++
			
			if bitIndex < len(encoded) {
				g8 = (g8 & 0xfe) | encoded[ bitIndex ]
				bitIndex++
			}
			if bitIndex < len(encoded) {
				b8 = (b8 & 0xfe) | encoded[ bitIndex ]
				bitIndex++
			}
			rgbaImg.Set( x, y, color.RGBA{r8, g8, b8, a8} )
		}
	}
	buf := new(bytes.Buffer)
	if err = bmp.Encode( buf, rgbaImg ); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func RevealFromBMP( decoy []byte ) ([]byte, error) {
	img, err := bmp.Decode( bytes.NewReader( decoy ) )
	if err != nil {
		return nil, err
	}

	//fmt.Println("Decode format:", format)
	encoded := []uint8{}
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	//fmt.Println("Total pixels:", width * height, "(", width, "x", height, ")")
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			r, g, b, _ := img.At( x, y ).RGBA()
			encoded = append( encoded, uint8(r & 0x1) )
			encoded = append( encoded, uint8(g & 0x1) )
			encoded = append( encoded, uint8(b & 0x1) )
		}
	}
	//fmt.Println("util.DecodeFromBinary()")
	decoded, err := util.DecodeFromBinary( encoded )
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
