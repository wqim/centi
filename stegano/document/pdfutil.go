package document
import (
	"os"
	"os/exec"
	"fmt"
	"bytes"
	"regexp"
	"strconv"
	"centi/stegano/util"
)

const (
	NumsRegex = `[\d\.\-]+`
	TjNumsRegex = `[\d\.\-]+(?![^\(]*\))(?![^\<]*\>)`

	pdftk = "pdftk"
	qpdf = "qpdf"
)

var (
	ioperators = []IOperator{}
)

func RunPdfProcessor( pdf []byte, action string ) ([]byte, error) {
	// create temporary files
	tempfile, err := util.CreateTempfile( pdf )
	if err != nil {
		return nil, err
	}
	defer util.ShredFile( tempfile )

	tempfile2, err := util.CreateTempfile( nil )
	if err != nil {
		return nil, err
	}
	defer util.ShredFile( tempfile2 )

	// try to find qpdf
	args := []string{}
	q, err := util.PathToProgram( qpdf )
	if err != nil {
		// qpdf is not installed, find pdftk
		q, err = util.PathToProgram( pdftk )
		if err != nil {
			return nil, err
		} else {
			args = []string{
				tempfile,
				"output",
				tempfile2,
				action,
			}
		}
	} else {
		args = []string{
			tempfile,
			"--stream-data=" + action,
			tempfile2,
		}
	}

	// path determined, arguments too, run the command
	cmd := exec.Command( q, args... )
	fmt.Println( cmd )
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println( "Failed to ", action, " pdf file: ", err.Error() )
		fmt.Println( string(out) )
		return nil, err
	}

	// just read the result
	return os.ReadFile( tempfile2 )
}

func CompressPdf( pdf []byte ) ([]byte, error) {
	// pdftk <in_file> output <out_file> compress
	// qpdf <in_file> --stream-data=compress <out_file>
	return RunPdfProcessor( pdf, "compress" )
}

func DecompressPdf( pdf []byte ) ([]byte, error) {
	// pdftk <in_file> output <out_file> uncompress 
	// qpdf <in_file> --stream-data=uncompress <out_file>
	return RunPdfProcessor( pdf, "uncompress" )
}

// oop-like things for encoding.
// originated from https://github.com/rcklemm/PDF_Steg_Research/blob/main/full_pdf_steg.py
type PdfStream struct {
	start		int
	end		int
	text		string
	viableStream	bool
}

func NewPdfStream( start, end int, text string ) *PdfStream {
	numNotPrintable := 0
	for _, t := range text {
		if strconv.IsPrint( t ) == false {
			numNotPrintable++
		}
	}

	pctUnprintable := 100 * ( numNotPrintable / len(text) )

	return &PdfStream{
		start,
		end,
		text,
		( pctUnprintable < 25 ),
	}
}

func(p *PdfStream) Viable() bool {
	return p.viableStream
}

// operator interfaee
type IOperator interface {
	FindAll( text []byte ) [][]int
	Embed( match []byte, bits []uint8 ) (string, int, error)
	Extract( match []byte ) ([]uint8, error)
	getRegexNumberCapture() *regexp.Regexp
	getBitsPerOperand() int
}

// operator type
type Operator struct {
	opStr			string
	minOperands		int
	maxOperands		int
	bitsPerOperand		int
	minValue		float64
	maxPcts			[]float64
	regexNumberCapture	*regexp.Regexp
	pattern			*regexp.Regexp
}

func NewOperator( 	opStr string,
			minOps, maxOps, bpo int,
			minVal float64,
			maxPcts []float64 ) Operator {

	patternStr := `(?:[\d\.\-]+\s+)` + "{" + 
			strconv.Itoa( minOps ) + "," + strconv.Itoa( maxOps ) + "}" +
			opStr + `[\[\s]`

	regNumCap := regexp.MustCompile( NumsRegex )
	pattern := regexp.MustCompile( patternStr )

	return Operator{
		opStr,
		minOps,
		maxOps,
		bpo,
		minVal,
		maxPcts,
		regNumCap,
		pattern,
	}
}

func(o Operator) getBitsPerOperand() int {
	return o.bitsPerOperand
}

func(o Operator) getRegexNumberCapture() *regexp.Regexp {
	return o.regexNumberCapture
}

func(o Operator) FindAll( text []byte ) [][]int {
	return o.pattern.FindAllIndex( text, -1 )
}

func(o Operator) Embed( match []byte, bits []uint8 ) (string, int, error) {

	parts := o.regexNumberCapture.FindAllIndex( match, -1 )
	if parts == nil || len(parts) == 0 {
		return "", -1, fmt.Errorf("Nowhere to embed")
	}

	totalNumBits := o.bitsPerOperand * len(parts)

	// 'problem line' split onto other lines
	bitPieces := [][]uint8{}
	minLen := len(bits)
	if minLen > totalNumBits {
		minLen = totalNumBits
	}

	for i := 0; i < minLen; i+= totalNumBits {
		bitPieces = append( bitPieces, bits[ i : i + o.bitsPerOperand ] )
	}

	// continuation
	replacement := ""
	matchIndex := 0
	partIndex := 0
	bitsHidden := 0

	// actual embedding...
	ip := 0
	ib := 0
	for {
		p := parts[ip]
		b := bitPieces[ib]

		num := match[ p[0] : p[1] ]
		replacement += string( embedBit( num, o.maxPcts[ partIndex ] / 100.0, o.bitsPerOperand, b, o.minValue ) )
		matchIndex = p[1]
		partIndex += 1
		bitsHidden += len( p )

		ip++
		ib++
		if ip >= len(parts) {
			break
		}
		if ib >= len(bitPieces) {
			break
		}
	}
	replacement += string(match[ matchIndex: ])
	return replacement, bitsHidden, nil
}

