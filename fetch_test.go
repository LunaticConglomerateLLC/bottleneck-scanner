package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockGH installs a fake `gh` binary in a temp dir, prepends that dir to PATH,
// and returns a cleanup func that restores PATH. The binary simply prints resp
// to stdout with exit code 0, or exits non-zero when fail is true.
func mockGH(t *testing.T, resp string, fail bool) func() {
	t.Helper()
	dir := t.TempDir()
	jsonFile := filepath.Join(dir, "resp.json")
	if err := os.WriteFile(jsonFile, []byte(resp), 0644); err != nil {
		t.Fatal(err)
	}
	var script string
	if fail {
		script = "#!/bin/sh\nexit 1\n"
	} else {
		script = fmt.Sprintf("#!/bin/sh\ncat '%s'\n", jsonFile)
	}
	ghPath := filepath.Join(dir, "gh")
	if err := os.WriteFile(ghPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	old := os.Getenv("PATH")
	os.Setenv("PATH", dir+":"+old)
	return func() { os.Setenv("PATH", old) }
}

// mockGHSlow installs a fake `gh` that sleeps before responding (to trigger timeout).
func mockGHSlow(t *testing.T) func() {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\nsleep 10\n"
	ghPath := filepath.Join(dir, "gh")
	if err := os.WriteFile(ghPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	old := os.Getenv("PATH")
	os.Setenv("PATH", dir+":"+old)
	return func() { os.Setenv("PATH", old) }
}

// happyResponse is a valid GraphQL JSON response with one PR that exercises
// all node conversion branches: reviews (including author-as-reviewer and
// duplicates), reviewRequests (including empty login), and file paths.
const happyResponse = `{
  "data": {
    "repository": {
      "pullRequests": {
        "nodes": [
          {
            "number": 42,
            "createdAt": "2026-06-01T10:00:00Z",
            "updatedAt": "2026-06-02T09:00:00Z",
            "mergedAt": "2026-06-02T10:00:00Z",
            "title": "Add feature X",
            "additions": 100,
            "deletions": 20,
            "author": {"login": "alice"},
            "reviews": {
              "nodes": [
                {"createdAt": "2026-06-01T14:00:00Z", "author": {"login": "bob"}},
                {"createdAt": "2026-06-01T16:00:00Z", "author": {"login": "alice"}},
                {"createdAt": "2026-06-01T18:00:00Z", "author": {"login": "bob"}}
              ]
            },
            "reviewRequests": {
              "nodes": [
                {"requestedReviewer": {"login": "carol"}},
                {"requestedReviewer": {"login": ""}}
              ]
            },
            "files": {
              "nodes": [
                {"path": "src/main.go"},
                {"path": "README.md"}
              ]
            }
          }
        ],
        "pageInfo": {"hasNextPage": false, "endCursor": ""}
      }
    }
  }
}`

// emptyResponse has no nodes.
const emptyResponse = `{
  "data": {
    "repository": {
      "pullRequests": {
        "nodes": [],
        "pageInfo": {"hasNextPage": false, "endCursor": ""}
      }
    }
  }
}`

func TestFetchPRs_HappyPath_Merged(t *testing.T) {
	cleanup := mockGH(t, happyResponse, false)
	defer cleanup()

	prs, err := fetchPRs("owner", "repo", 100, "MERGED", 10*time.Second, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(prs))
	}
	pr := prs[0]
	if pr.Number != 42 {
		t.Errorf("expected PR #42, got #%d", pr.Number)
	}
	if pr.Author != "alice" {
		t.Errorf("expected author alice, got %q", pr.Author)
	}
	if pr.Size != 120 {
		t.Errorf("expected size 120 (100+20), got %d", pr.Size)
	}
	if pr.FirstReviewAt == nil {
		t.Fatal("expected FirstReviewAt set")
	}
	// Only bob should be a reviewer (alice is the author, second bob deduped)
	if len(pr.Reviewers) != 1 || pr.Reviewers[0] != "bob" {
		t.Errorf("expected [bob] as reviewer, got %v", pr.Reviewers)
	}
	// carol requested, empty login skipped
	if len(pr.Requested) != 1 || pr.Requested[0] != "carol" {
		t.Errorf("expected [carol] requested, got %v", pr.Requested)
	}
	if len(pr.FilePaths) != 2 {
		t.Errorf("expected 2 file paths, got %d", len(pr.FilePaths))
	}
}

func TestFetchPRs_HappyPath_Open(t *testing.T) {
	// State=OPEN uses orderBy UPDATED_AT — exercises that branch
	cleanup := mockGH(t, happyResponse, false)
	defer cleanup()

	prs, err := fetchPRs("owner", "repo", 100, "OPEN", 10*time.Second, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 1 {
		t.Errorf("expected 1 PR, got %d", len(prs))
	}
}

func TestFetchPRs_SmallLimit(t *testing.T) {
	// limit < 100 → toFetch = remaining (exercises the remaining<100 branch)
	cleanup := mockGH(t, happyResponse, false)
	defer cleanup()

	prs, err := fetchPRs("owner", "repo", 5, "MERGED", 10*time.Second, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 1 {
		t.Errorf("expected 1 PR, got %d", len(prs))
	}
}

func TestFetchPRs_EmptyResponse(t *testing.T) {
	// No nodes → break immediately after first page
	cleanup := mockGH(t, emptyResponse, false)
	defer cleanup()

	prs, err := fetchPRs("owner", "repo", 100, "MERGED", 10*time.Second, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 0 {
		t.Errorf("expected 0 PRs, got %d", len(prs))
	}
}

func TestFetchPRs_CommandError(t *testing.T) {
	// gh exits non-zero → fetchPRs returns error
	cleanup := mockGH(t, "", true)
	defer cleanup()

	_, err := fetchPRs("owner", "repo", 100, "MERGED", 10*time.Second, 0)
	if err == nil {
		t.Error("expected error from gh command failure")
	}
}

func TestFetchPRs_InvalidJSON(t *testing.T) {
	cleanup := mockGH(t, "not-valid-json", false)
	defer cleanup()

	_, err := fetchPRs("owner", "repo", 100, "MERGED", 10*time.Second, 0)
	if err == nil {
		t.Error("expected JSON parse error")
	}
}

func TestFetchPRs_Timeout(t *testing.T) {
	cleanup := mockGHSlow(t)
	defer cleanup()

	_, err := fetchPRs("owner", "repo", 100, "MERGED", 50*time.Millisecond, 0)
	if err == nil {
		t.Error("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout message, got: %v", err)
	}
}

// TestFetchPRs_WithDelay tests the second-page delay path.
// We use a paginated response: first call returns hasNextPage=true, second returns empty.
func TestFetchPRs_Pagination(t *testing.T) {
	dir := t.TempDir()

	// Write a counter file so the script returns different responses per call
	counterFile := filepath.Join(dir, "count")
	os.WriteFile(counterFile, []byte("0"), 0644)

	page1File := filepath.Join(dir, "page1.json")
	page2File := filepath.Join(dir, "page2.json")

	page1 := `{
  "data": {
    "repository": {
      "pullRequests": {
        "nodes": [
          {
            "number": 1,
            "createdAt": "2026-05-01T10:00:00Z",
            "updatedAt": "2026-05-02T10:00:00Z",
            "mergedAt": "2026-05-02T10:00:00Z",
            "title": "PR 1",
            "additions": 10,
            "deletions": 5,
            "author": {"login": "alice"},
            "reviews": {"nodes": []},
            "reviewRequests": {"nodes": []},
            "files": {"nodes": []}
          }
        ],
        "pageInfo": {"hasNextPage": true, "endCursor": "cursor1"}
      }
    }
  }
}`
	page2 := emptyResponse

	os.WriteFile(page1File, []byte(page1), 0644)
	os.WriteFile(page2File, []byte(page2), 0644)

	// Script: on first invocation output page1, on second output page2
	script := fmt.Sprintf(`#!/bin/sh
COUNTER=$(cat '%s')
if [ "$COUNTER" = "0" ]; then
  echo 1 > '%s'
  cat '%s'
else
  cat '%s'
fi
`, counterFile, counterFile, page1File, page2File)

	ghPath := filepath.Join(dir, "gh")
	os.WriteFile(ghPath, []byte(script), 0755)

	old := os.Getenv("PATH")
	os.Setenv("PATH", dir+":"+old)
	defer os.Setenv("PATH", old)

	prs, err := fetchPRs("owner", "repo", 100, "MERGED", 10*time.Second, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 PR from page 1, page 2 is empty → loop ends
	if len(prs) != 1 {
		t.Errorf("expected 1 PR from pagination, got %d", len(prs))
	}
}
