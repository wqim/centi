package util
import (
	"sync"
	"centi/cryptography"
)

type Storage struct {
	storage map[string][]string
	mtx	sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{
		map[string][]string{},
		sync.Mutex{},
	}
}

// this function hashes the content we are storing in `name`,
// whatever it matters.
func(s *Storage) Add( name string, content []byte ) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	hash := cryptography.Hash( content )
	if hash != "" {
		if len(s.storage[hash]) == 0 {
			s.storage[hash] = []string{name}
		} else {
			s.storage[hash] = append( s.storage[hash], name )
		}
	}
}

func(s *Storage) Find( content []byte ) []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	hash := cryptography.Hash( content )
	return s.storage[hash]
}

func(s *Storage) Remove( content []byte ) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	hash := cryptography.Hash( content )
	delete( s.storage, hash )
}
