package document
import (
	"fmt"
	"bytes"
	"encoding/binary"
	"centi/stegano/util"
)

const (
	// modes of steganography
	AfterEOF = uint8(0)	// hide data after EOF in pdf...
	CRNL = uint8(1)	// carriage return, new line
	OperatorMode = uint8(2)	// mode in which we utilize pdf operators to hide data
)

var (
	zeros = []byte("\r\n")
	ones = []byte("\n")
)

func HideInPdf( mode uint8, decoy []byte, data []byte ) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return data, nil
	}

	switch mode {
	case AfterEOF:
		return EmbedAfterEOF( decoy, data )
	case CRNL:
		return EmbedUsingNewline( decoy, data )
	case OperatorMode:
		return EmbedUsingOperator( decoy, data )
	}
	return nil, fmt.Errorf("unknown mode")

}

func RevealFromPdf( mode uint8, data []byte ) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return data, nil
	}

	switch mode {
	case AfterEOF:
		return ExtractAfterEOF( data )
	case CRNL:
		return ExtractUsingNewline( data )
	case OperatorMode:
		return ExtractUsingOperator( data )
	}
	return nil, fmt.Errorf("unknown mode")
}

/*
 * different methods of embedding/extracting data in/from pdf
 */
func EmbedAfterEOF( pdf []byte, data []byte ) ([]byte, error) {
	//fmt.Println("Last bytes of pdf:", pdf[len(pdf) - 16:] )
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64( buf, uint64(len(data) ) )
	tmp := append( pdf, data... )
	return append( tmp, buf... ), nil
}

func ExtractAfterEOF( pdf []byte ) ([]byte, error) {
	idx := bytes.Index( pdf, []byte("%%EOF") )
	if idx < 0 {
		return nil, fmt.Errorf("Pdf EOF not found")
	}
	tmp := pdf[ idx+5: ]
	if len(tmp) < 8 {
		return nil, fmt.Errorf("No valid data embedded")
	}
	var size uint64
	binary.Decode( tmp[len(tmp)-8:], binary.LittleEndian, &size )
	if len(tmp) - 8 < int(size) {
		return nil, fmt.Errorf("No valid data embedded (2)")
	}
	//fmt.Println("[*] Size of data to extract:", size)
	return tmp[ len(tmp) - int(size) - 8 : ][:size], nil
}

func EmbedUsingNewline( pdf []byte, data_ []byte ) ([]byte, error) {

	data, err := util.EncodeToBinary( data_ )
	if err != nil {
		return nil, err
	}

	// check how many bits we can embed
	pdf = bytes.ReplaceAll( pdf, []byte("\r\n"), []byte("\n") )
	totalCount := bytes.Count( pdf, []byte("\n") )
	if totalCount < len(data) {
		return nil, fmt.Errorf("too small pdf document to embed data, try another method")
	}

	// embed data
	i := 0
	idx := 0
	for i < len(pdf) && idx < len(data) {
		if pdf[i] == byte('\n') {
			// check current bit
			if data[ idx ] == 0 {
				pdf = append( append( pdf[:i], zeros... ), pdf[ i + 1: ]... )
				i++	// because len("\r\n") = 2
			}
			idx++
		}
		i++
	}
	return pdf, nil
}

func ExtractUsingNewline( pdf []byte ) ([]byte, error) {

	result := []uint8{}
	i := 0
	for i < len(pdf) {
		if pdf[i] == byte('\r') {
			if i + 1 < len(pdf) && pdf[ i + 1 ] == byte('\n') {
				// found zeros sequence
				result = append( result, 0 )
				i++
			}
		} else if pdf[i] == byte('\n') {
			result = append( result, 1 )
		}
		i++
	}
	return util.DecodeFromBinary( result )
}

func EmbedUsingOperator( decoy, data []byte ) ([]byte, error) {
	msgbits, err := util.EncodeToBinary( data )
	if err != nil {
		return nil, err
	}

	decoy, err = DecompressPdf( decoy )
	if err != nil {
		return nil, err
	}

	streams := findAllStreams( decoy )
	newStreams := []*PdfStream{}

	msgIndex := 0
	bytesAdded := 0
	for _, s := range streams {
		text := s.text
		if msgIndex >= len( msgbits ) {
			newStreams = append( newStreams, s )
			continue
		}

		matches := collectAllMatches( ioperators, []byte(text) )
		if len(matches) == 0 {
			newStreams = append( newStreams, s )
			continue
		}

		newText := ""
		textIndex := 0
		for _, m := range matches {
			newText += text[ textIndex : m.s.start ]
			bits := msgbits[ msgIndex: ]
			replacement, bitsHidden, err := m.op.Embed( []byte(text[ m.s.start:m.s.end ]), bits )
			if err != nil {
				return nil, err
			}

			bytesAdded += len( replacement ) - ( m.s.end - m.s.start )
			msgIndex += bitsHidden
			newText += replacement

			textIndex = m.s.end
		}

		newText += text[ textIndex: ]
		sNew := &PdfStream{
			s.start,
			s.end,
			newText,
			s.viableStream,
		}
		newStreams = append( newStreams, sNew )
	}

	// assemble the output pdf
	fileIndex := 0
	it := 0
	in := 0
	out := []byte{}
	cursor := 0
	for {

		t := streams[it]
		n := newStreams[in]
		
		out = append( out, decoy[cursor : cursor + t.start - fileIndex]... )
		out = append( out, n.text... )
		cursor = fileIndex

		it++
		in++
		if it >= len(streams) {
			break
		}
		if in >= len(newStreams) {
			break
		}
	}
	out = append( out, decoy[cursor:]... )
	return CompressPdf( out )
}

func ExtractUsingOperator( decoy []byte ) ([]byte, error) {

	decoy, err := DecompressPdf( decoy )
	if err != nil {
		return nil, err
	}

	streams := findAllStreams( decoy )
	// get all bits
	msgbits := []uint8{}
	for _, s := range streams {
		text := s.text
		matches := collectAllMatches( ioperators, []byte(text) )
		for _, m := range matches {
			res, err := m.op.Extract( []byte(text[m.s.start:m.s.end]) )
			if err != nil {
				return nil, err
			}
			msgbits = append( msgbits, res... )
		}
	}

	return util.DecodeFromBinary( msgbits )
}
