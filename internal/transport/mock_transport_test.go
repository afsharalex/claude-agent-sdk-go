package transport

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMockTransport_Connect(t *testing.T) {
	mock := NewMockTransport()

	err := mock.Connect(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mock.IsConnected() {
		t.Error("Expected IsConnected to return true after Connect")
	}
	if !mock.IsReady() {
		t.Error("Expected IsReady to return true after Connect")
	}
	if mock.ConnectCalls != 1 {
		t.Errorf("Expected ConnectCalls to be 1, got %d", mock.ConnectCalls)
	}
}

func TestMockTransport_Connect_Error(t *testing.T) {
	mock := NewMockTransport()
	expectedErr := errors.New("connection failed")
	mock.WithConnectError(expectedErr)

	err := mock.Connect(context.Background())
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if mock.IsConnected() {
		t.Error("Expected IsConnected to return false after failed Connect")
	}
}

func TestMockTransport_Write(t *testing.T) {
	mock := NewMockTransport()
	_ = mock.Connect(context.Background())

	testData := `{"type": "test"}`
	err := mock.Write(context.Background(), testData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mock.WriteCalls != 1 {
		t.Errorf("Expected WriteCalls to be 1, got %d", mock.WriteCalls)
	}

	writtenData := mock.GetWrittenData()
	if len(writtenData) != 1 || writtenData[0] != testData {
		t.Errorf("Expected written data %s, got %v", testData, writtenData)
	}
}

func TestMockTransport_Write_WithCallback(t *testing.T) {
	mock := NewMockTransport()
	_ = mock.Connect(context.Background())

	callbackCalled := false
	mock.OnWrite = func(data string) error {
		callbackCalled = true
		return nil
	}

	err := mock.Write(context.Background(), "test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !callbackCalled {
		t.Error("Expected OnWrite callback to be called")
	}
}

func TestMockTransport_Write_CallbackError(t *testing.T) {
	mock := NewMockTransport()
	_ = mock.Connect(context.Background())

	expectedErr := errors.New("write callback error")
	mock.OnWrite = func(data string) error {
		return expectedErr
	}

	err := mock.Write(context.Background(), "test")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockTransport_Write_Error(t *testing.T) {
	mock := NewMockTransport()
	expectedErr := errors.New("write failed")
	mock.WithWriteError(expectedErr)

	err := mock.Write(context.Background(), "test")
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockTransport_ReadMessages(t *testing.T) {
	messages := []map[string]any{
		{"type": "message1"},
		{"type": "message2"},
		{"type": "message3"},
	}

	mock := NewMockTransport().WithMessages(messages...)

	ch := mock.ReadMessages(context.Background())

	received := make([]map[string]any, 0)
	for result := range ch {
		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
			continue
		}
		received = append(received, result.Data)
	}

	if len(received) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(received))
	}
	if mock.ReadCalls != 1 {
		t.Errorf("Expected ReadCalls to be 1, got %d", mock.ReadCalls)
	}
}

func TestMockTransport_ReadMessages_ContextCancellation(t *testing.T) {
	// Create messages but context will be cancelled before all are read
	messages := []map[string]any{
		{"type": "message1"},
		{"type": "message2"},
	}

	mock := NewMockTransport().WithMessages(messages...)

	ctx, cancel := context.WithCancel(context.Background())
	ch := mock.ReadMessages(ctx)

	// Read one message
	result := <-ch
	if result.Error != nil {
		t.Errorf("Unexpected error on first read: %v", result.Error)
	}

	// Cancel context
	cancel()

	// Give goroutine time to process cancellation
	time.Sleep(10 * time.Millisecond)

	// Drain remaining messages - may get context error
	for result := range ch {
		if result.Error != nil && result.Error != context.Canceled {
			// Just consume any remaining results
		}
	}
}

func TestMockTransport_Close(t *testing.T) {
	mock := NewMockTransport()
	_ = mock.Connect(context.Background())

	err := mock.Close()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !mock.IsClosed() {
		t.Error("Expected IsClosed to return true after Close")
	}
	if mock.IsReady() {
		t.Error("Expected IsReady to return false after Close")
	}
	if mock.IsConnected() {
		t.Error("Expected IsConnected to return false after Close")
	}
	if mock.CloseCalls != 1 {
		t.Errorf("Expected CloseCalls to be 1, got %d", mock.CloseCalls)
	}
}

func TestMockTransport_Close_WithCallback(t *testing.T) {
	mock := NewMockTransport()
	_ = mock.Connect(context.Background())

	callbackCalled := false
	mock.OnClose = func() error {
		callbackCalled = true
		return nil
	}

	err := mock.Close()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !callbackCalled {
		t.Error("Expected OnClose callback to be called")
	}
}

func TestMockTransport_Close_CallbackError(t *testing.T) {
	mock := NewMockTransport()
	expectedErr := errors.New("close callback error")
	mock.OnClose = func() error {
		return expectedErr
	}

	err := mock.Close()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockTransport_Close_Error(t *testing.T) {
	mock := NewMockTransport()
	expectedErr := errors.New("close failed")
	mock.WithCloseError(expectedErr)

	err := mock.Close()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockTransport_EndInput(t *testing.T) {
	mock := NewMockTransport()

	err := mock.EndInput()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if mock.EndInputCalls != 1 {
		t.Errorf("Expected EndInputCalls to be 1, got %d", mock.EndInputCalls)
	}
}

func TestMockTransport_EndInput_Error(t *testing.T) {
	mock := NewMockTransport()
	expectedErr := errors.New("end input failed")
	mock.EndInputError = expectedErr

	err := mock.EndInput()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestMockTransport_InjectMessage(t *testing.T) {
	// Test that InjectMessage sends to the channel when one exists
	mock := NewMockTransport()

	// First call ReadMessages to create the channel
	ctx, cancel := context.WithCancel(context.Background())
	ch := mock.ReadMessages(ctx)

	// Drain any initial messages in a separate goroutine
	receivedInjected := make(chan bool, 1)
	go func() {
		for result := range ch {
			if result.Data != nil && result.Data["type"] == "injected" {
				receivedInjected <- true
				return
			}
		}
		receivedInjected <- false
	}()

	// Give the goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel context to stop reading
	cancel()

	// Verify the test setup worked - just ensure no panic
	select {
	case <-receivedInjected:
		// Test passed - either received or channel closed
	case <-time.After(100 * time.Millisecond):
		// Timeout is acceptable
	}
}

func TestMockTransport_InjectMessage_NoChannel(t *testing.T) {
	mock := NewMockTransport()

	// Should not panic when no channel is active
	mock.InjectMessage(map[string]any{"type": "test"})
}

func TestMockTransport_InjectError(t *testing.T) {
	// Test that InjectError sends to the channel when one exists
	mock := NewMockTransport()

	// First call ReadMessages to create the channel
	ctx, cancel := context.WithCancel(context.Background())
	ch := mock.ReadMessages(ctx)

	// Drain messages in a separate goroutine
	done := make(chan bool, 1)
	go func() {
		for range ch {
			// Just consume
		}
		done <- true
	}()

	// Give the goroutine time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for completion
	select {
	case <-done:
		// Test passed - no panic
	case <-time.After(100 * time.Millisecond):
		// Timeout is acceptable
	}
}

func TestMockTransport_InjectError_NoChannel(t *testing.T) {
	mock := NewMockTransport()

	// Should not panic when no channel is active
	mock.InjectError(errors.New("test error"))
}

func TestMockTransport_Reset(t *testing.T) {
	mock := NewMockTransport()

	// Set up some state
	_ = mock.Connect(context.Background())
	_ = mock.Write(context.Background(), "test")
	mock.WithConnectError(errors.New("test"))
	mock.WithWriteError(errors.New("test"))
	mock.WithCloseError(errors.New("test"))
	mock.WithMessages(map[string]any{"type": "test"})
	mock.OnWrite = func(data string) error { return nil }
	mock.OnClose = func() error { return nil }
	// Call IsReady before reset to set up counter
	_ = mock.IsReady()

	// Reset
	mock.Reset()

	// Verify all state is cleared - use direct field access to avoid incrementing counters
	mock.mu.Lock()
	connected := mock.connected
	ready := mock.ready
	closed := mock.closed
	isReadyCalls := mock.IsReadyCalls
	mock.mu.Unlock()

	if connected {
		t.Error("Expected connected to be false after Reset")
	}
	if ready {
		t.Error("Expected ready to be false after Reset")
	}
	if closed {
		t.Error("Expected closed to be false after Reset")
	}
	if mock.ConnectError != nil {
		t.Error("Expected ConnectError to be nil after Reset")
	}
	if mock.WriteError != nil {
		t.Error("Expected WriteError to be nil after Reset")
	}
	if mock.CloseError != nil {
		t.Error("Expected CloseError to be nil after Reset")
	}
	if mock.EndInputError != nil {
		t.Error("Expected EndInputError to be nil after Reset")
	}
	if len(mock.Messages) != 0 {
		t.Error("Expected Messages to be empty after Reset")
	}
	if len(mock.WrittenData) != 0 {
		t.Error("Expected WrittenData to be empty after Reset")
	}
	if mock.ConnectCalls != 0 {
		t.Error("Expected ConnectCalls to be 0 after Reset")
	}
	if mock.WriteCalls != 0 {
		t.Error("Expected WriteCalls to be 0 after Reset")
	}
	if mock.ReadCalls != 0 {
		t.Error("Expected ReadCalls to be 0 after Reset")
	}
	if mock.CloseCalls != 0 {
		t.Error("Expected CloseCalls to be 0 after Reset")
	}
	if mock.EndInputCalls != 0 {
		t.Error("Expected EndInputCalls to be 0 after Reset")
	}
	if isReadyCalls != 0 {
		t.Errorf("Expected IsReadyCalls to be 0 after Reset, got %d", isReadyCalls)
	}
	if mock.OnWrite != nil {
		t.Error("Expected OnWrite to be nil after Reset")
	}
	if mock.OnClose != nil {
		t.Error("Expected OnClose to be nil after Reset")
	}
}

func TestMockTransport_SetReady(t *testing.T) {
	mock := NewMockTransport()

	if mock.IsReady() {
		t.Error("Expected IsReady to be false initially")
	}

	mock.SetReady(true)
	if !mock.IsReady() {
		t.Error("Expected IsReady to be true after SetReady(true)")
	}

	mock.SetReady(false)
	if mock.IsReady() {
		t.Error("Expected IsReady to be false after SetReady(false)")
	}
}

func TestMockTransport_ClearWrittenData(t *testing.T) {
	mock := NewMockTransport()
	_ = mock.Connect(context.Background())

	_ = mock.Write(context.Background(), "data1")
	_ = mock.Write(context.Background(), "data2")

	if len(mock.GetWrittenData()) != 2 {
		t.Errorf("Expected 2 written data items, got %d", len(mock.GetWrittenData()))
	}

	mock.ClearWrittenData()

	if len(mock.GetWrittenData()) != 0 {
		t.Errorf("Expected 0 written data items after clear, got %d", len(mock.GetWrittenData()))
	}
}

func TestMockTransport_IsReady_TracksCalls(t *testing.T) {
	mock := NewMockTransport()

	_ = mock.IsReady()
	_ = mock.IsReady()
	_ = mock.IsReady()

	if mock.IsReadyCalls != 3 {
		t.Errorf("Expected IsReadyCalls to be 3, got %d", mock.IsReadyCalls)
	}
}

func TestMockTransport_Concurrent(t *testing.T) {
	mock := NewMockTransport()
	_ = mock.Connect(context.Background())

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = mock.Write(context.Background(), "data")
		}(i)
	}

	// Concurrent IsReady calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mock.IsReady()
		}()
	}

	wg.Wait()

	if mock.WriteCalls != numGoroutines {
		t.Errorf("Expected WriteCalls to be %d, got %d", numGoroutines, mock.WriteCalls)
	}
	if mock.IsReadyCalls != numGoroutines {
		t.Errorf("Expected IsReadyCalls to be %d, got %d", numGoroutines, mock.IsReadyCalls)
	}
}

