package protocol

const (
	
	PkPct = uint8(0)		// packet with retransmitted public key of known peer
	DataPct = uint8(1)		// packet with actual data for higher layers of network (emails, messengers, etc.)
	RetransmitPct = uint8(2)	// packet which should be unpacked and resent over the network
	PkReqPct = uint8(3)		// packet storing public keys request(?)

	// border values for validation of packets
	MinPct = PkPct
	MaxPct = RetransmitPct

	// ways of public key distribution
	// that's a pity Go does not have enumerations...
	DistributePkViaProfilePicture = uint8(0)
	
	DistributePkMin = DistributePkViaProfilePicture
	DistributePkMax = DistributePkViaProfilePicture

	MinPacketSize = 2048
)
