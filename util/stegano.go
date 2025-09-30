package util
import (
	"os"
	"strings"
	"path/filepath"
)

func PickFileAtRandom( files []string ) (string, []string) {
	idx := RandInt( len(files) )
	file := files[idx]
	files = append( files[:idx], files[idx+1:]... )
	return file, files
}

func ReadFiles( folder string, supportedExtensions []string ) ([]string, error) {
	allFiles, err := os.ReadDir( folder )
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, f := range allFiles {
		for _, ext := range supportedExtensions {
			if strings.HasSuffix( f.Name(), "." + ext ) == true {
				result = append( result, filepath.Join( folder, f.Name() ) )
			}
		}
	}
	return result, nil
}
