package protocol
import (
	"fmt"
	"testing"
	"centi/cryptography"
)

func TestNewMsgBuffer(t *testing.T) {
	// Create a mock Peer

	randomBytes, _ := cryptography.GenRandom( 10000 )

	mockPeer := NewPeer("")
	mockPeer.SetKey( make([]byte, 32) )

	// Test cases with different packet sizes
	testCases := []struct {
		name		 string
		packetSize   uint
		data		[]byte
		expectedMsgSize uint
		expectedError error
		expectedPacketsCount int
	}{
		{
			name:		 "Valid packet size",
			packetSize:   2048,
			data:	make([]byte, 32),
			expectedMsgSize: 2048,
			expectedError: nil,
			expectedPacketsCount: 1,
		},
		{
			name:		 "Zero packet size",
			packetSize:   0,
			data:	make([]byte, 32),
			expectedMsgSize: 0,
			expectedError: fmt.Errorf("invalid packet size"),
			expectedPacketsCount: 0,
		},
		{
			name:		"Suspicious 106 bytes message.",
			packetSize:	4096,
			data:	make([]byte, 106),
			expectedMsgSize: 4096,
			expectedError: nil,
			expectedPacketsCount: 1,
		},
		{
			name:		"Very big packet",
			packetSize:	4096,
			data:	randomBytes,
			expectedMsgSize: 4096,
			expectedError: nil,
			expectedPacketsCount: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mb := NewMsgBuffer(mockPeer, tc.packetSize)
			if mb == nil {
				t.Errorf("NewMsgBuffer returned nil for valid input")
			}
			
			mb.Push( tc.data )
			
			data, err := mb.Next()
			if err != nil && tc.expectedError == nil {
				t.Errorf("Failed to pack data: %v", err)
			}

			if len(data) != tc.expectedPacketsCount {
				t.Errorf("Expected packets count %d, got %d.", tc.expectedPacketsCount, len(data) )
			}
			if tc.expectedError != nil {
				if mb.msgSize != 0 {
					t.Errorf("Expected error %v, got valid msgSize", tc.expectedError)
				}
			}
		})
	}
}
