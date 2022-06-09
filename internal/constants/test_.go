// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package constants

// Define valid test status types.
const (
	// TestStatusPass marks test execution as passed.
	TestStatusPass = "pass"

	// TestStatusFail marks test execution as failed.
	TestStatusFail = "fail"

	// TestStatusSkip marks test execution as skipped.
	TestStatusSkip = "skip"
)

// Define valid test types.
const (
	// TestTypeTest defines test type as test.
	TestTypeTest = "test"

	// TestTypeBenchmark defines test type as benchmark.
	TestTypeBenchmark = "benchmark"
)
