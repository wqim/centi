package archive
import (
	"fmt"
	"bytes"
	"encoding/binary"
	//"centi/util"
)

const (
	// zip file may not have any signature, but still...
	ZipMagic1 = "PK\x03\x04"
	//ZipMagic2 = "PK\x05\x06"	// empty
	//ZipMagic3 = "PK\x07\x08"	// spanned
	// other magic bytes
	CDFH = "\x50\x4b\x01\x02"	// central directory file header
	EOCD = "\x50\x4b\x05\x06"	// end of central directory record
	LFH = "\x50\x4b\x03\x04"	// local file header signature

	// structure sizes constants
	CDESize = 4 * 3 + 5 * 2		// 12 + 10 = 22
	CDSize = 4 * 6 + 11 * 2		// 24 + 22 = 46
	LFHSize = 4 * 4 + 2 * 7		// 30

)

// structures present in zip files
// end of central directory
type ZipCentralDirEnd struct {
	Signature		[4]byte //uint32	// dword
	DiskID			uint16	// word, number of this disk ( 0xffff for zip64 )
	ThisDIskID		uint16	// word, disk where central directory starts
	ThisDiskItemEntries	uint16	// word, number of central directories on this disk
	DiskItemEntries		uint16	// word, total number of central directory records
	SizeOfCentralDir	uint32	// dword, offset of start of central directory in bytes
	LocationOfCentralDir	uint32	// dword, offset of start of central directory, relative to start of archive
	ZipCommentLength	uint16	// word, comment length
	//Comment			[]byte	// optional comment to zip file
}

type ZipEoCD struct {
	eocd		*ZipCentralDirEnd
	Comment		[]byte
}

// central directory
type ZipCentralDir struct {
	Signature			uint32	// dword
	Version				uint16	// word
	ExtractVersion			uint16	// word, version needed to extract
	GeneralPurpose			uint16	// word, general purpose bit flag
	CompressionMethod		uint16	// word, compression method
	LastFileModTime			uint16	// word, last file modification time
	LastFileModDate			uint16	// word, last file modification date
	CRC32				uint32	// dword, crc32 of uncompressed data
	FileCompressedSize		uint32	// dword, compressed size or 0xffffffff for zip64
	FileUncompressedSize		uint32	// dword, uncompressed size or 0xffffffff for zip64
	FileNameLength			uint16	// word, filename length ( n )
	ExtraDataLength			uint16	// word, extra field length ( m )
	FileCommentLength		uint16	// word, file comment length ( k )
	DiskNumberStart			uint16	// word, disk number where file starts ( 0xffff for zip64 )
	InternalFileAttributes		uint16	// word
	ExternalFileAttributes		uint32	// dword
	RelativeOffsetOfLocalHeader	uint32	// dword, relative offset of local file header ( or 0xffffffff for zip 64 )
	// filename (n bytes)
	// extra field ( m bytes )
	// file comment ( k bytes )
}

// local file header
type LocalFileHeader struct {
	Signature		uint32
	Version			uint16
	GeneralPurpose		uint16
	CompressionMethod	uint16
	LastModTime		uint16
	LastModDate		uint16
	CRC32			uint32
	CompressedSize		uint32
	UncompressedSize	uint32
	FilenameLength		uint16 // n
	ExtraFieldLength	uint16 // m
	//FileName		[]byte	// n bytes
	//Extra			[]byte	// m bytes
}

type DataDescriptor struct {
	Signature		uint32	// optional, \x50\x4b\x07\x08
	CRC32			uint32
	CompressedSize		uint32	// may be 8 bytes for zip64
	UncompressedSize	uint32	// may be 8 bytes for zip64
}

// custom structures for easy zip file management
type File struct {
	Header			*LocalFileHeader
	Filename		string
	Extra			[]byte
	Comment			[]byte
	Content			[]byte
	UncompressedSize	uint32
}

type ZipFile struct {
	Files			[]*File
	CentralDirectories	[]*ZipCentralDir
	EoCD			*ZipEoCD
	Embedded		[]byte
}

