package models

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/selfstat"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeMetricFilterAfterApplyingGlobalTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Filter: Filter{
			TagInclude: []string{"b"},
		},
	}, "123")
	require.NoError(t, ri.Config.Filter.Compile())
	ri.SetDefaultTags(map[string]string{"a": "x", "b": "y"})

	m, err := metric.New("cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42,
		},
		now)
	require.NoError(t, err)

	actual := ri.MakeMetric(m)

	expected, err := metric.New("cpu",
		map[string]string{
			"b": "y",
		},
		map[string]interface{}{
			"value": 42,
		},
		now)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, expected, actual)
}

func TestMakeMetricNoFields(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	}, "123")

	m, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{},
		now,
		telegraf.Untyped)
	m = ri.MakeMetric(m)
	require.NoError(t, err)
	assert.Nil(t, m)
}

// nil fields should get dropped
func TestMakeMetricNilFields(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	}, "123")

	m, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
			"nil":   nil,
		},
		now,
		telegraf.Untyped)
	require.NoError(t, err)
	m = ri.MakeMetric(m)

	expected, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int(101),
		},
		now,
	)
	require.NoError(t, err)

	require.Equal(t, expected, m)
}

func TestMakeMetricWithPluginTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
	}, "123")

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	m = ri.MakeMetric(m)

	expected, err := metric.New("RITest",
		map[string]string{
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricFilteredOut(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
		Tags: map[string]string{
			"foo": "bar",
		},
		Filter: Filter{NamePass: []string{"foobar"}},
	}, "123")

	assert.NoError(t, ri.Config.Filter.Compile())

	m, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	m = ri.MakeMetric(m)
	require.NoError(t, err)
	assert.Nil(t, m)
}

func TestMakeMetricWithDaemonTags(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestRunningInput",
	}, "123")
	ri.SetDefaultTags(map[string]string{
		"foo": "bar",
	})

	m := testutil.MustMetric("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	m = ri.MakeMetric(m)
	expected, err := metric.New("RITest",
		map[string]string{
			"foo": "bar",
		},
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricNameOverride(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:         "TestRunningInput",
		NameOverride: "foobar",
	}, "123")

	m, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.NoError(t, err)
	m = ri.MakeMetric(m)
	expected, err := metric.New("foobar",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricNamePrefix(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:              "TestRunningInput",
		MeasurementPrefix: "foobar_",
	}, "123")

	m, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.NoError(t, err)
	m = ri.MakeMetric(m)
	expected, err := metric.New("foobar_RITest",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMakeMetricNameSuffix(t *testing.T) {
	now := time.Now()
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name:              "TestRunningInput",
		MeasurementSuffix: "_foobar",
	}, "123")

	m, err := metric.New("RITest",
		map[string]string{},
		map[string]interface{}{
			"value": int64(101),
		},
		now,
		telegraf.Untyped)
	require.NoError(t, err)
	m = ri.MakeMetric(m)
	expected, err := metric.New("RITest_foobar",
		nil,
		map[string]interface{}{
			"value": 101,
		},
		now,
	)
	require.NoError(t, err)
	require.Equal(t, expected, m)
}

func TestMetricErrorCounters(t *testing.T) {
	ri := NewRunningInput(&testInput{}, &InputConfig{
		Name: "TestMetricErrorCounters",
	}, "123")

	getGatherErrors := func() int64 {
		for _, r := range selfstat.Metrics() {
			tag, hasTag := r.GetTag("input")
			if r.Name() == "internal_gather" && hasTag && tag == "TestMetricErrorCounters" {
				errCount, ok := r.GetField("errors")
				if !ok {
					t.Fatal("Expected error field")
				}
				return errCount.(int64)
			}
		}
		return 0
	}

	before := getGatherErrors()

	ri.Log().Error("Oh no")

	after := getGatherErrors()

	require.Greater(t, after, before)
	require.GreaterOrEqual(t, int64(1), GlobalGatherErrors.Get())
}

type testInput struct{}

func (t *testInput) Description() string                   { return "" }
func (t *testInput) SampleConfig() string                  { return "" }
func (t *testInput) Gather(acc telegraf.Accumulator) error { return nil }
