// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azurelogsingestionexporter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkRecordsSingleChunk(t *testing.T) {
	records := []map[string]any{
		{"TimeGenerated": "2026-05-27T12:00:00Z", "Message": "test1"},
		{"TimeGenerated": "2026-05-27T12:00:01Z", "Message": "test2"},
	}

	chunks, err := chunkRecords(records)
	require.NoError(t, err)
	assert.Len(t, chunks, 1)

	var parsed []map[string]any
	require.NoError(t, json.Unmarshal(chunks[0], &parsed))
	assert.Len(t, parsed, 2)
	assert.Equal(t, "test1", parsed[0]["Message"])
	assert.Equal(t, "test2", parsed[1]["Message"])
}

func TestChunkRecordsMultipleChunks(t *testing.T) {
	// Create records large enough to exceed maxPayloadBytes
	var records []map[string]any
	largeMessage := strings.Repeat("x", 100*1024) // 100KB per record
	for i := 0; i < 15; i++ {
		records = append(records, map[string]any{
			"TimeGenerated": "2026-05-27T12:00:00Z",
			"Message":       largeMessage,
		})
	}

	chunks, err := chunkRecords(records)
	require.NoError(t, err)
	assert.Greater(t, len(chunks), 1)

	totalRecords := 0
	for _, chunk := range chunks {
		assert.LessOrEqual(t, len(chunk), maxPayloadBytes+1024) // allow some overhead for JSON encoding
		var parsed []map[string]any
		require.NoError(t, json.Unmarshal(chunk, &parsed))
		totalRecords += len(parsed)
	}
	assert.Equal(t, 15, totalRecords)
}

func TestChunkRecordsEmpty(t *testing.T) {
	chunks, err := chunkRecords(nil)
	require.NoError(t, err)
	assert.Empty(t, chunks)
}

func TestChunkRecordsSingleLargeRecord(t *testing.T) {
	records := []map[string]any{
		{"TimeGenerated": "2026-05-27T12:00:00Z", "Message": strings.Repeat("x", 500*1024)},
	}

	chunks, err := chunkRecords(records)
	require.NoError(t, err)
	assert.Len(t, chunks, 1)
}
