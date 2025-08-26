package tools

import (
	"errors"
	"fmt"
	"io"
	"log"
	"testing"
)

// mockLogger is a simple mock that implements the MyLogger interface
type mockLogger struct {
	infoMessages  []string
	errorMessages []string
}

func (m *mockLogger) Debug(msg string, v ...any) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLogger) Warn(msg string, v ...any) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLogger) Fatal(msg string, v ...any) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLogger) GetDefaultLogger() (*log.Logger, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLogger) Info(msg string, v ...any) {
	m.infoMessages = append(m.infoMessages, fmt.Sprintf(msg, v...))
}

func (m *mockLogger) Error(msg string, v ...any) {
	m.errorMessages = append(m.errorMessages, fmt.Sprintf(msg, v...))
}

func TestPrintWantedReceived(t *testing.T) {
	mockLog := &mockLogger{}
	wantBody := "test string"
	receivedJson := []byte(`{"key": "value"}`)
	//"RECEIVED :%T - %#v\n", receivedJson, string(receivedJson))
	PrintWantedReceived(wantBody, []byte(receivedJson), mockLog)

	expectedInfoMessages := []string{
		"WANTED   :string - \"test string\"\n",
		"RECEIVED :[]uint8 - \"{\\\"key\\\": \\\"value\\\"}\"\n",
	}

	if len(mockLog.infoMessages) != 2 {
		t.Errorf("Expected 2 info messages, got %d", len(mockLog.infoMessages))
	}

	for i, msg := range mockLog.infoMessages {
		if msg != expectedInfoMessages[i] {
			t.Errorf("Expected info message %q, got %q", expectedInfoMessages[i], msg)
		}
	}
}

type mockReadCloser struct {
	closeErr error
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *mockReadCloser) Close() error {
	return m.closeErr
}

func TestCloseBody(t *testing.T) {
	tests := []struct {
		name           string
		closeErr       error
		expectedLogMsg string
	}{
		{
			name:           "No error on close",
			closeErr:       nil,
			expectedLogMsg: "",
		},
		{
			name:           "Error on close",
			closeErr:       errors.New("close error"),
			expectedLogMsg: "Error: close error in test message doing Body.Close().\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := &mockLogger{}
			mockBody := &mockReadCloser{closeErr: tt.closeErr}

			CloseBody(mockBody, "test message", mockLog)

			if tt.expectedLogMsg == "" {
				if len(mockLog.errorMessages) != 0 {
					t.Errorf("Expected no error messages, got %d", len(mockLog.errorMessages))
				}
			} else {
				if len(mockLog.errorMessages) != 1 {
					t.Errorf("Expected 1 error message, got %d", len(mockLog.errorMessages))
				} else if mockLog.errorMessages[0] != tt.expectedLogMsg {
					t.Errorf("Expected error message %q, got %q", tt.expectedLogMsg, mockLog.errorMessages[0])
				}
			}
		})
	}
}
