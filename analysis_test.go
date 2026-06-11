package main

import (
	"sort"
	"testing"
	"time"
)

// ---- toSet ----

func TestToSet(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		if got := toSet(nil); len(got) != 0 {
			t.Errorf("expected empty set, got %v", got)
		}
	})
	t.Run("basic values", func(t *testing.T) {
		s := toSet([]string{"a", "b", "c"})
		for _, v := range []string{"a", "b", "c"} {
			if !s[v] {
				t.Errorf("missing %q in set", v)
			}
		}
		if s["d"] {
			t.Error("unexpected 'd' in set")
		}
	})
	t.Run("duplicates deduplicated", func(t *testing.T) {
		if got := toSet([]string{"a", "a", "b"}); len(got) != 2 {
			t.Errorf("expected 2 keys, got %d", len(got))
		}
	})
}

// ---- annotateServiceAccounts ----

func TestAnnotateServiceAccounts(t *testing.T) {
	t.Run("marks known bot", func(t *testing.T) {
		prs := []PullRequest{{Author: "bot"}, {Author: "human"}}
		annotateServiceAccounts(prs, map[string]bool{"bot": true})
		if !prs[0].IsServiceAccount {
			t.Error("bot should be marked as service account")
		}
		if prs[1].IsServiceAccount {
			t.Error("human should not be marked as service account")
		}
	})
	t.Run("empty SA set marks nobody", func(t *testing.T) {
		prs := []PullRequest{{Author: "bot"}}
		annotateServiceAccounts(prs, nil)
		if prs[0].IsServiceAccount {
			t.Error("should not mark anyone with empty SA set")
		}
	})
}

// ---- splitBySA ----

func TestSplitBySA(t *testing.T) {
	t.Run("separate keeps two slices", func(t *testing.T) {
		prs := []PullRequest{
			{Author: "human1"},
			{Author: "bot1", IsServiceAccount: true},
			{Author: "human2"},
			{Author: "bot2"},
		}
		human, bots := splitBySA(prs, map[string]bool{"bot2": true}, SAModeSeparate)
		if len(human) != 2 {
			t.Errorf("expected 2 humans, got %d", len(human))
		}
		if len(bots) != 2 {
			t.Errorf("expected 2 bots, got %d", len(bots))
		}
	})
	t.Run("exclude drops bots and returns nil bot slice", func(t *testing.T) {
		prs := []PullRequest{{Author: "human"}, {Author: "bot", IsServiceAccount: true}}
		human, bots := splitBySA(prs, nil, SAModeExclude)
		if len(human) != 1 {
			t.Errorf("expected 1 human, got %d", len(human))
		}
		if bots != nil {
			t.Errorf("expected nil bots, got %v", bots)
		}
	})
	t.Run("label merges all into human slice", func(t *testing.T) {
		prs := []PullRequest{{Author: "human"}, {Author: "bot", IsServiceAccount: true}}
		merged, bots := splitBySA(prs, nil, SAModeLabel)
		if len(merged) != 2 {
			t.Errorf("expected 2 total, got %d", len(merged))
		}
		if bots != nil {
			t.Errorf("expected nil bots slice, got %v", bots)
		}
	})
	t.Run("nil input returns nil slices", func(t *testing.T) {
		human, bots := splitBySA(nil, nil, SAModeSeparate)
		if human != nil || bots != nil {
			t.Error("expected nil slices for empty input")
		}
	})
}

// ---- calcAvgMedian ----

