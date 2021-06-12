// +build release

package main

const (
	MINIMUM_LOG_LEVEL = LOG_LEVEL_WARN
	BUILD_TYPE        = RELEASE
)

func start_pprof() {}

func init_function_time_tracker() {}

func measure_execution_time(name string) func() { return func() {} }

func close_function_time_tracker() {}
