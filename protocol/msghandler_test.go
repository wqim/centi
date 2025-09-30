package protocol

import (
	"fmt"
	"sync"
	"bytes"
	"testing"
)

// TestNewMsgChannel verifies creation of a new MsgChannel.
func TestNewMsgChannel(t *testing.T) {
	ch := NewMsgChannel("peer1", 3, 1)
	if ch.peerAlias != "peer1" {
		t.Errorf("Expected peerAlias 'peer1', got '%s'", ch.peerAlias)
	}
	if ch.total != 3 {
		t.Errorf("Expected total 3, got %d", ch.total)
	}
	if len(ch.messages) != 3 {
		t.Errorf("Expected messages length 3, got %d", len(ch.messages))
	}
	if ch.compressed != 1 {
		t.Errorf("Expected compressed 1, got %d", ch.compressed)
	}
}

// TestPushAndIsFull tests pushing messages and checking if channel is full.
func TestPushAndIsFull(t *testing.T) {
	ch := NewMsgChannel("peer2", 2, 0)

	// Initially, should not be full
	if ch.IsFull() {
		t.Error("Channel should not be full initially")
	}

	// Push first message
	ch.Push([]byte("msg1"), 0)
	if ch.IsFull() {
		t.Error("Channel should not be full after one message")
	}

	// Push second message
	ch.Push([]byte("msg2"), 1)
	if !ch.IsFull() {
		t.Error("Channel should be full after all messages pushed")
	}
}

// TestData retrieves combined data from the channel.
func TestData(t *testing.T) {
	ch := NewMsgChannel("peer3", 2, 0)
	msg1 := []byte("hello")
	msg2 := []byte("world")
	ch.Push(msg1, 0)
	ch.Push(msg2, 1)

	data, err := ch.Data()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := append(msg1, msg2...)
	if !bytes.Equal(data, expected) {
		t.Errorf("Data mismatch. Got: %s, Expected: %s", data, expected)
	}
}

// TestDataWithIncompleteMessages tests Data() error when messages are incomplete.
func TestDataWithIncompleteMessages(t *testing.T) {
	ch := NewMsgChannel("peer4", 2, 0)
	ch.Push([]byte("only one"), 0)

	data, err := ch.Data()
	if err == nil {
		t.Error("Expected error for incomplete message, got nil")
	}
	if data != nil {
		t.Errorf("Expected nil data, got: %v", data)
	}
}

// TestMsgHandler_AddPacket_AddsChannelAndMessages tests adding packets and retrieving data.
func TestMsgHandler_AddPacket_AddsChannelAndMessages(t *testing.T) {
	handler := NewMsgHandler()
	
	// Add packets for a new peer
	handler.AddPacket("peer5", 0, 2, 0, []byte("part1"))
	handler.AddPacket("peer5", 1, 2, 0, []byte("part2"))
	
	// The channel should be created and full
	data := handler.ByAlias("peer5")
	if data == nil {
		t.Fatal("Expected data for peer5, got nil")
	}
	expected := append([]byte("part1"), []byte("part2")...)
	if !bytes.Equal(data, expected) {
		t.Errorf("Data mismatch. Got: %s, Expected: %s", data, expected)
	}

	// After retrieval, the channel should be removed
	if handler.exists("peer5", false) {
		t.Error("Channel for peer5 should be removed after data retrieval")
	}
}

// TestMsgHandler_ByAlias_ReturnsNilWhenNoData tests returning nil for incomplete or missing channels.
func TestMsgHandler_ByAlias_ReturnsNilWhenNoData(t *testing.T) {
	handler := NewMsgHandler()

	// No channels added yet
	result := handler.ByAlias("nonexistent")
	if result != nil {
		t.Error("Expected nil for nonexistent alias")
	}

	// Add a channel with incomplete data
	handler.AddChannel("peer6", 2, 0)
	result = handler.ByAlias("peer6")
	if result != nil {
		t.Error("Expected nil for incomplete data")
	}
}

// TestConcurrentAccess tests thread safety of MsgHandler.
func TestConcurrentAccess(t *testing.T) {
	handler := NewMsgHandler()

	// Add channels concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			alias := fmt.Sprintf("peer%d", i)
			handler.AddChannel(alias, 1, 0)
			handler.AddPacket(alias, 0, 1, 0, []byte(fmt.Sprintf("msg%d", i)))
		}(i)
	}
	wg.Wait()

	// Check that all channels contain the expected data
	for i := 0; i < 10; i++ {
		alias := fmt.Sprintf("peer%d", i)
		data := handler.ByAlias(alias)
		expected := []byte(fmt.Sprintf("msg%d", i))
		if !bytes.Equal(data, expected) {
			t.Errorf("Data mismatch for %s: got %s, expected %s", alias, data, expected)
		}
	}
}
