// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package awss3exporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter"

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// attrsMarshaler serializes log record attributes as newline-delimited JSON.
// Unlike bodyMarshaler, it reads directly from lr.Attributes() instead of
// lr.Body(), eliminating the need for a preceding set(body, attributes) OTTL
// transform that deep-copies the entire attribute map.
type attrsMarshaler struct{}

func (*attrsMarshaler) format() string {
	return "json"
}

func (*attrsMarshaler) compressed() bool {
	return false
}

func newAttrsMarshaler() attrsMarshaler {
	return attrsMarshaler{}
}

func (attrsMarshaler) MarshalLogs(ld plog.Logs) ([]byte, error) {
	buf := bytes.Buffer{}
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)
		ills := rl.ScopeLogs()
		for j := 0; j < ills.Len(); j++ {
			ils := ills.At(j)
			logs := ils.LogRecords()
			for k := 0; k < logs.Len(); k++ {
				lr := logs.At(k)
				raw := lr.Attributes().AsRaw()
				jsonBytes, err := json.Marshal(raw)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal log attributes: %w", err)
				}
				buf.Write(jsonBytes)
				buf.WriteString("\n")
			}
		}
	}
	return buf.Bytes(), nil
}

func (s attrsMarshaler) MarshalTraces(_ ptrace.Traces) ([]byte, error) {
	return nil, fmt.Errorf("traces can't be marshaled into %s format", s.format())
}

func (s attrsMarshaler) MarshalMetrics(_ pmetric.Metrics) ([]byte, error) {
	return nil, fmt.Errorf("metrics can't be marshaled into %s format", s.format())
}
