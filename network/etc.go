package network
import (
	"time"
	"centi/util"
	"centi/protocol"
)


func Sleep( c *protocol.ConnManagement ) {
	// sleep for some period of time
	var duration time.Duration
	integer := int64(util.RandInt( 0xffffffff ))
	if c.Config.MinDelay != c.Config.MaxDelay {
		duration = time.Duration( (( integer % int64(c.Config.MinDelay) ) + 
		( int64(c.Config.MaxDelay) - int64(c.Config.MinDelay)))) * time.Millisecond
	} else {
		duration = time.Duration( c.Config.MinDelay ) * time.Millisecond
	}

	time.Sleep( duration )
}