func TestCalcAvgMedian(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		avg, median := calcAvgMedian(nil)
		if avg != 0 || median != 0 {
			t.Errorf("expected (0, 0), got (%v, %v)", avg, median)
		}
	})
	t.Run("single PR", func(t *testing.T) {
		avg, median := calcAvgMedian([]PullRequest{makeMergedPR(10*time.Hour, 0)})
		if avg != 10*time.Hour {
			t.Errorf("avg: want 10h, got %v", avg)
		}
		if median != 10*time.Hour {
			t.Errorf("median: want 10h, got %v", median)
		}
	})
	t.Run("even count uses midpoint average", func(t *testing.T) {
		prs := []PullRequest{makeMergedPR(4*time.Hour, 0), makeMergedPR(8*time.Hour, 0)}
		avg, median := calcAvgMedian(prs)
		if avg != 6*time.Hour {
			t.Errorf("avg: want 6h, got %v", avg)
		}
		if median != 6*time.Hour {
			t.Errorf("median: want 6h, got %v", median)
		}
	})
	t.Run("odd count picks middle value", func(t *testing.T) {
		prs := []PullRequest{
			makeMergedPR(2*time.Hour, 0),
			makeMergedPR(4*time.Hour, 0),
			makeMergedPR(12*time.Hour, 0),
		}
		_, median := calcAvgMedian(prs)
		if median != 4*time.Hour {
			t.Errorf("median: want 4h, got %v", median)
		}
	})
}

// ---- calcAvgFirstReview ----

func TestCalcAvgFirstReview(t *testing.T) {
	n := time.Now()

	t.Run("no reviews returns zero", func(t *testing.T) {
		prs := []PullRequest{{CreatedAt: n.Add(-10 * time.Hour)}}
		if d := calcAvgFirstReview(prs); d != 0 {
			t.Errorf("expected 0 for no reviews, got %v", d)
		}
	})
	t.Run("single review", func(t *testing.T) {
		rev := n.Add(-5 * time.Hour)
		pr := PullRequest{CreatedAt: n.Add(-10 * time.Hour), FirstReviewAt: &rev}
		if d := calcAvgFirstReview([]PullRequest{pr}); d != 5*time.Hour {
			t.Errorf("expected 5h, got %v", d)
		}
	})
	t.Run("negative wait skipped", func(t *testing.T) {
		rev := n.Add(-20 * time.Hour) // review timestamp before CreatedAt
		pr := PullRequest{CreatedAt: n.Add(-10 * time.Hour), FirstReviewAt: &rev}
		if d := calcAvgFirstReview([]PullRequest{pr}); d != 0 {
			t.Errorf("expected 0 for negative wait, got %v", d)
		}
	})
	t.Run("averages multiple waits", func(t *testing.T) {
		rev1 := n.Add(-3 * time.Hour) // wait = 7h (created 10h ago)
		rev2 := n.Add(-8 * time.Hour) // wait = 2h (created 10h ago)
		prs := []PullRequest{
			{CreatedAt: n.Add(-10 * time.Hour), FirstReviewAt: &rev1},
			{CreatedAt: n.Add(-10 * time.Hour), FirstReviewAt: &rev2},
		}
		want := (7*time.Hour + 2*time.Hour) / 2
		if d := calcAvgFirstReview(prs); d != want {
			t.Errorf("expected %v, got %v", want, d)
		}
	})
}

// ---- countStale ----

