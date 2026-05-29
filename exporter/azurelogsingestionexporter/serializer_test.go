// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azurelogsingestionexporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestSerializeLogRecord(t *testing.T) {
	record := plog.NewLogRecord()
	record.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)))
	record.SetSeverityText("ERROR")
	record.Body().SetStr("something went wrong")
	record.Attributes().PutStr("channel", "SQL_EXEC")
	record.Attributes().PutStr("node_id", "3")

	resource := pcommon.NewResource()

	result := serializeLogRecord(record, resource)

	assert.Equal(t, "ERROR", result["Severity"])
	assert.Equal(t, "something went wrong", result["Message"])
	assert.Equal(t, "SQL_EXEC", result["Channel"])
	assert.Equal(t, "3", result["NodeID"])

	ts, ok := result["TimeGenerated"].(string)
	require.True(t, ok)
	assert.Contains(t, ts, "2026-05-27T12:00:00")
}

func TestSerializeLogRecordMissingAttributes(t *testing.T) {
	record := plog.NewLogRecord()
	record.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)))
	record.SetSeverityText("INFO")
	record.Body().SetStr("test message")

	resource := pcommon.NewResource()

	result := serializeLogRecord(record, resource)

	assert.Equal(t, "INFO", result["Severity"])
	assert.Equal(t, "test message", result["Message"])
	_, hasChannel := result["Channel"]
	assert.False(t, hasChannel)
	_, hasNodeID := result["NodeID"]
	assert.False(t, hasNodeID)
}

func TestSerializeLogRecordZeroTimestamp(t *testing.T) {
	record := plog.NewLogRecord()
	record.Body().SetStr("no timestamp")

	resource := pcommon.NewResource()

	result := serializeLogRecord(record, resource)

	ts, ok := result["TimeGenerated"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, ts)

	parsed, err := time.Parse(time.RFC3339Nano, ts)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().UTC(), parsed, 5*time.Second)
}

func TestSerializeLogRecordEnvVars(t *testing.T) {
	t.Setenv(envClusterID, "cluster-123")
	t.Setenv(envClusterName, "my-cluster")

	record := plog.NewLogRecord()
	record.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	record.Body().SetStr("test")

	resource := pcommon.NewResource()

	result := serializeLogRecord(record, resource)

	assert.Equal(t, "cluster-123", result["CloudClusterID"])
	assert.Equal(t, "my-cluster", result["CloudClusterName"])
}
