package util
import (
	"log"
)

const (
	DebugMode = true
)


func DebugPrintln( args ...any ) {
	if DebugMode == true {
		log.Println( args... )
	}
}

func DebugPrintf( format string, args ...any ) {
	if DebugMode == true {
		log.Printf( format, args... )
	}
}
