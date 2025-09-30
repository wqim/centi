package protocol
import (
	"fmt"
	//"strconv"
	"encoding/json"
	"centi/util"
	"centi/cryptography"
)

type Packet struct {
	Head	PacketHead		`json:"h"`
	Body	PacketBody		`json:"b"`
}

type PacketHead struct {
	Typ		uint8		`json:"t"`	// typ of packet, just 1 digit
	Seq		uint64		`json:"s"`	// sequence number (to identify, which packets were sent earlier)
	Total		uint64		`json:"l"`	// total parts of packet
	Compressed	uint8		`json:"c"`	// if the total data in packets is compressed, 0 or 1
}

type PacketBody struct {
	// actual data. very situatively parameter, may be used for both `network` and `application` layers.
	Data		string		`json:"d"`
	OrigSize	uint		`json:"o"`	// idk how else can correctly decode data
	Hmac		string		`json:"h"`	// do the really need this?
}


func PackData( typ, isCompressed uint8, seq, total uint64, data, skey []byte ) ([]byte, error) {
	// generate temporary encryption key
	// skey, err := cryptography.GenRandom( cryptography.SymKeySize )
	//randint := uint64( util.RandInt( 0xffffffff ) )

	hmac := cryptography.HMAC( data, skey ) //cryptography.Hash( append( data, skey... ) )
	strData := cryptography.EncodeData( data )
	packet := Packet{
		PacketHead{
			typ,
			seq,
			total,
			isCompressed,
			//randint,
		},
		PacketBody{
			strData,
			uint(len(data)),
			hmac,
		},
	}

	//util.DebugPrintln("[PackData] Original size of data [packed]:",packet.Body.OrigSize)
	//util.DebugPrintln("Hash put in the packet:", hmac)
	//util.DebugPrintln("Binary data           :")
	//util.DebugPrintln(string(data))

	packed, err := json.Marshal( packet )
	//jsonedDelta := len( packed ) - len(hmac) - len(strData) - len( strconv.Itoa(len(data)) )
	//util.DebugPrintln("jsonedDelta = ", jsonedDelta )
	return packed, err
}

func UnpackData( data, skey []byte ) ([]byte, error) {
	packet, err := UnpackDataToPacket( data, skey )
	if err != nil {
		return nil, err
	}
	// skip the error as it's already handler earlier anyway
	binData, _ := cryptography.DecodeData( packet.Body.Data )
	binData = binData[:packet.Body.OrigSize]
	return binData, nil
}

func UnpackDataToPacket( data, skey []byte ) (*Packet, error) {
	var packet Packet
	if err := json.Unmarshal( data, &packet ); err != nil {
		return nil, err
	}
	if packet.Head.Total < packet.Head.Seq {
		return nil, fmt.Errorf("Invalid packet sequence number.")
	}
	binData, err := cryptography.DecodeData( packet.Body.Data )
	if err != nil {
		return nil, err
	}
	if uint(len(binData)) < packet.Body.OrigSize {
		return nil, fmt.Errorf("Invalid data size(less than specified)")
	}

	//util.DebugPrintln("[UnpackDataToPacket] Original size of data [packed]:",packet.Body.OrigSize)
	util.DebugPrintln("[UnpackDataToPacket] Data:", string( binData ) )

	binData = binData[:packet.Body.OrigSize]
	//hmac := cryptography.//cryptography.Hash( append( binData, skey...) )
	//if hmac != packet.Body.Hmac || hmac == nil || packet.Body.Hmac == nil {
	if cryptography.VerifyHMAC( binData, skey, packet.Body.Hmac ) == false {
	/*util.DebugPrintln("Hash in the packet:", packet.Body.Hmac)
		util.DebugPrintln("Hash real         :", hmac)
		util.DebugPrintln("Binary data       :", binData) */
		return nil, fmt.Errorf("Packet is messed up (2)")
	}
	return &packet, nil
}
