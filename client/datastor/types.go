package datastor

import (
	"errors"
)

// Errors that get returned in case the server returns
// an unexpected results. The client can return these errors,
// but if any of these errors get returned,
// it means there is a bug in the zstordb code,
// and it should be reported at:
// http://github.com/zero-os/0-stor/issues
var (
	ErrMissingKey     = errors.New("zstor: missing object key (zstordb bug?)")
	ErrMissingData    = errors.New("zstor: missing object data (zstordb bug?)")
	ErrMissingRefList = errors.New("zstor: missing object reference list (zstordb bug?)")
	ErrInvalidStatus  = errors.New("zstor: invalid object status (zstordb bug?)")
	ErrInvalidLabel   = errors.New("zstor: invalid namespace label (zstordb bug?)")
)

// Errors that can be expected to be returned by a zstordb server,
// in "normal" scenarios.
var (
	ErrKeyNotFound            = errors.New("zstordb: key is no found")
	ErrObjectDataCorrupted    = errors.New("zstordb: object data is corrupted")
	ErrObjectCorrupted        = errors.New("zstordb: object (data and/or refList) is corrupted")
	ErrObjectRefListCorrupted = errors.New("zstordb: object reflist is corrupted")
	ErrPermissionDenied       = errors.New("zstordb: JWT token does not permit requested action")
)

type (
	// Namespace contains information about a namespace.
	// None of this information is directly stored somewhere,
	// and instead it is gathered upon request.
	Namespace struct {
		Label               string
		ReadRequestPerHour  int64
		WriteRequestPerHour int64
		NrObjects           int64
	}

	// Object contains the information stored for an object.
	// The Data and ReferenceList are stored separately,
	// but are composed together in this data structure upon request.
	Object struct {
		Key           []byte
		Data          []byte
		ReferenceList []string
	}

	// ObjectKeyResult is the (stream) data type,
	// used as the result data type, when fetching the keys
	// of all objects stored in the current namespace.
	//
	// Only in case of an error, the Error property will be set,
	// in all other cases only the Key property will be set.
	ObjectKeyResult struct {
		Key   []byte
		Error error
	}
)

// ObjectStatus defines the status of an object,
// it can be retrieved using the Check Method of the Client API.
type ObjectStatus uint8

// ObjectStatus enumeration values.
const (
	// The Object is missing.
	ObjectStatusMissing ObjectStatus = iota
	// The Object is OK.
	ObjectStatusOK
	// The Object is corrupted.
	ObjectStatusCorrupted
)

// String implements Stringer.String
func (status ObjectStatus) String() string {
	return _ObjectStatusEnumToStringMapping[status]
}

// private constants for the string

const _ObjectStatusStrings = "missingokcorrupted"

var _ObjectStatusEnumToStringMapping = map[ObjectStatus]string{
	ObjectStatusMissing:   _ObjectStatusStrings[:7],
	ObjectStatusOK:        _ObjectStatusStrings[7:9],
	ObjectStatusCorrupted: _ObjectStatusStrings[9:],
}

// JWTTokenGetter defines the interface of a type which can provide us
// with a valid JWT token at all times.
//
// The implementation can return an error,
// should it not be possible to return a valid JWT Token for whatever reason,
// if no error is returned, it should be assumed that the returned value is valid.
type JWTTokenGetter interface {
	// Get a cached JWT token or create a new JWT token should it be invalid.
	GetJWTToken(namespace string) (string, error)
}