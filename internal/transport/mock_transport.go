package transport

import (
	"context"
	"sync"
)

// MockTransport is a mock implementation of the Transport interface for testing.
type MockTransport struct {
	mu sync.Mutex

	// Connection state
	connected bool
	ready     bool
	closed    bool

	// Configurable behaviors
	ConnectError  error
	WriteError    error
	CloseError    error
	EndInputError error

	// Injected messages to return from ReadMessages
	Messages []map[string]any

	// Captured data written to transport
	WrittenData []string

	// Internal channel for message delivery
	messageCh chan ReadResult

	// Track method calls for assertions
	ConnectCalls   int
	WriteCalls     int
	ReadCalls      int
	CloseCalls     int
	EndInputCalls  int
	IsReadyCalls   int

	// OnWrite callback for custom write handling
	OnWrite func(data string) error

	// OnClose callback for custom close handling
	OnClose func() error
}

// NewMockTransport creates a new MockTransport with default settings.
func NewMockTransport() *MockTransport {
	return &MockTransport{
		Messages:    make([]map[string]any, 0),
		WrittenData: make([]string, 0),
	}
}

// WithMessages sets the messages to be returned by ReadMessages.
func (m *MockTransport) WithMessages(msgs ...map[string]any) *MockTransport {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = msgs
	return m
}

// WithConnectError sets an error to be returned by Connect.
func (m *MockTransport) WithConnectError(err error) *MockTransport {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConnectError = err
	return m
}

// WithWriteError sets an error to be returned by Write.
func (m *MockTransport) WithWriteError(err error) *MockTransport {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WriteError = err
	return m
}

// WithCloseError sets an error to be returned by Close.
func (m *MockTransport) WithCloseError(err error) *MockTransport {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CloseError = err
	return m
}

// Connect implements Transport.Connect.
func (m *MockTransport) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConnectCalls++

	if m.ConnectError != nil {
		return m.ConnectError
	}

	m.connected = true
	m.ready = true
	return nil
}

// Write implements Transport.Write.
func (m *MockTransport) Write(ctx context.Context, data string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.WriteCalls++

	if m.OnWrite != nil {
		if err := m.OnWrite(data); err != nil {
			return err
		}
	}

	if m.WriteError != nil {
		return m.WriteError
	}

	m.WrittenData = append(m.WrittenData, data)
	return nil
}

// ReadMessages implements Transport.ReadMessages.
func (m *MockTransport) ReadMessages(ctx context.Context) <-chan ReadResult {
	m.mu.Lock()
	m.ReadCalls++

	// Create a new channel for this read session
	m.messageCh = make(chan ReadResult, len(m.Messages)+1)

	// Copy messages to avoid race conditions
	messages := make([]map[string]any, len(m.Messages))
	copy(messages, m.Messages)
	m.mu.Unlock()

	// Send all messages then close
	go func() {
		for _, msg := range messages {
			select {
			case <-ctx.Done():
				m.messageCh <- ReadResult{Error: ctx.Err()}
				close(m.messageCh)
				return
			case m.messageCh <- ReadResult{Data: msg}:
			}
		}
		close(m.messageCh)
	}()

	return m.messageCh
}

// Close implements Transport.Close.
func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++

	if m.OnClose != nil {
		if err := m.OnClose(); err != nil {
			return err
		}
	}

	if m.CloseError != nil {
		return m.CloseError
	}

	m.connected = false
	m.ready = false
	m.closed = true

	return nil
}

// IsReady implements Transport.IsReady.
func (m *MockTransport) IsReady() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.IsReadyCalls++
	return m.ready
}

// EndInput implements Transport.EndInput.
func (m *MockTransport) EndInput() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.EndInputCalls++

	if m.EndInputError != nil {
		return m.EndInputError
	}

	return nil
}

// SetReady sets the ready state for testing.
func (m *MockTransport) SetReady(ready bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ready = ready
}

// IsConnected returns whether Connect was called successfully.
func (m *MockTransport) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

// IsClosed returns whether Close was called.
func (m *MockTransport) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

// GetWrittenData returns a copy of all data written to the transport.
func (m *MockTransport) GetWrittenData() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.WrittenData))
	copy(result, m.WrittenData)
	return result
}

// ClearWrittenData clears the captured written data.
func (m *MockTransport) ClearWrittenData() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.WrittenData = make([]string, 0)
}

// InjectMessage sends a message to the current read channel.
// This is useful for simulating messages during an active read.
func (m *MockTransport) InjectMessage(msg map[string]any) {
	m.mu.Lock()
	ch := m.messageCh
	m.mu.Unlock()

	if ch != nil {
		ch <- ReadResult{Data: msg}
	}
}

// InjectError sends an error to the current read channel.
func (m *MockTransport) InjectError(err error) {
	m.mu.Lock()
	ch := m.messageCh
	m.mu.Unlock()

	if ch != nil {
		ch <- ReadResult{Error: err}
	}
}

// Reset resets the mock to its initial state.
func (m *MockTransport) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	m.ready = false
	m.closed = false
	m.ConnectError = nil
	m.WriteError = nil
	m.CloseError = nil
	m.EndInputError = nil
	m.Messages = make([]map[string]any, 0)
	m.WrittenData = make([]string, 0)
	m.ConnectCalls = 0
	m.WriteCalls = 0
	m.ReadCalls = 0
	m.CloseCalls = 0
	m.EndInputCalls = 0
	m.IsReadyCalls = 0
	m.OnWrite = nil
	m.OnClose = nil
	m.messageCh = nil
}

// Ensure MockTransport implements Transport interface
var _ Transport = (*MockTransport)(nil)
