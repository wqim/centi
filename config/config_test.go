package config
import (
	"testing"
	"centi/util"
)

func TestSaveAndLoadConfig( t *testing.T ) {
	conf := FullConfig{
		NetworkConfig{},
		ServerConfiguration{},
		SteganoConfig{},
		util.LoggerInfo{},
		[]ConnectionInfo{},
		"test.db",
		"test-password",
		10000,
		KeysConfig{},
	}
	key := make([]byte, 32)// a dummy key
	filename := "/tmp/centi-test-config.enc"
	if err := SaveConfig(filename, key, &conf); err != nil {
		t.Errorf("Failed to save configuration: %s", err.Error())
	}
	conf2, err := LoadConfig( filename, key )
	if err != nil {
		t.Errorf("Failed to load configuration: %s", err.Error())
	}
	// use only some parameters as if encryption was fine, everything will be equal anyway
	if conf.DbFile != conf2.DbFile || conf.DbPassword != conf2.DbPassword {
		t.Errorf("[CRITICAL] Configuration was changed during encryption/decryption process")
	}
}

func TestExtractNetworkSubkeys( t *testing.T ) {
	subkeys := map[string]string{
		"test": "test",
	}
	nSubkeys := ExtractNetworkSubkeys( subkeys )
	if len(nSubkeys) != len(subkeys) {
		t.Errorf("Invalid amount of extracted keys: %d != %d", len(subkeys), len(nSubkeys) )
	}
	// todo: check result of the test
}