func(o Operator) Extract( match []byte ) ([]uint8, error) {
	parts := o.regexNumberCapture.FindAllIndex( match, -1 )
	if parts == nil || len(parts) == 0 {
		return nil, fmt.Errorf("Nothing to extract")
	}

	bits := []uint8{}
	for _, p := range parts {
		num := match[ p[0]:p[1] ]
		str := 	formatExtracted(
				int(extractBit(num, o.bitsPerOperand)),
				o.bitsPerOperand,
			)

		for _, ch := range str {
			if ch == '0' {
				bits = append(bits, 0)
			} else if ch == '1' {
				bits = append(bits, 1)
			} else {
				fmt.Println("[ERROR] operator::Extract: ch = ", ch)
			}
		}
	}
	return bits, nil
}

// and tj operator type, inherited from Operator.
type TjOperator struct {
	opStr			string
	maxPct			int
	bitsPerOperand		int
	minValue		float64
	regexNumberCapture	*regexp.Regexp
	pattern			*regexp.Regexp
}

func NewTjOperator(	opStr string,
			maxPct int,
			bitsPerOperand int,
			minValue float64 ) TjOperator {

	pattern := regexp.MustCompile( `\[.+?\]\s*?TJ` )
	regNumCap := regexp.MustCompile( TjNumsRegex )

	return TjOperator{
		opStr,
		maxPct,
		bitsPerOperand,
		minValue,
		regNumCap,
		pattern,
	}
}

// dummy for compiler to calm down
func(o TjOperator) FindAll( text []byte ) [][]int {
	return nil
}

func(o TjOperator) getRegexNumberCapture() *regexp.Regexp {
	return o.regexNumberCapture
}

func(o TjOperator) getBitsPerOperand() int {
	return o.bitsPerOperand
}

func(o TjOperator) Embed( match []byte, bits []uint8 ) (string, int, error) {
	re, err := regexp.Compile(`\(`)
	if err != nil {
		return "", -1, err
	}
	match = re.ReplaceAll( match, []byte("..") )
	parts := o.regexNumberCapture.FindAllIndex( match, -1 )
	if parts == nil || len(parts) == 0 {
		return "", -1, fmt.Errorf("Nowhere to embed (tj operator)")
	}

	totalNumBits := o.bitsPerOperand * len(parts)
	
	// split problem line again...
	bitPieces := [][]uint8{}
	minLen := len(bits)
	if minLen > totalNumBits {
		minLen = totalNumBits
	}
	for i := 0; i < minLen; i+= totalNumBits {
		bitPieces = append( bitPieces, bits[ i : i + o.bitsPerOperand ] )
	}

	// actually finding replacement
	replacement := ""
	matchIndex := 0
	bitsHidden := 0

	ip := 0
	ib := 0

	for {
		p := parts[ip]
		b := bitPieces[ib]

		replacement += string( match[ matchIndex : p[0] ] )
		num := match[ p[0] : p[1] ]
		replacement += string( embedBit( num, float64(o.maxPct) / 100.0, o.bitsPerOperand, b, o.minValue ) )
		matchIndex = p[1]
		bitsHidden += len(b)

		ip++
		ib++
		if ip >= len(parts) {
			break
		}
		if ib >= len(bitPieces) {
			break
		}
	}
	replacement += string( match[ matchIndex: ] )
	return replacement, bitsHidden, nil
}

func(o TjOperator) Extract( match []byte ) ([]uint8, error) {

	re, err := regexp.Compile(`\(`)
	if err != nil {
		return nil, err
	}
	match = re.ReplaceAll( match, []byte("..") )
	parts := o.regexNumberCapture.FindAllIndex( match, -1 )	// check out this line
	if parts == nil || len(parts) == 0 {
		return nil, fmt.Errorf("Nothing to extract (tj operator)")
	}

	bits := []uint8{}
	for _, p := range parts {
		num := match[p[0]:p[1]]
		toAppend := formatExtracted(
			int(extractBit(num, o.bitsPerOperand)),
			o.bitsPerOperand,
		)
		for _, bit := range toAppend {
			if bit == '0' {
				bits = append( bits, 0 )
			} else if bit == '1' {
				bits = append( bits, 1 )
			}
		}
	}
	return bits, nil
}