func TestCountStale(t *testing.T) {
	n := time.Now()
	t.Run("mixed fresh and stale", func(t *testing.T) {
		prs := []PullRequest{
			{UpdatedAt: n.Add(-8 * 24 * time.Hour)},  // stale
			{UpdatedAt: n.Add(-1 * 24 * time.Hour)},  // fresh
			{UpdatedAt: n.Add(-10 * 24 * time.Hour)}, // stale
		}
		if c := countStale(prs); c != 2 {
			t.Errorf("expected 2 stale, got %d", c)
		}
	})
	t.Run("none stale", func(t *testing.T) {
		prs := []PullRequest{{UpdatedAt: n.Add(-1 * time.Hour)}}
		if c := countStale(prs); c != 0 {
			t.Errorf("expected 0 stale, got %d", c)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if c := countStale(nil); c != 0 {
			t.Errorf("expected 0, got %d", c)
		}
	})
}

// ---- countGhosts / ghostNames ----

func TestCountGhosts(t *testing.T) {
	n := time.Now()
	t.Run("old PRs count requested reviewers", func(t *testing.T) {
		prs := []PullRequest{
			{CreatedAt: n.Add(-72 * time.Hour), Requested: []string{"alice", "bob"}},
			{CreatedAt: n.Add(-24 * time.Hour), Requested: []string{"carol"}}, // too new
		}
		if c := countGhosts(prs); c != 2 {
			t.Errorf("expected 2 ghosts, got %d", c)
		}
	})
	t.Run("none", func(t *testing.T) {
		prs := []PullRequest{{CreatedAt: n.Add(-1 * time.Hour), Requested: []string{"alice"}}}
		if c := countGhosts(prs); c != 0 {
			t.Errorf("expected 0, got %d", c)
		}
	})
	t.Run("deduplicates same reviewer across PRs", func(t *testing.T) {
		prs := []PullRequest{
			{CreatedAt: n.Add(-72 * time.Hour), Requested: []string{"alice"}},
			{CreatedAt: n.Add(-72 * time.Hour), Requested: []string{"alice"}},
		}
		if c := countGhosts(prs); c != 1 {
			t.Errorf("expected 1 unique ghost, got %d", c)
		}
	})
}

func TestGhostNames(t *testing.T) {
	n := time.Now()
	t.Run("returns names from old PRs only", func(t *testing.T) {
		prs := []PullRequest{
			{CreatedAt: n.Add(-72 * time.Hour), Requested: []string{"alice", "bob"}},
			{CreatedAt: n.Add(-1 * time.Hour), Requested: []string{"carol"}},
		}
		names := ghostNames(prs)
		sort.Strings(names)
		if len(names) != 2 || names[0] != "alice" || names[1] != "bob" {
			t.Errorf("unexpected ghost names: %v", names)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if names := ghostNames(nil); len(names) != 0 {
			t.Errorf("expected empty, got %v", names)
		}
	})
}

// ---- buildHeroList ----

func TestBuildHeroList(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		if h := buildHeroList(nil); h != nil {
			t.Errorf("expected nil, got %v", h)
		}
	})
	t.Run("no reviews", func(t *testing.T) {
		if h := buildHeroList([]PullRequest{{Author: "alice"}}); h != nil {
			t.Errorf("expected nil for no reviews, got %v", h)
		}
	})
	t.Run("hero detected at 80%", func(t *testing.T) {
		prs := []PullRequest{
			{Reviewers: []string{"alice"}},
			{Reviewers: []string{"alice"}},
			{Reviewers: []string{"alice"}},
			{Reviewers: []string{"alice"}},
			{Reviewers: []string{"bob"}},
		}
		heroes := buildHeroList(prs)
		if len(heroes) == 0 {
			t.Fatal("expected at least one hero")
		}
		if heroes[0].Login != "alice" {
			t.Errorf("expected alice, got %q", heroes[0].Login)
		}
		if heroes[0].Reviews != 4 {
			t.Errorf("expected 4 reviews, got %d", heroes[0].Reviews)
		}
	})
	t.Run("no hero when reviews well distributed", func(t *testing.T) {
		prs := make([]PullRequest, 10)
		for i := range prs {
			prs[i].Reviewers = []string{string(rune('a' + i))}
		}
		if h := buildHeroList(prs); len(h) != 0 {
			t.Errorf("expected no heroes, got %v", h)
		}
	})
}

// ---- buildActivityList ----

func TestBuildActivityList(t *testing.T) {
	t.Run("counts PRs and reviews per member", func(t *testing.T) {
		prs := []PullRequest{
			{Author: "alice", Reviewers: []string{"bob"}},
			{Author: "alice", Reviewers: []string{"carol"}},
			{Author: "bob", Reviewers: []string{"alice"}},
		}
		filter := map[string]bool{"alice": true, "bob": true}
		list := buildActivityList(prs, filter)
		byName := map[string]JSONMemberActivity{}
		for _, e := range list {
			byName[e.Login] = e
		}
		if byName["alice"].PRs != 2 || byName["alice"].Reviews != 1 {
			t.Errorf("alice: want 2PRs/1rev, got %d/%d", byName["alice"].PRs, byName["alice"].Reviews)
		}
		if byName["bob"].PRs != 1 || byName["bob"].Reviews != 1 {
			t.Errorf("bob: want 1PRs/1rev, got %d/%d", byName["bob"].PRs, byName["bob"].Reviews)
		}
	})
	t.Run("empty PRs still lists all members with zero counts", func(t *testing.T) {
		list := buildActivityList(nil, map[string]bool{"alice": true})
		if len(list) != 1 || list[0].PRs != 0 {
			t.Errorf("expected alice with 0 activity, got %v", list)
		}
	})
}

