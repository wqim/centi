package util
import (
	"fmt"
	"syscall"
	"golang.org/x/term"
)

// just a wrapper for term...
func GetPasswd( prompt string ) ([]byte, error) {
	fmt.Print( prompt )
	bytepw, err := term.ReadPassword( int(syscall.Stdin) )
	fmt.Println()
	return bytepw, err
}
