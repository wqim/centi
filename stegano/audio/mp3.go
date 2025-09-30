package audio
import (
	"os"
	"fmt"
	"bytes"
	"encoding/base64"
	id3 "github.com/bogem/id3v2/v2"
	"centi/stegano/util"
)

func HideInMP3( description string, file, data []byte ) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return data, nil
	}

	// embed data in id3v2 tag of mp3 flie...
	tempfile, err := util.CreateTempfile( file )
	if err != nil {
		return nil, err
	}
	defer util.ShredFile( tempfile )
	tag, err := id3.Open( tempfile, id3.Options{ Parse: true } )
	if err != nil {
		return nil, err
	}
	defer tag.Close()

	// just add a comment
	comment := id3.CommentFrame{
		Encoding: id3.EncodingUTF8,
		Language: "eng",
		Description: description,
		Text: base64.StdEncoding.EncodeToString( data ),
	}
	tag.AddCommentFrame( comment )

	// write tag to buffer
	if err = tag.Save(); err != nil {
		return nil, err
	}
	return os.ReadFile( tempfile )
}

func RevealFromMP3( description string, data []byte ) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return data, nil
	}

	// todo: extract data from id3v2 tag of mp3 file...
	tag, err := id3.ParseReader( bytes.NewReader(data), id3.Options{ Parse: true } )
	if err != nil {
		return nil, err
	}
	comments := tag.GetFrames( tag.CommonID("Comments") )
	for _, f := range comments {
		comment, ok := f.(id3.CommentFrame)
		if ok {
			if comment.Description == description {
				return base64.StdEncoding.DecodeString( comment.Text )
			}
		}
	}
	return nil, fmt.Errorf("Failed to find a comment")
}