// ---- humanizeDuration ----

func TestHumanizeDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{25 * 24 * time.Hour, "25d 0h"},
		{45 * 24 * time.Hour, "1mo 15d"},
		{400 * 24 * time.Hour, "1y 1mo"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := humanizeDuration(tc.d); got != tc.want {
				t.Errorf("humanizeDuration(%v) = %q, want %q", tc.d, got, tc.want)
			}
		})
	}
}

// ---- humanizeDurationShort ----

func TestHumanizeDurationShort(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "-"},
		{30 * time.Minute, "30m"},
		{2*time.Hour + 30*time.Minute, "2h30m"},
		{3*24*time.Hour + 5*time.Hour, "3d5h"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := humanizeDurationShort(tc.d); got != tc.want {
				t.Errorf("humanizeDurationShort(%v) = %q, want %q", tc.d, got, tc.want)
			}
		})
	}
}

// ---- limitString ----

func TestLimitString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		limit int
		want  string
	}{
		{"under limit passes through", "hi", 10, "hi"},
		{"over limit truncates with ellipsis", "hello world!", 5, "hello..."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := limitString(tc.input, tc.limit); got != tc.want {
				t.Errorf("limitString(%q, %d) = %q, want %q", tc.input, tc.limit, got, tc.want)
			}
		})
	}
}

// ---- filterOutliers ----

func TestFilterOutliers(t *testing.T) {
	t.Run("too few PRs passes through unchanged", func(t *testing.T) {
		prs := []PullRequest{
			makeMergedPR(1*time.Hour, 0),
			makeMergedPR(2*time.Hour, 0),
			makeMergedPR(3*time.Hour, 0),
		}
		if got := filterOutliers(prs); len(got) != 3 {
			t.Errorf("expected 3 (no filtering for <4), got %d", len(got))
		}
	})
	t.Run("5% cut from each end for 20 PRs", func(t *testing.T) {
		prs := make([]PullRequest, 20)
		for i := range prs {
			prs[i] = makeMergedPR(time.Duration(i+1)*time.Hour, 0)
		}
		if got := filterOutliers(prs); len(got) != 18 {
			t.Errorf("expected 18, got %d", len(got))
		}
	})
	t.Run("cut clamps to 1 for very small lists", func(t *testing.T) {
		// 5% of 4 = 0.2 → int truncates to 0 → clamped to 1 → 4-2 = 2 remain
		prs := make([]PullRequest, 4)
		for i := range prs {
			prs[i] = makeMergedPR(time.Duration(i+1)*time.Hour, 0)
		}
		if got := filterOutliers(prs); len(got) != 2 {
			t.Errorf("expected 2 after filtering, got %d", len(got))
		}
	})
}

// ---- buildJSONRepoStats ----

func TestBuildJSONRepoStats(t *testing.T) {
	t.Run("with full data", func(t *testing.T) {
		merged := prSet()
		open := openPRSet()
		members := map[string]bool{"alice": true, "bob": true}
		saSet := map[string]bool{"renovate": true}
		stats := buildJSONRepoStats("org/repo", merged, open, members, saSet)

		if stats.Repo != "org/repo" {
			t.Errorf("expected repo name, got %q", stats.Repo)
		}
		if stats.MergedCount != len(merged) {
			t.Errorf("expected %d merged, got %d", len(merged), stats.MergedCount)
		}
		if stats.OpenCount != len(open) {
			t.Errorf("expected %d open, got %d", len(open), stats.OpenCount)
		}
		if len(stats.MemberActivity) == 0 {
			t.Error("expected member activity")
		}
		if len(stats.ServiceAccountActivity) == 0 {
			t.Error("expected service account activity")
		}
	})
	t.Run("empty data produces non-empty duration strings", func(t *testing.T) {
		stats := buildJSONRepoStats("r", nil, nil, nil, nil)
		if stats.AvgMergeTime == "" {
			t.Error("AvgMergeTime should not be empty string")
		}
	})
}
