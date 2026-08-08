package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOTLPReceiverAndExporter(t *testing.T) {
	ts := NewTraceStore()
	receiver := NewOTLPReceiver("127.0.0.1:43189", ts)

	if err := receiver.Start(); err != nil {
		t.Fatalf("failed to start OTLP receiver: %v", err)
	}
	defer func() {
		_ = receiver.Stop(context.Background())
	}()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	exporter := NewExporter(testServer.URL)
	tr := Trace{
		TraceID: "test-trace-id",
		Spans: []*Span{
			{TraceID: "test-trace-id", SpanID: "span-1", ServiceName: "test-service"},
		},
	}

	err := exporter.ExportTrace(context.Background(), tr)
	if err != nil {
		t.Fatalf("expected export to succeed, got %v", err)
	}
}
