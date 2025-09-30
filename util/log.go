package util
import (
	"os"
	"sync"
	"time"
	//"strings"
	"centi/cryptography"
)

/*
 * a custom logger, loader of configuration and other stuff.
 */
const (
	Error = 1
	Warning = 2
	Info = 4

	RedColor = "\033[31m"
	YellowColor = "\033[33m"
	GreenColor = "\033[32m"
	CyanColor = "\033[36m"
	BlueColor = "\033[34m"
	MagentaColor = "\033[35m"
	ResetColor = "\033[0m"
)

type LoggerInfo struct {
	Filename	string		`json:"filename"`
	Password	string		`json:"password"`
	IsEncrypted	bool		`json:"is_encrypted"`
	IsColored	bool		`json:"is_colored"`
	SaveTime	bool		`json:"save_time"`
	Mode		uint8		`json:"mode"`
}

type Logger struct {
	li		*LoggerInfo
	mtx		sync.Mutex
}

func NewLogger( li *LoggerInfo ) *Logger {
	return &Logger{
		li,
		sync.Mutex{},
	}
}

func(l *Logger) colorize( line string, color string ) string {
	if l.li.IsColored {
		return color + line + ResetColor
	}
	return line
}

func(l *Logger) prepareString( str string, clr string ) string {
	toWrite := l.colorize( str, clr ) + " "
	if l.li.SaveTime {
		toWrite += time.Now().String() + " "
	}
	return toWrite
}

func(l *Logger) LogString( s string ) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if l.li.IsEncrypted == false {
		// just append line
		f, err := os.OpenFile( l.li.Filename, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0600 )
		if err == nil {
			defer f.Close()
			f.WriteString( s + "\n" )
		}
	} else {
		pass, saltBytes, err := cryptography.SplitWithSalt( l.li.Password )
		if err == nil {
			key := cryptography.DeriveKey( pass, saltBytes )
			data, err := os.ReadFile( l.li.Filename )
			if err == nil {
				currentLog, err := cryptography.Decrypt( data, key )
				if err == nil {
					newData := append( currentLog, []byte(s)... )
					newData, err = cryptography.Encrypt( newData, key )
					if err == nil {
						os.WriteFile( l.li.Filename, newData, 0660 )
					}
				}
			}
		}
	}
}

func(l *Logger) LogError(err error) {
	if l.li.Mode & Error == Error {
		toWrite := l.prepareString("[ERROR]", RedColor) + err.Error()
		l.LogString( toWrite )
	}
}

func(l *Logger) LogWarning( warning string ) {
	if l.li.Mode & Warning == Warning {
		toWrite := l.prepareString("[WARNING]", YellowColor) + warning
		l.LogString( toWrite )
	}
}


func(l *Logger) LogInfo( info string ) {
	if l.li.Mode & Info == Info {
		toWrite := l.prepareString( "[INFO]", CyanColor ) + info
		l.LogString( toWrite )
	}
}
