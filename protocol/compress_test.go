package protocol

import (
	"bytes"
	"testing"
	//"fmt"
	"centi/cryptography"
)

func TestCompress(t *testing.T) {
	randbytes, _ := cryptography.GenRandom(128)
	// Test cases with various data sizes and contents
	testCases := []struct {
		name	string
		data	[]byte
		expectedCompressionStatus uint8
		expectedCompressed []byte
		expectedError error
	}{
		{
			name: "Empty data",
			data: []byte{},
			expectedCompressionStatus: 0,
			expectedCompressed: []byte{},
			expectedError: nil,
		},
		{
			name: "Small data",
			data: bytes.Repeat([]byte("a"), 150),
			expectedCompressionStatus: 1,
			expectedCompressed: []byte{}, // Will be compressed
			expectedError: nil,
		},
		{
			name: "Large data",
			data: bytes.Repeat([]byte("a"), 1024), // Large data for compression
			expectedCompressionStatus: 1,
			expectedCompressed: []byte{}, // Will be compressed
			expectedError: nil,
		},
		{
			name: "Data not compressible",
			data: []byte{0x01, 0x02, 0x03, 0x04},
			expectedCompressionStatus: 0,
			expectedCompressed: []byte{}, // Will be compressed
			expectedError: nil,
		},
		{
			name: "Large data, no compression",
			data:  randbytes, // Small data, unlikely to compress
			expectedCompressionStatus: 0,
			expectedCompressed: []byte{},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualCompressionStatus, actualCompressed, actualError := Compress(tc.data)

			// Check for errors
			if actualError != nil && tc.expectedError == nil {
				t.Errorf("Unexpected error: %v", actualError)
			}
			if actualError == nil && tc.expectedError != nil {
				t.Errorf("Expected error %v, but got none", tc.expectedError)
			}

			// Check compression status
			if actualCompressionStatus != tc.expectedCompressionStatus {
				t.Errorf("Expected compression status %d, got %d",
				tc.expectedCompressionStatus, actualCompressionStatus)
			}

			// Check compressed data (only if no error and compression expected)
			if actualError == nil && tc.expectedCompressionStatus == 1 && len(tc.expectedCompressed) == 0 {
				if len(actualCompressed) == 0 {
					t.Logf("Compressed data: %v", actualCompressed)
				}
				if len(actualCompressed) > 0 && len(actualCompressed) >= len(tc.data) {
					t.Errorf("Compressed data length is not smaller than original data")
				}
			}

			if actualCompressionStatus == 1 {
				// decompress data
				decompressed, err := Decompress( actualCompressed )
				if err != nil {
					t.Errorf("Failed to decompress: %s, %v", err.Error(), actualCompressed)
				} else if bytes.Equal( decompressed, tc.data ) == false {
					t.Errorf("Compress/decompress breaks the data. Original: %v; Decompressed: %v",
						tc.data, decompressed)
				}
			}
		})
	}
}
