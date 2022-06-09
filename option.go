// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package dd_sdk_go_testing

import (
	"runtime"
	"sync"

	testingext "github.com/DataDog/dd-sdk-go-testing/ext"
	"github.com/DataDog/dd-sdk-go-testing/internal/constants"
	"github.com/DataDog/dd-sdk-go-testing/internal/utils"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	// tags contains information detected from CI/CD environment variables.
	tags      map[string]string
	tagsMutex sync.Mutex
)

type config struct {
	skip       int
	spanOpts   []ddtrace.StartSpanOption
	finishOpts []ddtrace.FinishOption
}

// Option represents an option that can be passed to NewServeMux or WrapHandler.
type Option func(*config)

func defaults(cfg *config) {
	// When StartSpanWithFinish is called directly from test function.
	cfg.skip = 1
	cfg.spanOpts = []ddtrace.StartSpanOption{
		tracer.SpanType(testingext.SpanTypeTest),
		tracer.Tag(testingext.SpanKind, spanKind),
		tracer.Tag(ext.ManualKeep, true),
	}

	// Ensure CI tags
	ensureCITags()
	forEachCITags(func(k, v string) {
		cfg.spanOpts = append(cfg.spanOpts, tracer.Tag(k, v))
	})

	cfg.finishOpts = []ddtrace.FinishOption{}
}

func ensureCITags() {
	if tags != nil {
		return
	}

	localTags := utils.GetProviderTags()
	localTags[constants.OSPlatform] = utils.OSName()
	localTags[constants.OSVersion] = utils.OSVersion()
	localTags[constants.OSArchitecture] = runtime.GOARCH
	localTags[constants.RuntimeName] = runtime.Compiler
	localTags[constants.RuntimeVersion] = runtime.Version()

	gitData, _ := utils.LocalGetGitData()

	// Guess Git metadata from a local Git repository otherwise.
	if _, ok := localTags[constants.CIWorkspacePath]; !ok {
		localTags[constants.CIWorkspacePath] = gitData.SourceRoot
	}
	if _, ok := localTags[constants.GitRepositoryURL]; !ok {
		localTags[constants.GitRepositoryURL] = gitData.RepositoryUrl
	}
	if _, ok := localTags[constants.GitCommitSHA]; !ok {
		localTags[constants.GitCommitSHA] = gitData.CommitSha
	}
	if _, ok := localTags[constants.GitBranch]; !ok {
		localTags[constants.GitBranch] = gitData.Branch
	}

	if localTags[constants.GitCommitSHA] == gitData.CommitSha {
		if _, ok := localTags[constants.GitCommitAuthorDate]; !ok {
			localTags[constants.GitCommitAuthorDate] = gitData.AuthorDate.String()
		}
		if _, ok := localTags[constants.GitCommitAuthorName]; !ok {
			localTags[constants.GitCommitAuthorName] = gitData.AuthorName
		}
		if _, ok := localTags[constants.GitCommitAuthorEmail]; !ok {
			localTags[constants.GitCommitAuthorEmail] = gitData.AuthorEmail
		}
		if _, ok := localTags[constants.GitCommitCommitterDate]; !ok {
			localTags[constants.GitCommitCommitterDate] = gitData.CommitterDate.String()
		}
		if _, ok := localTags[constants.GitCommitCommitterName]; !ok {
			localTags[constants.GitCommitCommitterName] = gitData.CommitterName
		}
		if _, ok := localTags[constants.GitCommitCommitterEmail]; !ok {
			localTags[constants.GitCommitCommitterEmail] = gitData.CommitterEmail
		}
		if _, ok := localTags[constants.GitCommitMessage]; !ok {
			localTags[constants.GitCommitMessage] = gitData.CommitMessage
		}
	}

	// Replace global tags with local copy
	tagsMutex.Lock()
	defer tagsMutex.Unlock()

	tags = localTags
}

func getFromCITags(key string) (string, bool) {
	tagsMutex.Lock()
	defer tagsMutex.Unlock()

	if value, ok := tags[key]; ok {
		return value, ok
	}

	return "", false
}

// ForEachCITags will load (if necessary) and iterate through the CI tags that
// should be added to a span for compatibility with DataDog's Continuous
// Integration Visibility.
//
// See https://docs.datadoghq.com/continuous_integration/
func ForEachCITags(itemFunc func(string, string)) {
	ensureCITags()
	forEachCITags(itemFunc)
}

func forEachCITags(itemFunc func(string, string)) {
	tagsMutex.Lock()
	defer tagsMutex.Unlock()

	for k, v := range tags {
		itemFunc(k, v)
	}
}

// WithSpanOptions defines a set of additional ddtrace.StartSpanOption to be added
// to spans started by the integration.
func WithSpanOptions(opts ...ddtrace.StartSpanOption) Option {
	return func(cfg *config) {
		cfg.spanOpts = append(cfg.spanOpts, opts...)
	}
}

// WithSkipFrames defines a how many frames should be skipped for caller autodetection.
// The value should be changed if StartSpanWithFinish is called from a custom wrapper.
func WithSkipFrames(skip int) Option {
	return func(cfg *config) {
		cfg.skip = skip
	}
}

// WithIncrementSkipFrame increments how many frames should be skipped for caller by 1.
func WithIncrementSkipFrame() Option {
	return func(cfg *config) {
		cfg.skip = cfg.skip + 1
	}
}
