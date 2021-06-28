package encrypt

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"
	"io"

	wrapping "github.com/hashicorp/go-kms-wrapping"
	"github.com/hashicorp/go-kms-wrapping/wrappers/aead"
	"github.com/hashicorp/go-kms-wrapping/wrappers/multiwrapper"
	"golang.org/x/crypto/hkdf"
)

// derivedKeyPurpose represents the purpose of the derived key.
type derivedKeyPurpose uint

const (
	// derivedKeyPurposeUnknown is unknown
	derivedKeyPurposeUnknown = iota

	// derivedKeyPurposeEvent is per event operations
	derivedKeyPurposeEvent
)

// String returns a representative string for the key's purpose
func (k derivedKeyPurpose) String() string {
	switch k {
	case derivedKeyPurposeEvent:
		return "event"
	default:
		return "unknown"
	}
}

// RotateWrapper defines an interface for eventlogger payloads which include
// rotated wrapper data.  This interface allows for the rotation of the wrapper,
// salt and info
type RotateWrapper interface {
	// Wrapper to use for event encryption or hmac-sha256 operations
	Wrapper() wrapping.Wrapper

	// HmacSalt to use for event hmac-sha256 operations
	HmacSalt() []byte

	// HmacInfo to use for event hmac-sha256 operations
	HmacInfo() []byte
}

// EventWrapperInfo defines and interface for eventlogger payloads which include
// data used to derive a per event wrapper.
type EventWrapperInfo interface {
	// Event ID to use when deriving keys for crypto operations on the event
	// payload
	EventId() string

	// HmacSalt to use for the event hmac-sha256 operations
	HmacSalt() []byte

	// HmacInfo to use for the event hmac-sha256 operations
	HmacInfo() []byte
}

// NewEventWrapper is used by the Filter to derive a wrapper to use
// for a specific event.  The event must implement the WrapperPayload interface
// for per event wrappers to be derived.
func NewEventWrapper(wrapper wrapping.Wrapper, eventId string) (wrapping.Wrapper, error) {
	const op = "encrypt.NewEventWrapper"
	if wrapper == nil {
		return nil, fmt.Errorf("%s: missing wrapper: %w", op, ErrInvalidParameter)
	}
	if eventId == "" {
		return nil, fmt.Errorf("%s: missing event id: %w", op, ErrInvalidParameter)
	}

	keyId := derivedKeyId(derivedKeyPurposeEvent, wrapper.KeyID(), eventId)

	reader, err := NewDerivedReader(wrapper, 32, []byte(eventId), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	privKey, _, err := ed25519.GenerateKey(reader)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to generate key: %w", op, ErrInvalidParameter)
	}
	derivedWrapper := aead.NewWrapper(nil)
	if _, err := derivedWrapper.SetConfig(map[string]string{
		"key_id": keyId,
	}); err != nil {
		return nil, fmt.Errorf("%s: error setting config on aead wrapper for event id %s: %w", op, eventId, err)
	}
	if err := derivedWrapper.SetAESGCMKeyBytes(privKey); err != nil {
		return nil, fmt.Errorf("%s: error setting key bytes on aead wrapper for event id %s: %w", op, eventId, err)
	}
	return derivedWrapper, nil
}

// derivedKeyId returns a key that represents the derived key
func derivedKeyId(purpose derivedKeyPurpose, wrapperKeyId, eventId string) string {
	return fmt.Sprintf("%s.%s.%s", purpose.String(), wrapperKeyId, eventId)
}

// DerivedReader returns a reader from which keys can be read, using the
// given wrapper, reader length limit, salt and context info. Salt and info can
// be nil.
//
// Example:
//	reader, _ := NewDerivedReader(wrapper, userId, jobId)
// 	key := ed25519.GenerateKey(reader)
func NewDerivedReader(wrapper wrapping.Wrapper, lenLimit int64, salt, info []byte) (*io.LimitedReader, error) {
	const op = "encrypt.NewDerivedReader"
	if wrapper == nil {
		return nil, fmt.Errorf("%s: missing wrapper: %w", op, ErrInvalidParameter)
	}
	if lenLimit < 20 {
		return nil, fmt.Errorf("%s: lenLimit must be >= 20: %w", op, ErrInvalidParameter)
	}
	var aeadWrapper *aead.Wrapper
	switch w := wrapper.(type) {
	case *multiwrapper.MultiWrapper:
		raw := w.WrapperForKeyID("__base__")
		var ok bool
		if aeadWrapper, ok = raw.(*aead.Wrapper); !ok {
			return nil, fmt.Errorf("%s: unexpected wrapper type from multiwrapper base: %w", op, ErrInvalidParameter)
		}
	case *aead.Wrapper:
		if w.GetKeyBytes() == nil {
			return nil, fmt.Errorf("%s: aead wrapper missing bytes: %w", op, ErrInvalidParameter)
		}
		aeadWrapper = w
	default:
		return nil, fmt.Errorf("%s: unknown wrapper type: %w", op, ErrInvalidParameter)
	}
	reader := hkdf.New(sha256.New, aeadWrapper.GetKeyBytes(), salt, info)
	return &io.LimitedReader{
		R: reader,
		N: lenLimit,
	}, nil
}
