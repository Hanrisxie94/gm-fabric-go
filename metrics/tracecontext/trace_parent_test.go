package tracecontext

import "testing"

func TestTraceParent(t *testing.T) {
	// test that we can parse a TraceParent that we generate
	for i := 0; i < 1000; i++ {
		tp1 := GenerateTraceParent()
		tps := tp1.String()
		tp2, err := ParseTraceParent(tps)
		if err != nil {
			t.Fatalf("unable to parse '%s'", tps)
		}
		if tp2 != tp1 {
			t.Fatalf("parse mismatch: %s != %s", tp2, tp1)
		}

		if !tp1.VersionIsValid() {
			t.Fatal("invalid version")
		}

		if !tp1.IsTraceable() {
			t.Fatal("not traceable")
		}
	}

	tp1 := GenerateTraceParent()
	tp2 := tp1.WithNewSpanID()

	if tp1.Version != tp2.Version {
		t.Fatalf("version mismatch: %d != %d", tp1.Version, tp2.Version)
	}

	if tp1.TraceID != tp2.TraceID {
		t.Fatalf("TraceID mismatch: %s != %s",
			tp1.TraceIDAsString(), tp2.TraceIDAsString())
	}

	if tp1.SpanID == tp2.SpanID {
		t.Fatalf("SpanID match: %s != %s",
			tp1.SpanIDAsString(), tp2.SpanIDAsString())
	}

	if tp1.Options != tp2.Options {
		t.Fatalf("Options mismatch: %d != %d", tp1.Options, tp2.Options)
	}
}
