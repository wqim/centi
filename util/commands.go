package util
import (
	"os"
	"fmt"
	"os/exec"
	"strings"
	"strconv"
	"path/filepath"
	"centi/cryptography"
)

const (
	TextEditor = "/usr/bin/vi"
	TextEditorVariableName = "CENTI_EDITOR"
	ShredCount = 10
)

/*
 * user-related functions which are required 
 * in order not to fuck up user with constant decrypt-edit-encrypt things.
 */
func EditConfig( config, saltFile string, password []byte ) error {
	// decrypt config, put it into temporary file, edit,
	// read, shred temporary file and put encrypted configuration
	// back.
	te := TextEditor	// setup default text editor
	environments := os.Environ()
	for _, variable := range environments {
		parts := strings.Split( variable, "=" )
		if len(parts) == 2 { // it is strange, if not
			if parts[0] == TextEditorVariableName {
				te = parts[1]
				break
			}
		}
	}

	// ok, text editor found, decrypt file
	data, err := os.ReadFile( config )
	if err != nil {
		return fmt.Errorf("Failed to read configuration: %s", err.Error() )
	}

	saltBytes, err := os.ReadFile( saltFile )
	if err != nil {
		return fmt.Errorf("Failed to read file with salt: %s", err.Error() )
	}

	key := cryptography.DeriveKey( password, saltBytes )
	pt, err := cryptography.Decrypt( data, key )
	if err != nil {
		return fmt.Errorf("Failed to decrypt configuration: %s", err.Error() + "; Invalid password?")
	}

	// write it into temporary file
	tempFile := filepath.Join( os.TempDir(), fmt.Sprintf("tmp-%d", RandInt( 10000 ) ) )
	if err = os.WriteFile( tempFile, pt, 0660 ); err != nil {
		return fmt.Errorf("Failed to write into temporary file: %s", err.Error())
	}

	defer ShredFile( tempFile )	// not to forget to securely delete file
	//fmt.Println("Temporary file:", tempFile)
	cmd := exec.Command( te, tempFile )
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("Failed to edit file using %v: %s", te, err.Error())
	}
	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("Failed to edit file (2) using %v: %s", te, err.Error())
	}

	// file was edited fine
	pt, err = os.ReadFile( tempFile )
	if err != nil {
		return fmt.Errorf("Failed to read temporary file: %s", err.Error())
	}

	// encrypt file and put the ciphertext back
	data, err = cryptography.Encrypt( pt, key )
	if err != nil {
		return err
	}

	return os.WriteFile( config, data, 0660 )
}

func ReadLog( log, saltFile string, password []byte ) error {
	// just read the logs and print it to the screen
	saltBytes, err := os.ReadFile( saltFile )
	if err != nil {
		return fmt.Errorf("Failed to read file with salt: %s", err.Error() )
	}

	key := cryptography.DeriveKey( password, saltBytes )
	data, err := os.ReadFile( log )
	if err != nil {
		return fmt.Errorf("Failed to read file: %s", err.Error())
	}
	logs, err := cryptography.Decrypt( data, key )
	if err != nil {
		// logs are unencrypted?
		// checking for plaintext
		strLogs := string(data)
		for _, run := range strLogs {
			if strconv.IsPrint( run ) == false {
				return fmt.Errorf("Failed to decrypt logs: invalid password.")
			}
		}
		// logs are unencrypted
		fmt.Println( strLogs )
		return nil
	}
	fmt.Println( string(logs) )
	return nil
}

// some auxilary things here
func ShredFile( filename string ) error {
	info, err := os.Stat( filename )
	if err != nil {
		// something really bad
		return err
	}
	var finalError error
	if info.Size() > 0 {
		for i := 0; i < ShredCount; i++ {
			content, err := cryptography.GenRandom( uint(info.Size()) )
			if err == nil {
				os.WriteFile( filename, content, 0660 )
			} else {
				finalError = err
			}
		}
	}
	if err = os.Remove( filename ); err != nil {
		finalError = err
	}
	return finalError
}
