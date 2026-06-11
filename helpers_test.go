package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"
)

// captureStdout captures everything written to os.Stdout during f().
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	return buf.String()
}

// makeMergedPR returns a PR whose merge duration equals createdAgo − mergedAgo.
func makeMergedPR(createdAgo, mergedAgo time.Duration) PullRequest {
	n := time.Now()
	return PullRequest{
		CreatedAt: n.Add(-createdAgo),
		MergedAt:  n.Add(-mergedAgo),
		UpdatedAt: n.Add(-mergedAgo),
	}
}

// prSet returns a realistic set of merged PRs for use across tests.
func prSet() []PullRequest {
	n := time.Now()
	rev1 := n.Add(-46 * time.Hour)
	return []PullRequest{
		{
			Number:        1,
			Author:        "alice",
			Title:         "Feature A",
			CreatedAt:     n.Add(-48 * time.Hour),
			MergedAt:      n.Add(-24 * time.Hour),
			UpdatedAt:     n.Add(-24 * time.Hour),
			Size:          100,
			FilePaths:     []string{"src/foo.go", "src/bar.go"},
			Reviewers:     []string{"bob"},
			FirstReviewAt: &rev1,
		},
		{
			Number:    2,
			Author:    "bob",
			Title:     "Fix B",
			CreatedAt: n.Add(-5 * 24 * time.Hour),
			MergedAt:  n.Add(-4 * 24 * time.Hour),
			UpdatedAt: n.Add(-4 * 24 * time.Hour),
			Size:      200,
			FilePaths: []string{"lib/main.go"},
			Reviewers: []string{"alice", "carol"},
		},
		{
			Number:    3,
			Author:    "carol",
			Title:     "Refactor C that is a very long title to test string truncation here",
			CreatedAt: n.Add(-30 * 24 * time.Hour),
			MergedAt:  n.Add(-29 * 24 * time.Hour),
			UpdatedAt: n.Add(-29 * 24 * time.Hour),
			Size:      500,
			FilePaths: []string{"src/main.go"},
			Reviewers: []string{"bob", "alice"},
		},
	}
}

// openPRSet returns a realistic set of open PRs.
func openPRSet() []PullRequest {
	n := time.Now()
	return []PullRequest{
		{
			Number:    10,
			Author:    "alice",
			Title:     "WIP feature",
			CreatedAt: n.Add(-10 * 24 * time.Hour),
			UpdatedAt: n.Add(-9 * 24 * time.Hour),
			Requested: []string{"bob", "carol"},
		},
		{
			Number:    11,
			Author:    "bob",
			Title:     "Draft",
			CreatedAt: n.Add(-1 * time.Hour),
			UpdatedAt: n.Add(-1 * time.Hour),
		},
	}
}
