package protocol
import (
	"bytes"
	"testing"

	"centi/util"
	"centi/cryptography"
)

func TestInvalidOnePacketPacking(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 20)	// this was fixed with decompression
	key := make([]byte, 32)
	packetSize := uint(4096)
	peer := NewPeer("test")
	peer.SetKey( key )

	mb := NewMsgBuffer( peer, packetSize )
	mb.Push( data )
	packets, err := mb.Next()

	for err == nil {
		util.DebugPrintln("Amount of packets:", len(packets))
		for i, packet := range packets {
			p, err := peer.Unpack( packet, packetSize )
			if err != nil {
				t.Error("Failed to unpack packet #", i, ":", err)
			} else {
				util.DebugPrintln("Data in packet:", p.Body.Data)
				decoded, err := cryptography.DecodeData( p.Body.Data )
				if err != nil {
					t.Error("Failed to decode data:", err)
				} else if bytes.Equal( decoded, data ) == false {
					t.Error("Packing/Unpacking spoiled data:", data, "!=", decoded)
				} else {
					util.DebugPrintln("[+] OK")
				}
			}
		}
		packets, err = mb.Next()
	}
}
