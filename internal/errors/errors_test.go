package errors

import (
	"errors"
	"testing"
)

func TestErrorCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrUnknown, "UNKNOWN"},
		{ErrNotFound, "NOT_FOUND"},
		{ErrInvalidArg, "INVALID_ARG"},
		{ErrPermission, "PERMISSION"},
		{ErrTimeout, "TIMEOUT"},
		{ErrConnection, "CONNECTION"},
		{ErrAuth, "AUTH"},
		{ErrRateLimit, "RATE_LIMIT"},
		{ErrInternal, "INTERNAL"},
		{ErrorCode(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.code.String(); got != tt.expected {
			t.Errorf("ErrorCode.String() = %q, want %q", got, tt.expected)
		}
	}
}

func TestHelixError(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		err := New(ErrNotFound, "item not found")
		if err.Code != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err.Code)
		}
		if err.Message != "item not found" {
			t.Errorf("expected 'item not found', got %q", err.Message)
		}
		if err.Stack == "" {
			t.Error("expected non-empty stack")
		}
	})

	t.Run("Error", func(t *testing.T) {
		err := New(ErrNotFound, "item not found")
		expected := "[NOT_FOUND] item not found"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Wrap", func(t *testing.T) {
		original := errors.New("original error")
		err := Wrap(original, ErrInternal, "wrapped error")

		if err.Err != original {
			t.Error("expected wrapped error")
		}
		if err.Code != ErrInternal {
			t.Errorf("expected ErrInternal, got %v", err.Code)
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		original := errors.New("original error")
		err := Wrap(original, ErrInternal, "wrapped error")

		if err.Unwrap() != original {
			t.Error("expected original error")
		}
	})
}

func TestIs(t *testing.T) {
	err := New(ErrNotFound, "not found")

	if !Is(err, ErrNotFound) {
		t.Error("expected Is to return true")
	}

	if Is(err, ErrInternal) {
		t.Error("expected Is to return false")
	}
}

func TestAs(t *testing.T) {
	err := New(ErrNotFound, "not found")

	var target *HelixError
	if !As(err, &target) {
		t.Error("expected As to return true")
	}
	if target.Code != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", target.Code)
	}
}

func TestWrapf(t *testing.T) {
	original := errors.New("original error")
	err := Wrapf(original, ErrInternal, "error in %s", "module")

	if err.Err != original {
		t.Error("expected wrapped error")
	}
	if err.Message != "error in module" {
		t.Errorf("expected 'error in module', got %q", err.Message)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name    string
		err     *HelixError
		code    ErrorCode
		message string
	}{
		{"NotFound", NotFound("not found"), ErrNotFound, "not found"},
		{"NotFoundf", NotFoundf("not found %s", "item"), ErrNotFound, "not found item"},
		{"InvalidArg", InvalidArg("invalid"), ErrInvalidArg, "invalid"},
		{"InvalidArgf", InvalidArgf("invalid %s", "arg"), ErrInvalidArg, "invalid arg"},
		{"Permission", Permission("denied"), ErrPermission, "denied"},
		{"Timeout", Timeout("timed out"), ErrTimeout, "timed out"},
		{"Connection", Connection("connection failed"), ErrConnection, "connection failed"},
		{"Auth", Auth("unauthorized"), ErrAuth, "unauthorized"},
		{"RateLimit", RateLimit("rate limited"), ErrRateLimit, "rate limited"},
		{"Internal", Internal("internal error"), ErrInternal, "internal error"},
		{"Internalf", Internalf("error in %s", "func"), ErrInternal, "error in func"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("expected code %v, got %v", tt.code, tt.err.Code)
			}
			if tt.err.Message != tt.message {
				t.Errorf("expected message %q, got %q", tt.message, tt.err.Message)
			}
		})
	}
}

func TestErrorChain(t *testing.T) {
	err1 := errors.New("level 1")
	err2 := Wrap(err1, ErrInternal, "level 2")
	err3 := Wrap(err2, ErrInternal, "level 3")

	if err3.Error() != "[INTERNAL] level 3: [INTERNAL] level 2: level 1" {
		t.Errorf("unexpected error message: %s", err3.Error())
	}

	if !errors.Is(err3, err1) {
		t.Error("expected errors.Is to find level 1")
	}
}