// extraction things
func FromBytes( data []byte ) (*ZipFile, error) {

	// find the central directory
	var err error
	idx := len(data) - len(CDFH)
	cdSign := []byte(CDFH)
	eocdSign := []byte(EOCD)
	embedded := []byte{}

	files := []*File{}
	cds := []*ZipCentralDir{}
	var eocd *ZipEoCD
	totalDirs := -1
	firstIdx := len(data)
	lastIdx := 0

	//util.DebugPrintln( "Start of file:", data[:16] )
	for {	// extract information about central directory
		if idx < 0 {
			break
		}

		// check if we have an end of central directory
		if bytes.Equal( data[idx : idx + len(eocdSign)], eocdSign ) == true {
			// parse end of central directory
			if eocd == nil {
				eocd, _, err = extractEOCD( data, idx )
				if err != nil {
					return nil, err
				}
				totalDirs = int(eocd.eocd.DiskItemEntries)
			} else {
				return nil, fmt.Errorf("have 2 EoCDs")
			}
		}

		// we can find a central directory entry only if end of central directory was found
		if eocd != nil && totalDirs > 0 {
			// check if we have a central directory header
			if bytes.Equal( data[idx : idx + len(cdSign)], cdSign ) == true {
				// found a central directory header
				cd, err := extractCD( data, idx )
				if err != nil {
					return nil, err
				}
				cds = append( cds, cd )
				// extract comment to file
				i := idx + int( cd.FileNameLength ) + int( cd.ExtraDataLength )
				comment := data[ i : i + int( cd.FileCommentLength ) ]
				// extract file
				f, fi, li, err := extractFile( data, cd )
				if err != nil {
					return nil, err
				}
				// find the end of last file
				if li > lastIdx {
					lastIdx = li
				}
				// find the start of first file
				if fi < firstIdx {
					firstIdx = fi
				}
				// append file
				f.Comment = comment
				/*if len(comment) > 0 {
					util.DebugPrintln("Comment to file: ", string( comment ) )
				}*/
				files = append( files, f )
				totalDirs--
			}
		}
		if totalDirs == 0 {
			break
		}
		idx--
	}

	if eocd == nil {
		return nil, fmt.Errorf("Did not find an end of central directory")
	}

	// check if any data was embedded in the start
	if firstIdx > 0 {
		// found any data before first file entry
		embedded = append( embedded, data[ :firstIdx ]... )
	}

	// check if any data was embedded in the end
	if lastIdx < idx {
		embedded = append( embedded, data[ lastIdx : idx ]... )
	}

	return &ZipFile{
		files,
		cds,
		eocd,
		embedded,
	}, nil
}

func extractEOCD( data []byte, idx int ) (*ZipEoCD, int, error) {
	if idx + CDESize > len(data) {
		return nil, -1, fmt.Errorf("[extractEOCD] too small data to unpack")
	}
	packed := data[idx:idx + CDESize]
	var eocd ZipCentralDirEnd
	_, err := binary.Decode( packed, binary.LittleEndian, &eocd )
	if err != nil {
		return nil, -1, err
	}
	idx += int( CDESize )
	comment := data[ idx : idx + int(eocd.ZipCommentLength) ]
	idx += int( eocd.ZipCommentLength )

	zipEoCD := &ZipEoCD{ &eocd, comment }
	//util.DebugPrintln("EOCD structure:", eocd)
	return zipEoCD, idx, nil
}

func extractCD( data []byte, idx int ) (*ZipCentralDir, error) {
	if idx + CDSize > len(data) {
		return nil, fmt.Errorf("[extractCD] too small data to unpack")
	}
	packed := data[idx:idx + CDSize]
	var cd ZipCentralDir
	_, err := binary.Decode( packed, binary.LittleEndian, &cd )
	if err != nil {
		return nil, err
	}
	//util.DebugPrintln("Extracted central directory:", cd)
	return &cd, nil
}

