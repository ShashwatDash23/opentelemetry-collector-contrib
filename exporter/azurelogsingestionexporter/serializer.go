// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package azurelogsingestionexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/azurelogsingestionexporter"

import (
	"os"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

const (
	envClusterID   = "INTRUSION_CRDB_CLUSTER_ID"
	envClusterName = "CLUSTER_NAME"
)

func serializeLogRecord(record plog.LogRecord, resource pcommon.Resource) map[string]any {
	entry := map[string]any{
		"TimeGenerated": formatTimestamp(record.Timestamp()),
		"Severity":      record.SeverityText(),
		"Message":       record.Body().AsString(),
	}

	if v, ok := record.Attributes().Get("channel"); ok {
		entry["Channel"] = v.AsString()
	}

	if v, ok := record.Attributes().Get("node_id"); ok {
		entry["NodeID"] = v.AsString()
	}

	if v := os.Getenv(envClusterID); v != "" {
		entry["CloudClusterID"] = v
	}

	if v := os.Getenv(envClusterName); v != "" {
		entry["CloudClusterName"] = v
	}

	return entry
}

func formatTimestamp(ts pcommon.Timestamp) string {
	if ts == 0 {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}
	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}