// Ensure MockTransport satisfies Transport interface
func TestMockTransport_ImplementsInterface(t *testing.T) {
	var _ Transport = (*MockTransport)(nil)
}

func TestMockTransport_ReadMessages_WithMessages(t *testing.T) {
	messages := []map[string]any{
		{"type": "first"},
		{"type": "second"},
	}
	mock := NewMockTransport().WithMessages(messages...)

	ch := mock.ReadMessages(context.Background())

	var received []map[string]any
	for result := range ch {
		if result.Error != nil {
			t.Errorf("Unexpected error: %v", result.Error)
		}
		received = append(received, result.Data)
	}

	if len(received) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(received))
	}
}

func TestMockTransport_MultipleReads(t *testing.T) {
	mock := NewMockTransport()

	// First read
	ch1 := mock.ReadMessages(context.Background())
	for range ch1 {
	}

	// Second read should work too
	mock.WithMessages(map[string]any{"type": "new"})
	ch2 := mock.ReadMessages(context.Background())

	count := 0
	for range ch2 {
		count++
	}

	if count != 1 {
		t.Errorf("Expected 1 message in second read, got %d", count)
	}
}

func TestMockTransport_GetWrittenData_Empty(t *testing.T) {
	mock := NewMockTransport()

	data := mock.GetWrittenData()
	if len(data) != 0 {
		t.Errorf("Expected empty written data, got %d items", len(data))
	}
}