// actual bit embedding/extraction functions
func embedBit( strOp []byte, pct float64, n int, bits []byte, minValue float64) []byte {
	// general properties of property of pdf object
	orig := strOp
	negative := bytes.Contains( strOp, []byte("-") )
	floatingPoint := bytes.Contains( strOp, []byte(".") )
	pointLoc := len( strOp )
	if floatingPoint == true {
		pointLoc = bytes.Index( strOp, []byte(".") )
	}
	strOp = bytes.ReplaceAll( strOp, []byte("-"), []byte{} )
	
	// count leading zeroes
	leadingZeroes := 0
	if floatingPoint == true {
		for _, c := range bytes.ReplaceAll( strOp, []byte("."), []byte{} ) {
			if c == 0x30 {	// '0'
				leadingZeroes++
			} else {
				break
			}
		}
	}

	// find the mask
	intOp, _ := strconv.Atoi( string(bytes.ReplaceAll(strOp, []byte("."), []byte{} )) )
	if intOp > 255 {
		fmt.Println("[ERROR] stegano/pdfutil.go: embedBit: intOp = ", intOp)
	}

	mask := util.FromBin( bits )
	if extractBit( strOp, n ) == mask {
		return orig
	}

	// special case for operator = 0
	if intOp == 0 {
		// i don't really understand what's going on here...
		if float64(util.FromBin( bits )) <= minValue {
			return []byte( strconv.Itoa( int(util.FromBin( bits )) ) )
		}

		numZeroes := 0
		for {
			strVal := "0." + string( bytes.Repeat([]byte("0"), numZeroes) ) + strconv.Itoa( int(util.FromBin(bits)) )
			var tmp float64
			if _, err := fmt.Scanf("%f", strVal, &tmp); err != nil {
				return nil
			}
			if tmp < minValue {
				return []byte( strVal )
			}
			numZeroes++
		}
	}

	// determine if our change is small enough...?
	smallEnoughChange := false
	workingIntOp := intOp
	for smallEnoughChange == false {
		// todo...
		newIntOp := workingIntOp ^ ( workingIntOp & int(util.FromBin( bytes.Repeat([]byte{1}, n) )) )
		newIntOp = newIntOp | int(mask)

		if util.Abs( newIntOp - workingIntOp ) <= int(pct) * workingIntOp {
			smallEnoughChange = true
			workingIntOp = int(newIntOp)
		} else {
			workingIntOp *= 10
		}
	}

	returnStr := ""
	if negative {
		returnStr += "-"
	}
	returnStr += string( bytes.Repeat( []byte("0"), leadingZeroes ) )
	returnStr += strconv.Itoa( workingIntOp )

	if returnStr[ pointLoc: ] != "" {
		returnStr = returnStr[ 0 : pointLoc ] + "." + returnStr[ pointLoc: ]
	}
	return []byte(returnStr)
}

func extractBit( strOp []byte, n int ) byte {
	intOp, _ := strconv.Atoi( string( bytes.ReplaceAll(strOp, []byte("."), []byte{}) ) )
	mask := util.FromBin( bytes.Repeat([]byte("1"), n)  )
	return byte(intOp & 0xff) & mask
}

func formatExtracted( extracted, n int ) string {
	e := strconv.Itoa( extracted )
	return string( bytes.Repeat([]byte("0"), ( n - len(e) ) ) ) + e
}

func findAllStreams( file []byte ) []*PdfStream {
	//cursor := 0
	start := 0
	end := 0
	streams := []*PdfStream{}
	// find all streams in the pdf file that are likely to be text streams
	for {
		_start := bytes.Index( file, []byte("stream") )
		_end := bytes.Index( file, []byte("endstream") )

		if _start >= 0 && _end >= 0 {
			start = _start
			end = _end

			if bytes.Equal( file[ start + 6 : start + 8 ], []byte("\r\n") ) == true {
				stream := NewPdfStream( start + 8, end - 2, string(file[ start + 8 : end - 2]) )
				if stream.Viable() {
					streams = append( streams, stream )
				}
			} else {
				stream := NewPdfStream( start + 7, end - 1, string(file[ start + 7 : end - 1]) )
				if stream.Viable() {
					streams = append( streams, stream )
				}
			}
			start = end + 9
		}
	}
	return streams
}

// auxilary structure
type Match struct {
	op	IOperator
	s	*PdfStream
}

func collectAllMatches( operators []IOperator, text []byte ) []*Match {
	matches := []*Match{}
	for _, op := range operators {
		opMatches := op.FindAll( text )
		for _, o := range opMatches {
			stream := NewPdfStream( o[0], o[1], string(text[o[0]:o[1]]) )
			matches = append( matches, &Match{ op, stream } )
			//matches = append( matches, op )
		}
	}
	// todo: sort matches by start of match
	return matches
}

func Stat( file []byte ) int {
	streams := findAllStreams( file )
	numBits := 0
	for _, s := range streams {
		text := s.text
		matches := collectAllMatches( ioperators, []byte(text) )
		for _, m := range matches {
			numberCount := m.op.getBitsPerOperand() * len(
				m.op.getRegexNumberCapture().FindAll(
					[]byte(text[ m.s.start:m.s.end ]),
					-1,
				),
			)
			numBits += numberCount
		}
	}
	bytesAvailable := numBits / 8
	return bytesAvailable
}
