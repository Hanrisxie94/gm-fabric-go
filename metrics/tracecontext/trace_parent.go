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

type parseFunc func(*TraceParent, string) error

// ParseTraceParent creates a TraceParent from a string
func ParseTraceParent(s string) (TraceParent, error) {
	var err error
	var tp TraceParent
	var parseFuncs = []parseFunc{
		parseVersion,
		parseTraceID,
		parseSpanID,
		parseOptions,
	}

	parts := strings.Split(s, "-")

	if len(parts) != len(parseFuncs) {
		return TraceParent{}, errors.Errorf(
			"Not enough '-', found %d, expected %d",
			len(parts),
			len(parseFuncs),
		)
	}

	for i := 0; i < len(parseFuncs); i++ {
		if err = parseFuncs[i](&tp, parts[i]); err != nil {
			return TraceParent{}, errors.Wrapf(err, "parseFunc[%d] %s",
				i, parts[i],
			)
		}
	}

	return tp, nil
}

func parseVersion(tp *TraceParent, data string) error {
	version, err := strconv.Atoi(data)
	if err != nil {
		return errors.Wrapf(err, "strconv.Atoi(%s)", data)
	}

	if version < minTraceParentVersion || version > maxTraceParentVersion {
		return errors.Errorf(
			"Invalid version: expected %d <= %d <= %d",
			minTraceParentVersion,
			version,
			maxTraceParentVersion,
		)
	}

	tp.Version = uint8(version)
	return nil
}

func parseTraceID(tp *TraceParent, data string) error {
	traceBytes, err := hex.DecodeString(data)
	if err != nil {
		return errors.Wrapf(err, "hex.DecodeString(%s)", data)
	}
	copy(tp.TraceID[:], traceBytes)

	return nil
}

func parseSpanID(tp *TraceParent, data string) error {
	spanBytes, err := hex.DecodeString(data)
	if err != nil {
		return errors.Wrapf(err, "hex.DecodeString(%s)", data)
	}
	copy(tp.SpanID[:], spanBytes)

	return nil
}

func parseOptions(tp *TraceParent, data string) error {
	options, err := strconv.Atoi(data)
	if err != nil {
		return errors.Wrapf(err, "strconv.Atoi(%s)", data)
	}

	tp.Options = uint8(options)

	return nil
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
