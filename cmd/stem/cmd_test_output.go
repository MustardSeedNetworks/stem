// SPDX-License-Identifier: BUSL-1.1

package main

// Rendering a daemon-owned result (#1166). The result arrives as JSON from
// whichever module ran it, so the CLI reads what is actually there rather
// than type-switching on in-process dataplane structs it no longer builds —
// that switch could never match a decoded result and would have degraded
// every run to a %+v dump.

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// renderResult prints the daemon's result as sorted key/value lines.
func renderResult(data any) {
	for _, field := range flatten("", data) {
		_, _ = fmt.Fprintf(os.Stdout, "  %-28s %s\n", field.key, field.value)
	}
}

// renderCSV prints the result as CSV, taking its columns from the result
// itself. The columns a run produces are the module's to decide, so a fixed
// header here would silently drop whatever did not match it.
func renderCSV(data any) {
	rows := asRows(data)
	if len(rows) == 0 {
		return
	}

	columns := columnsOf(rows)
	_, _ = fmt.Fprintln(os.Stdout, strings.Join(columns, ","))
	for _, row := range rows {
		values := make([]string, 0, len(columns))
		for _, column := range columns {
			values = append(values, csvField(row[column]))
		}
		_, _ = fmt.Fprintln(os.Stdout, strings.Join(values, ","))
	}
}

// field is one leaf of a decoded result, named by its full path.
type field struct {
	key   string
	value string
}

// flatten walks a decoded result depth-first, naming nested leaves by path
// (latency.avgNs) so nothing is hidden inside an object the operator cannot
// see.
func flatten(prefix string, data any) []field {
	switch value := data.(type) {
	case nil:
		return nil
	case map[string]any:
		var fields []field
		for _, key := range sortedKeys(value) {
			fields = append(fields, flatten(join(prefix, key), value[key])...)
		}
		return fields
	case []any:
		var fields []field
		for i, item := range value {
			fields = append(fields, flatten(fmt.Sprintf("%s[%d]", prefix, i), item)...)
		}
		return fields
	default:
		return []field{{key: prefix, value: scalar(value)}}
	}
}

func join(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// asRows normalises a result into the rows CSV wants: one object, or a list
// of them. Anything else has no tabular reading and yields none.
func asRows(data any) []map[string]any {
	switch value := data.(type) {
	case map[string]any:
		return []map[string]any{value}
	case []any:
		rows := make([]map[string]any, 0, len(value))
		for _, item := range value {
			if row, ok := item.(map[string]any); ok {
				rows = append(rows, row)
			}
		}
		return rows
	default:
		return nil
	}
}

// columnsOf returns every scalar column present in any row, sorted, so rows
// with differing keys still line up under one header.
func columnsOf(rows []map[string]any) []string {
	seen := make(map[string]struct{})
	for _, row := range rows {
		for key, value := range row {
			if isScalar(value) {
				seen[key] = struct{}{}
			}
		}
	}
	columns := make([]string, 0, len(seen))
	for key := range seen {
		columns = append(columns, key)
	}
	sort.Strings(columns)
	return columns
}

func isScalar(value any) bool {
	switch value.(type) {
	case map[string]any, []any, nil:
		return false
	default:
		return true
	}
}

// scalar formats one JSON value. Numbers keep the shortest representation
// that round-trips, so 98.5 does not print as 98.500000.
func scalar(value any) string {
	if number, ok := value.(float64); ok {
		return strconv.FormatFloat(number, 'f', -1, 64)
	}
	return fmt.Sprintf("%v", value)
}

// csvField quotes only what would otherwise break the row.
func csvField(value any) string {
	if value == nil {
		return ""
	}
	text := scalar(value)
	if strings.ContainsAny(text, ",\"\n") {
		return `"` + strings.ReplaceAll(text, `"`, `""`) + `"`
	}
	return text
}
