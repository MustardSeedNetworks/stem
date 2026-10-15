// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/MustardSeedNetworks/stem/internal/help"
)

// helpSetting is one Parameters row of the CLI or web UI help.
type helpSetting struct {
	source   string
	testType string
	setting  string
}

// uiHelpSettings collects the `setting:` of every parameter in the web UI's
// help topics, keyed by the topic that lists it.
func uiHelpSettings(t *testing.T) []helpSetting {
	t.Helper()

	files, err := filepath.Glob("../../ui/src/data/help/tests/*.ts")
	if err != nil {
		t.Fatal(err)
	}
	topicRe := regexp.MustCompile(`(?m)^  ([a-z0-9_]+): \{$`)
	settingRe := regexp.MustCompile(`setting: '([^']*)'`)

	var rows []helpSetting
	for _, file := range files {
		data, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		text := string(data)
		topics := topicRe.FindAllStringSubmatchIndex(text, -1)
		for i, topic := range topics {
			end := len(text)
			if i+1 < len(topics) {
				end = topics[i+1][0]
			}
			testType := text[topic[2]:topic[3]]
			for _, m := range settingRe.FindAllStringSubmatch(text[topic[1]:end], -1) {
				rows = append(rows, helpSetting{source: file + " " + testType, testType: testType, setting: m[1]})
			}
		}
	}
	return rows
}

// jsonField returns the field of struct v whose JSON name is name.
func jsonField(v reflect.Value, name string) (reflect.Value, bool) {
	for i := range v.NumField() {
		if tag, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("json"), ","); tag == name {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

// setNonZero gives v, a config field, a value other than its zero value.
func setNonZero(t *testing.T, v reflect.Value) {
	t.Helper()

	switch p := v.Addr().Interface().(type) {
	case *bool:
		*p = true
	case *int:
		*p = 7
	case *uint8:
		*p = 7
	case *uint32:
		*p = 7
	case *uint64:
		*p = 7
	case *float64:
		*p = 7.5
	case *string:
		*p = "x"
	case *[]uint32:
		*p = []uint32{64}
	default:
		t.Fatalf("a config field of type %s is not covered", v.Type())
	}
}

// stepReads is what a run plan step hands its module: the module config, and
// the frame sizes the run plan measures one by one.
func stepReads(testType string, cfg *TestConfig) any {
	return []any{
		convertToModuleConfig("eth0", testType, cfg),
		stepFrameSizes(RunPlanStep{TestType: testType, Config: cfg}),
	}
}

// checkSettingReaches fails unless row.setting names a field of the config
// block row.testType reads, and a value in it changes what the step hands its
// module.
func checkSettingReaches(t *testing.T, row helpSetting) {
	t.Helper()

	blockName, fieldName, _ := strings.Cut(row.setting, ".")
	cfg := &TestConfig{}
	block, ok := jsonField(reflect.ValueOf(cfg).Elem(), blockName)
	if !ok {
		t.Errorf("%s: %q names no config block", row.source, row.setting)
		return
	}
	block.Set(reflect.New(block.Type().Elem()))
	field, ok := jsonField(block.Elem(), fieldName)
	if !ok {
		t.Errorf("%s: %q names no field of the %s block", row.source, row.setting, blockName)
		return
	}
	unset := stepReads(row.testType, cfg)
	setNonZero(t, field)
	if reflect.DeepEqual(stepReads(row.testType, cfg), unset) {
		t.Errorf("%s: %q never reaches the %s module", row.source, row.setting, row.testType)
	}
}

// TestHelpParameterSettingsReachTheModule holds every Parameters row of the
// CLI and web UI help to a setting the daemon reads: a field of the config
// block the topic's test type reads, which changes what its module is given.
// A row naming a setting nothing reads tells an operator to tune a knob that
// does nothing (#1581).
func TestHelpParameterSettingsReachTheModule(t *testing.T) {
	rows := uiHelpSettings(t)
	if len(rows) == 0 {
		t.Fatal("found no web UI help settings; the extraction is broken, not the docs")
	}
	for id, topic := range help.GetAllTests() {
		for _, p := range topic.Parameters {
			rows = append(rows, helpSetting{
				source:   "CLI help topic " + id + " " + p.Name,
				testType: id,
				setting:  p.Setting,
			})
		}
	}

	for _, row := range rows {
		checkSettingReaches(t, row)
	}
}