func extractFile( data []byte, cd *ZipCentralDir ) (*File, int, int, error) {
	// find file in the archive
	off := cd.RelativeOffsetOfLocalHeader
	if off + LFHSize > uint32(len(data)) {
		return nil, -1, -1, fmt.Errorf("invalid offset specified")
	}

	tmp := data[ off : off + LFHSize ]
	var fileHdr LocalFileHeader
	if _, err := binary.Decode( tmp, binary.LittleEndian, &fileHdr ); err != nil {
		return nil, -1, -1, err
	}
	// check for data descriptor
	if cd.GeneralPurpose & 0x8 == 0x8 {
		if cd.CRC32 == 0 && cd.FileCompressedSize == 0 {
			// todo...
		}
	}
	// parsed local file header, get file content now...
	off += LFHSize // local file header size
	// extract name of file
	filename := data[ off : off + uint32(fileHdr.FilenameLength) ]
	off += uint32(fileHdr.FilenameLength)
	// extract some extra information
	extra := data[ off : off + uint32(fileHdr.ExtraFieldLength) ]
	off += uint32(fileHdr.ExtraFieldLength)

	// extract content of file
	content := data[ off : off + cd.FileCompressedSize ]
	// now we have a full file info,
	/*if len(filename) < 100 {
		util.DebugPrintln("Extracted file: ", string(filename))
	}
	if len(extra) > 0 {
		util.DebugPrintln("Extra info:", extra)
	}*/
	return &File{
		&fileHdr,
		string( filename ),
		extra,
		nil,
		content,
		cd.FileUncompressedSize,
	}, int(cd.RelativeOffsetOfLocalHeader), int(off) + int(cd.FileCompressedSize), nil
}

// steganography things
func(z *ZipFile) Embed( data []byte ) error {
	//util.DebugPrintln("Data embedded previously:", z.Embedded)
	z.Embedded = data
	return nil
}

func(z *ZipFile) Extract() ([]byte, error) {
	return z.Embedded, nil
}

// other functions
func(z *ZipFile) SetComment( comment []byte ) {
	z.EoCD.Comment = comment
	z.EoCD.eocd.ZipCommentLength = uint16( len(comment) )
}

// dumping things
func(z *ZipFile) Bytes() ([]byte, error) {
	
	// dump of files
	res := []byte{}
	// dump of central directories
	cds := []byte{}

	// go over all of the central directories
	for i := len(z.Files)-1; i >= 0; i-- {
		// len(centralDirectories) = len(files)
		cd := z.CentralDirectories[i]
		f := z.Files[i]
		// dump local file header
		fdump := make([]byte, LFHSize)
		if _, err := binary.Encode( fdump, binary.LittleEndian, *f.Header ); err != nil {
			return nil, err
		}

		// dump file
		fdump = append( fdump, []byte( f.Filename )... )
		fdump = append( fdump, f.Extra... )
		fdump = append( fdump, f.Content... )

		/*if f.Filename[len(f.Filename)-1] == byte('/') {
			util.DebugPrintln("Found entry: ", f.Filename )
			util.DebugPrintln("Len(extra):", len(f.Extra), "/", f.Header.ExtraFieldLength)
			util.DebugPrintln("Len(content):", len(f.Content) )
			util.DebugPrintln("Len(comment):", len(f.Comment) )
		}*/

		// fix cd values
		//cd.RelativeOffsetOfLocalHeader = uint32(len(res))

		res = append( res, fdump... )

		// dump central directory
		buff := make([]byte, CDSize)
		if _, err := binary.Encode( buff, binary.LittleEndian, *cd ); err != nil {
			return nil, err
		}

		// append filename, extra field and file comment
		buff = append( buff, []byte(f.Filename)... )
		buff = append( buff, []byte(f.Extra[:cd.ExtraDataLength])... )
		buff = append( buff, []byte(f.Comment)... )
		
		cds = append( cds, buff... )
	}
	// if any embedded data known, embed it
	result := res
	result = append( result, z.Embedded... )
	
	//z.EoCD.eocd.LocationOfCentralDir = uint32(len(result))
	result = append( result, cds... )

	//util.DebugPrintln("EoCD signature (int):", z.EoCD.eocd.Signature)
	// dump the end of central directory

	buffer := make( []byte, CDESize + len(z.EoCD.Comment) )
	if _, err := binary.Encode( buffer, binary.LittleEndian, *z.EoCD.eocd ); err != nil {
		return nil, err
	}
	//util.DebugPrintln("After encoding binary:", buffer[:4])
	// append comment to zip file
	buffer = append( buffer, z.EoCD.Comment... )
	// append the end of central directory
	result = append( result, buffer... )
	return result, nil
}
