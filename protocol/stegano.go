package protocol
import (
	"os"
	"fmt"
	"strings"
	"centi/util"
	//"centi/config"
	"centi/stegano/img"
	"centi/stegano/text"
	"centi/stegano/audio"
	"centi/stegano/archive"
	//"centi/stegano/video"
	"centi/stegano/document"
)

const (
	TextFile = int8(0)	// actually, it also can be a code file
	ImageFile = int8(1)
	AudioFile = int8(2)
	//VideoFile = int8(3)
	DocumentFile = int8(4)
	ArchiveFile = int8(5)
	UnknownFile = int8(-1)
)

func DetermineFileType( ext string ) int8 {

	supportedImages := []string{"png", "jpeg", "jpg", "gif"}
	supportedTexts := []string{
		"txt", "md", "py", "java", "rs",
		"go", "sql", "c", "cpp", "h", "hpp",
		"ts", "js", "nim", "toml", "conf",
	}
	supportedAudios := []string{"wav"}
	//supportedVideos := []string{"mp4"}
	supportedDocuments := []string{"pdf"}
	// some documents are easier to treat as archives.
	supportedArchives := []string{ "zip", "xlsx", "xlsm", "docx", "odt", "xlsx", "ods" }

	types := map[int8][]string{
		TextFile: supportedTexts,
		ImageFile: supportedImages,
		AudioFile: supportedAudios,
		//VideoFile: supportedVideos,
		DocumentFile: supportedDocuments,
		ArchiveFile: supportedArchives,
	}

	//util.DebugPrintln("DetermineFileType: ext =", ext)
	for t, v := range types {
		for _, val := range v {
			if val == ext {
				return t
			}
		}
	}
	return UnknownFile
}

func DetermineSteganoMethods( fileExtensions []string ) []int8 {
	res := []int8{}
	for _, ext := range fileExtensions {
		typ := DetermineFileType( ext )
		doContinue := false
		for _, t := range res {
			if t == typ {
				doContinue = true
			}
		}
		if doContinue == false {
			res = append( res, typ )
		}
	}
	return res
}

func HideInFile(
	folder string,
	extensions []string,
	data []byte ) (string, []byte, error) {

	files, err := util.ReadFiles( folder, extensions )
	if err != nil {
		return "", nil, err
	}
	file, _ := util.PickFileAtRandom( files )
	fileBytes, err := os.ReadFile( file )
	if err != nil {
		return "", nil, err
	}

	parts := strings.Split( file, "." )
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("Unknown file format.")
	}

	typ := DetermineFileType( parts[len(parts) - 1] )
	switch typ {
	case TextFile:
		data, err = text.Hide( fileBytes, data )
	case ImageFile:
		//util.DebugPrintln("This is an image file:", file)
		data, err = img.Hide( fileBytes, data )
	case AudioFile:
		data, err = audio.Hide( "test", fileBytes, data )
	/*case VideoFile:
		data, err = video.Hide( fileBytes, data )*/
	case DocumentFile:
		data, err = document.Hide( fileBytes, data )
	case ArchiveFile:
		zip, err := archive.FromBytes( fileBytes )
		if err != nil {
			return "", nil, err
		}
		if err = zip.Embed( data ); err != nil {
			return "", nil, err
		}
		res, err := zip.Bytes()
		return file, res, err
	case UnknownFile:
		util.DebugPrintln("[-] Failed to determine file type of", file)
	}
	return file, data, err
}

func RevealFromFile( filename string, data []byte ) ([]byte, error) {
	
	parts := strings.Split( filename, "." )
	if len(parts) < 2 {
		return nil, fmt.Errorf("Unknown file format.")
	}

	typ := DetermineFileType( parts[ len(parts)-1 ] )
	switch typ {
	case TextFile:
		return text.Reveal( data )
	case ImageFile:
		return img.Reveal( data )
	case AudioFile:
		return audio.Reveal( "test", data )
	/*case VideoFile:
		return video.Reveal( data )*/
	case DocumentFile:
		return document.Reveal( data )
	case ArchiveFile:
		zip, err := archive.FromBytes( data )
		if err != nil {
			return nil, err
		}
		return zip.Extract()
	case UnknownFile:
		util.DebugPrintln("[-] Failed to determine file type (2)")
		return data, nil
	}
	return nil, nil
}