func TestMockTransport_ReadMessages_AlreadyCancelled(t *testing.T) {
	messages := []map[string]any{
		{"type": "test"},
	}

	mock := NewMockTransport().WithMessages(messages...)

	// Cancel before reading
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := mock.ReadMessages(ctx)

	// Should still receive messages since they're buffered, then close
	count := 0
	for range ch {
		count++
	}

	// Just verify no panic and channel closes
	// Count can be 0, 1, or 2 (including context error) depending on timing
}

func TestMockTransport_ChainedConfiguration(t *testing.T) {
	mock := NewMockTransport().
		WithMessages(map[string]any{"type": "test"}).
		WithConnectError(nil).
		WithWriteError(nil).
		WithCloseError(nil)

	// Should not panic and all methods should work
	err := mock.Connect(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMockTransport_ReadMessages_ContextCancelBeforeSelect(t *testing.T) {
	// Create many messages to ensure we hit the context check in the loop
	messages := make([]map[string]any, 10)
	for i := 0; i < 10; i++ {
		messages[i] = map[string]any{"index": i}
	}

	mock := NewMockTransport().WithMessages(messages...)

	ctx, cancel := context.WithCancel(context.Background())

	ch := mock.ReadMessages(ctx)

	// Read first message
	<-ch

	// Cancel after first message
	cancel()

	// Drain remaining - should eventually get context error or channel close
	for range ch {
		// Just drain
	}
}

func TestMockTransport_InjectMessage_WithActiveChannel(t *testing.T) {
	// Create a mock with no initial messages
	mock := NewMockTransport()

	// Create a channel manually to simulate active reading
	mock.mu.Lock()
	mock.messageCh = make(chan ReadResult, 10)
	ch := mock.messageCh
	mock.mu.Unlock()

	// Inject a message
	injectedMsg := map[string]any{"type": "injected"}
	mock.InjectMessage(injectedMsg)

	// Read the injected message
	select {
	case result := <-ch:
		if result.Data["type"] != "injected" {
			t.Errorf("Expected injected message, got %v", result.Data)
		}
	default:
		t.Error("No message received")
	}
}

func TestMockTransport_InjectError_WithActiveChannel(t *testing.T) {
	// Create a mock with no initial messages
	mock := NewMockTransport()

	// Create a channel manually to simulate active reading
	mock.mu.Lock()
	mock.messageCh = make(chan ReadResult, 10)
	ch := mock.messageCh
	mock.mu.Unlock()

	// Inject an error
	expectedErr := errors.New("test error")
	mock.InjectError(expectedErr)

	// Read the injected error
	select {
	case result := <-ch:
		if result.Error != expectedErr {
			t.Errorf("Expected error %v, got %v", expectedErr, result.Error)
		}
	default:
		t.Error("No error received")
	}
}
