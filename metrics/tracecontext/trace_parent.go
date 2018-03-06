package tracecontext

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// TraceParentHeaderName is the name used in an HTTP header
const TraceParentHeaderName = "Trace-Parent"

// TraceParent is a trace context header that is used to pass trace context
// information across systems for a HTTP request.
type TraceParent struct {
	// Version of the data
	Version uint8

	// TraceID traces an entire transaction
	// Shouldremain constant through the life of the TraceParent
	TraceID [16]byte

	// SpanID identifies a single node in the trace tree
	// "In a Dapper trace tree, the tree nodes are basic units of
	// work which we refer to as spans".
	SpanID [8]byte

	// Options are setsable bits describing the transaction
	Options uint8
}

// CurrentTraceParentVersion is the version of distributed-tracing that we
// currently support
const CurrentTraceParentVersion = 0
const minTraceParentVersion = 0
const maxTraceParentVersion = 0

// IsTraceable is an option bit determining whether data should be sampled or not
const IsTraceable = 0x01

// GenerateTraceParent creates a new TraceParent for use in headers
func GenerateTraceParent() TraceParent {
	return TraceParent{
		Version: CurrentTraceParentVersion,
		TraceID: uuid.New(),
		SpanID:  generateSpanID(),
		Options: IsTraceable,
	}
}

// ParseTraceParent creates a TraceParent from a string
func ParseTraceParent(s string) (TraceParent, error) {
	var err error
	var tp TraceParent

	parts := strings.Split(s, "-")

	if len(parts) != 4 {
		return TraceParent{}, errors.Errorf(
			"Not enough '-', found %d, expected %d",
			len(parts),
			4,
		)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return TraceParent{}, errors.Wrapf(err, "strconv.Atoi(%s)", parts[0])
	}

	if version < minTraceParentVersion || version > maxTraceParentVersion {
		return TraceParent{}, errors.Errorf(
			"Invalid version: expected %d <= %d <= %d",
			minTraceParentVersion,
			version,
			maxTraceParentVersion,
		)
	}
	tp.Version = uint8(version)

	traceBytes, err := hex.DecodeString(parts[1])
	if err != nil {
		return TraceParent{}, errors.Wrapf(err, "hex.DecodeString(%s)", parts[1])
	}
	copy(tp.TraceID[:], traceBytes)

	spanBytes, err := hex.DecodeString(parts[2])
	if err != nil {
		return TraceParent{}, errors.Wrapf(err, "hex.DecodeString(%s)", parts[2])
	}
	copy(tp.SpanID[:], spanBytes)

	options, err := strconv.Atoi(parts[3])
	if err != nil {
		return TraceParent{}, errors.Wrapf(err, "strconv.Atoi(%s)", parts[3])
	}
	tp.Options = uint8(options)

	return tp, nil
}

// String returns a formatted string suitable for use in an HTTP header
func (tp TraceParent) String() string {
	return fmt.Sprintf(
		"%0d-%s-%s-%02d",
		tp.Version,
		tp.TraceIDAsString(),
		tp.SpanIDAsString(),
		tp.Options,
	)
}

// WithNewSpanID returns a TraceParent with the SpanID set to represent a new Node
func (tp TraceParent) WithNewSpanID() TraceParent {
	return TraceParent{
		Version: tp.Version,
		TraceID: tp.TraceID,
		SpanID:  generateSpanID(),
		Options: tp.Options,
	}
}

// TraceIDAsString returns the string form of the TraceID
func (tp TraceParent) TraceIDAsString() string {
	return hex.EncodeToString(tp.TraceID[:])
}

// SpanIDAsString returns the string form of the SpanID
func (tp TraceParent) SpanIDAsString() string {
	return hex.EncodeToString(tp.SpanID[:])
}

// VersionIsValid returns true if the15 TraceParent has an acceptble version
func (tp TraceParent) VersionIsValid() bool {
	return tp.Version == CurrentTraceParentVersion
}

// IsTraceable returns true id the Traceable option is set
func (tp TraceParent) IsTraceable() bool {
	return tp.Options&IsTraceable > 0
}

func generateSpanID() [8]byte {
	var s [8]byte
	u := uuid.New()

	// just give them half of a uuid, maybe something more specific later
	for i := 0; i < 8; i++ {
		s[i] = u[8+i]
	}

	return s
}
