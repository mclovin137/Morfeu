package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNewLogger_InfoLevel(t *testing.T) {
	log := NewLogger("info")
	if log == nil {
		t.Fatal("Expected logger, got nil")
	}

	// Should not panic
	log.Info("test message")
}

func TestNewLogger_DebugLevel(t *testing.T) {
	log := NewLogger("debug")
	if log == nil {
		t.Fatal("Expected logger, got nil")
	}

	log.Debug("debug message")
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	log := NewLogger("invalid")
	if log == nil {
		t.Fatal("Expected logger to default to info level")
	}
}

func TestLogger_StringField(t *testing.T) {
	log := NewLogger("info")
	field := log.String("key", "value")

	// Check that field is of correct type
	if field.Type != zapcore.StringType {
		t.Errorf("Expected String field type, got %v", field.Type)
	}
}

func TestLogger_IntField(t *testing.T) {
	log := NewLogger("info")
	field := log.Int("key", 42)

	if field.Type != zapcore.Int64Type {
		t.Errorf("Expected Int field type, got %v", field.Type)
	}
}

func TestLogger_BoolField(t *testing.T) {
	log := NewLogger("info")
	field := log.Bool("key", true)

	if field.Type != zapcore.BoolType {
		t.Errorf("Expected Bool field type, got %v", field.Type)
	}
}
