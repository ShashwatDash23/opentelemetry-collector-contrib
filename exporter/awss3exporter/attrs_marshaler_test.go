// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awss3exporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestAttrsMarshalerBasic(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	sl := rls.ScopeLogs().AppendEmpty()

	ts := pcommon.Timestamp(int64(0) * time.Millisecond.Nanoseconds())
	lr := sl.LogRecords().AppendEmpty()
	lr.SetTimestamp(ts)
	lr.Attributes().PutStr("channel", "STORAGE")
	lr.Attributes().PutStr("message", "compacting L0")
	lr.Attributes().PutInt("node_id", 3)

	marshaler := &attrsMarshaler{}
	require.NotNil(t, marshaler)
	body, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Equal(t, `{"channel":"STORAGE","message":"compacting L0","node_id":3}`+"\n", string(body))
}

func TestAttrsMarshalerNestedMap(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	sl := rls.ScopeLogs().AppendEmpty()

	lr := sl.LogRecords().AppendEmpty()
	lr.Attributes().PutStr("channel", "TELEMETRY")
	event := lr.Attributes().PutEmptyMap("event")
	event.PutStr("EventType", "sampled_query")
	event.PutStr("Statement", "SELECT 1")

	marshaler := &attrsMarshaler{}
	body, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Contains(t, string(body), `"channel":"TELEMETRY"`)
	assert.Contains(t, string(body), `"EventType":"sampled_query"`)
}

// TestAttrsMarshalerMatchesBodyMarshaler verifies that the attrs marshaler
// produces identical output to the body marshaler when body is set from
// attributes (the current pipeline pattern: set(body, attributes)).
func TestAttrsMarshalerMatchesBodyMarshaler(t *testing.T) {
	attrs := map[string]string{
		"channel":    "STORAGE",
		"message":    "sstable created 12345",
		"hostname":   "crl-prod-5wr-us-west-2-node1",
		"tag":        "cockroach.storage",
		"cluster_id": "37f88f09-8c04-4817-ab17-184dc8d50d26",
		"date":       "2026-06-01T20:00:00.000000Z",
	}

	// Build logs for the body marshaler: set(body, attributes) pattern.
	bodyLogs := plog.NewLogs()
	bLR := bodyLogs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	for k, v := range attrs {
		bLR.Attributes().PutStr(k, v)
	}
	// Simulate set(body, attributes): deep-copy attributes into body.
	bLR.Attributes().CopyTo(bLR.Body().SetEmptyMap())

	// Build logs for the attrs marshaler: same attributes, body untouched.
	attrsLogs := plog.NewLogs()
	aLR := attrsLogs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	for k, v := range attrs {
		aLR.Attributes().PutStr(k, v)
	}

	bodyM := &bodyMarshaler{}
	attrsM := &attrsMarshaler{}

	bodyOut, err := bodyM.MarshalLogs(bodyLogs)
	require.NoError(t, err)

	attrsOut, err := attrsM.MarshalLogs(attrsLogs)
	require.NoError(t, err)

	assert.Equal(t, string(bodyOut), string(attrsOut),
		"attrs marshaler output must match body marshaler output when body is set from attributes")
}

func TestAttrsMarshalerMultipleRecords(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	sl := rls.ScopeLogs().AppendEmpty()

	lr1 := sl.LogRecords().AppendEmpty()
	lr1.Attributes().PutStr("msg", "first")

	lr2 := sl.LogRecords().AppendEmpty()
	lr2.Attributes().PutStr("msg", "second")

	marshaler := &attrsMarshaler{}
	body, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Equal(t, "{\"msg\":\"first\"}\n{\"msg\":\"second\"}\n", string(body))
}

func TestAttrsMarshalerEmptyAttributes(t *testing.T) {
	logs := plog.NewLogs()
	rls := logs.ResourceLogs().AppendEmpty()
	sl := rls.ScopeLogs().AppendEmpty()
	sl.LogRecords().AppendEmpty()

	marshaler := &attrsMarshaler{}
	body, err := marshaler.MarshalLogs(logs)
	assert.NoError(t, err)
	assert.Equal(t, "{}\n", string(body))
}
