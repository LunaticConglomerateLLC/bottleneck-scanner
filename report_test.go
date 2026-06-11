package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// ---- printGeneralStats ----

func TestPrintGeneralStats(t *testing.T) {
	out := captureStdout(t, func() { printGeneralStats(prSet()) })
	for _, want := range []string{"GENERAL STATISTICS", "Count:", "Average:", "Median:", "Min:", "Max:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output", want)
		}
	}
}

// ---- printReviewStats ----

func TestPrintReviewStats(t *testing.T) {
	t.Run("with reviews shows timing", func(t *testing.T) {
		out := captureStdout(t, func() { printReviewStats(prSet()) })
		if !strings.Contains(out, "REVIEW EFFICIENCY") {
			t.Error("expected REVIEW EFFICIENCY header")
		}
		if !strings.Contains(out, "First Review") {
			t.Errorf("expected review timing in output: %s", out)
		}
	})
	t.Run("no reviews shows message", func(t *testing.T) {
		prs := []PullRequest{{CreatedAt: time.Now().Add(-1 * time.Hour), MergedAt: time.Now()}}
		out := captureStdout(t, func() { printReviewStats(prs) })
		if !strings.Contains(out, "No reviews detected") {
			t.Errorf("expected no-reviews message: %s", out)
		}
	})
	t.Run("review timestamp before creation clamped to zero", func(t *testing.T) {
		n := time.Now()
		rev := n.Add(-20 * time.Hour) // before CreatedAt
		pr := PullRequest{CreatedAt: n.Add(-10 * time.Hour), MergedAt: n, FirstReviewAt: &rev}
		out := captureStdout(t, func() { printReviewStats([]PullRequest{pr}) })
		if !strings.Contains(out, "REVIEW EFFICIENCY") {
			t.Error("expected REVIEW EFFICIENCY header")
		}
	})
	t.Run("review timestamp after merge clamped to zero", func(t *testing.T) {
		n := time.Now()
		rev := n.Add(2 * time.Hour) // after MergedAt
		pr := PullRequest{CreatedAt: n.Add(-10 * time.Hour), MergedAt: n, FirstReviewAt: &rev}
		out := captureStdout(t, func() { printReviewStats([]PullRequest{pr}) })
		if !strings.Contains(out, "REVIEW EFFICIENCY") {
			t.Error("expected REVIEW EFFICIENCY header")
		}
	})
}

// ---- printSizeAnalysis ----

func TestPrintSizeAnalysis(t *testing.T) {
	t.Run("basic output", func(t *testing.T) {
		out := captureStdout(t, func() { printSizeAnalysis(prSet()) })
		if !strings.Contains(out, "SIZE vs SPEED") {
			t.Error("expected SIZE vs SPEED header")
		}
		if !strings.Contains(out, "Correlation Coeff") {
			t.Error("expected correlation line")
		}
	})
	t.Run("strong positive correlation", func(t *testing.T) {
		n := time.Now()
		// size ∝ duration: 100→1h, 200→2h, 300→3h
		prs := []PullRequest{
			{Size: 100, CreatedAt: n.Add(-1 * time.Hour), MergedAt: n},
			{Size: 200, CreatedAt: n.Add(-2 * time.Hour), MergedAt: n},
			{Size: 300, CreatedAt: n.Add(-3 * time.Hour), MergedAt: n},
		}
		out := captureStdout(t, func() { printSizeAnalysis(prs) })
		if !strings.Contains(out, "Strong Positive") {
			t.Errorf("expected strong positive correlation: %s", out)
		}
	})
	t.Run("zero denominator still outputs", func(t *testing.T) {
		// identical PRs → denominator = 0
		n := time.Now()
		prs := []PullRequest{
			{Size: 100, CreatedAt: n.Add(-1 * time.Hour), MergedAt: n},
			{Size: 100, CreatedAt: n.Add(-1 * time.Hour), MergedAt: n},
		}
		out := captureStdout(t, func() { printSizeAnalysis(prs) })
		if !strings.Contains(out, "Correlation Coeff") {
			t.Error("expected output even for zero denominator")
		}
	})
	t.Run("moderate correlation range outputs correlation", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{Size: 50, CreatedAt: n.Add(-1 * time.Hour), MergedAt: n},
			{Size: 100, CreatedAt: n.Add(-4 * time.Hour), MergedAt: n},
			{Size: 150, CreatedAt: n.Add(-2 * time.Hour), MergedAt: n},
			{Size: 200, CreatedAt: n.Add(-6 * time.Hour), MergedAt: n},
			{Size: 250, CreatedAt: n.Add(-3 * time.Hour), MergedAt: n},
			{Size: 300, CreatedAt: n.Add(-8 * time.Hour), MergedAt: n},
		}
		out := captureStdout(t, func() { printSizeAnalysis(prs) })
		if !strings.Contains(out, "Correlation Coeff") {
			t.Error("expected correlation output")
		}
	})
	t.Run("moderate text branch outputs RESULT", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{Size: 10, CreatedAt: n.Add(-3 * time.Hour), MergedAt: n},
			{Size: 20, CreatedAt: n.Add(-1 * time.Hour), MergedAt: n},
			{Size: 30, CreatedAt: n.Add(-4 * time.Hour), MergedAt: n},
			{Size: 40, CreatedAt: n.Add(-2 * time.Hour), MergedAt: n},
			{Size: 50, CreatedAt: n.Add(-5 * time.Hour), MergedAt: n},
		}
		out := captureStdout(t, func() { printSizeAnalysis(prs) })
		if !strings.Contains(out, "RESULT") {
			t.Errorf("expected RESULT in output: %s", out)
		}
	})
}

// ---- printHotspots ----

func TestPrintHotspots(t *testing.T) {
	t.Run("shows top directories", func(t *testing.T) {
		out := captureStdout(t, func() { printHotspots(prSet()) })
		if !strings.Contains(out, "DIRECTORY HOTSPOTS") {
			t.Error("expected DIRECTORY HOTSPOTS header")
		}
		if !strings.Contains(out, "src") {
			t.Errorf("expected 'src' dir in output: %s", out)
		}
	})
	t.Run("root-level files labeled as root", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{FilePaths: []string{"README.md"}, CreatedAt: n.Add(-2 * time.Hour), MergedAt: n},
		}
		out := captureStdout(t, func() { printHotspots(prs) })
		if !strings.Contains(out, "(root files)") {
			t.Errorf("expected root files label: %s", out)
		}
	})
	t.Run("no paths still shows header", func(t *testing.T) {
		prs := []PullRequest{{CreatedAt: time.Now().Add(-1 * time.Hour), MergedAt: time.Now()}}
		out := captureStdout(t, func() { printHotspots(prs) })
		if !strings.Contains(out, "DIRECTORY HOTSPOTS") {
			t.Error("expected header even with no paths")
		}
	})
	t.Run("caps output at top 5 directories", func(t *testing.T) {
		n := time.Now()
		dirs := []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg"}
		prs := make([]PullRequest, len(dirs))
		for i, d := range dirs {
			dur := time.Duration(len(dirs)-i) * time.Hour // aaa=7h, bbb=6h, ..., ggg=1h
			prs[i] = PullRequest{
				FilePaths: []string{d + "/file.go"},
				CreatedAt: n.Add(-dur),
				MergedAt:  n,
			}
		}
		out := captureStdout(t, func() { printHotspots(prs) })
		if !strings.Contains(out, "DIRECTORY HOTSPOTS") {
			t.Error("expected DIRECTORY HOTSPOTS header")
		}
		if strings.Contains(out, "ggg") {
			t.Errorf("should not show 7th directory (ggg): %s", out)
		}
	})
}

// ---- printLongTailAuthors ----

func TestPrintLongTailAuthors(t *testing.T) {
	t.Run("shows header", func(t *testing.T) {
		out := captureStdout(t, func() { printLongTailAuthors(prSet()) })
		if !strings.Contains(out, "LONG TAIL") {
			t.Error("expected LONG TAIL header")
		}
	})
	t.Run("single PR", func(t *testing.T) {
		prs := []PullRequest{makeMergedPR(24*time.Hour, 0)}
		out := captureStdout(t, func() { printLongTailAuthors(prs) })
		if !strings.Contains(out, "LONG TAIL") {
			t.Error("expected header")
		}
	})
	t.Run("caps output at top 5 slow authors", func(t *testing.T) {
		n := time.Now()
		// 60 PRs total → lim = 6 (10%). Slowest 6: ggg(1006h)…bbb(1001h).
		// aaa(1000h) is the 7th slowest, outside the slow-6, so is not printed.
		prs := make([]PullRequest, 0, 60)
		authors := []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg"}
		for i, a := range authors {
			prs = append(prs, PullRequest{
				Author:    a,
				CreatedAt: n.Add(-time.Duration(1000+i) * time.Hour),
				MergedAt:  n,
			})
		}
		for i := 0; i < 53; i++ {
			prs = append(prs, PullRequest{
				Author:    "zzzz",
				CreatedAt: n.Add(-time.Duration(i+1) * time.Hour),
				MergedAt:  n,
			})
		}
		out := captureStdout(t, func() { printLongTailAuthors(prs) })
		if !strings.Contains(out, "LONG TAIL") {
			t.Error("expected LONG TAIL header")
		}
		if strings.Contains(out, "aaa") {
			t.Errorf("should not print aaa (7th slowest): %s", out)
		}
	})
}

// ---- printTrends ----

func TestPrintTrends(t *testing.T) {
	t.Run("shows header", func(t *testing.T) {
		out := captureStdout(t, func() { printTrends(prSet()) })
		if !strings.Contains(out, "MONTHLY TRENDS") {
			t.Error("expected MONTHLY TRENDS header")
		}
	})
	t.Run("multiple months shows header", func(t *testing.T) {
		n := time.Now()
		prs := make([]PullRequest, 4)
		for i := range prs {
			prs[i] = PullRequest{
				CreatedAt: n.AddDate(0, -i-1, 0).Add(-10 * time.Hour),
				MergedAt:  n.AddDate(0, -i-1, 0),
			}
		}
		out := captureStdout(t, func() { printTrends(prs) })
		if !strings.Contains(out, "MONTHLY TRENDS") {
			t.Error("expected MONTHLY TRENDS header")
		}
	})
	t.Run("improving trend shows rocket", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{CreatedAt: n.AddDate(0, -2, 0).Add(-10 * time.Hour), MergedAt: n.AddDate(0, -2, 0)},
			{CreatedAt: n.AddDate(0, -1, 0).Add(-3 * time.Hour), MergedAt: n.AddDate(0, -1, 0)},
		}
		out := captureStdout(t, func() { printTrends(prs) })
		if !strings.Contains(out, "🚀") {
			t.Errorf("expected 🚀 trend: %s", out)
		}
	})
	t.Run("slowing trend shows turtle", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{CreatedAt: n.AddDate(0, -2, 0).Add(-3 * time.Hour), MergedAt: n.AddDate(0, -2, 0)},
			{CreatedAt: n.AddDate(0, -1, 0).Add(-10 * time.Hour), MergedAt: n.AddDate(0, -1, 0)},
		}
		out := captureStdout(t, func() { printTrends(prs) })
		if !strings.Contains(out, "🐢") {
			t.Errorf("expected 🐢 trend: %s", out)
		}
	})
	t.Run("stable trend shows dash", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{CreatedAt: n.AddDate(0, -2, 0).Add(-5 * time.Hour), MergedAt: n.AddDate(0, -2, 0)},
			{CreatedAt: n.AddDate(0, -1, 0).Add(-5 * time.Hour), MergedAt: n.AddDate(0, -1, 0)},
		}
		out := captureStdout(t, func() { printTrends(prs) })
		if !strings.Contains(out, "➖") {
			t.Errorf("expected ➖ (stable) trend: %s", out)
		}
	})
}

// ---- printForecast ----

func TestPrintForecast(t *testing.T) {
	t.Run("not enough data shows header only", func(t *testing.T) {
		out := captureStdout(t, func() { printForecast(prSet()) })
		if !strings.Contains(out, "FORECAST") {
			t.Error("expected FORECAST header")
		}
	})
	t.Run("stable trend", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{CreatedAt: n.AddDate(0, -3, 0).Add(-5 * time.Hour), MergedAt: n.AddDate(0, -3, 0)},
			{CreatedAt: n.AddDate(0, -2, 0).Add(-5 * time.Hour), MergedAt: n.AddDate(0, -2, 0)},
			{CreatedAt: n.AddDate(0, -1, 0).Add(-5 * time.Hour), MergedAt: n.AddDate(0, -1, 0)},
		}
		out := captureStdout(t, func() { printForecast(prs) })
		if !strings.Contains(out, "PREDICTION") {
			t.Errorf("expected PREDICTION: %s", out)
		}
		if !strings.Contains(out, "Stable") {
			t.Errorf("expected Stable trend: %s", out)
		}
	})
	t.Run("speeding up trend", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{CreatedAt: n.AddDate(0, -3, 0).Add(-10 * time.Hour), MergedAt: n.AddDate(0, -3, 0)},
			{CreatedAt: n.AddDate(0, -2, 0).Add(-7 * time.Hour), MergedAt: n.AddDate(0, -2, 0)},
			{CreatedAt: n.AddDate(0, -1, 0).Add(-3 * time.Hour), MergedAt: n.AddDate(0, -1, 0)},
		}
		out := captureStdout(t, func() { printForecast(prs) })
		if !strings.Contains(out, "Speeding Up") {
			t.Errorf("expected Speeding Up: %s", out)
		}
	})
	t.Run("slowing down trend", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{CreatedAt: n.AddDate(0, -3, 0).Add(-3 * time.Hour), MergedAt: n.AddDate(0, -3, 0)},
			{CreatedAt: n.AddDate(0, -2, 0).Add(-7 * time.Hour), MergedAt: n.AddDate(0, -2, 0)},
			{CreatedAt: n.AddDate(0, -1, 0).Add(-10 * time.Hour), MergedAt: n.AddDate(0, -1, 0)},
		}
		out := captureStdout(t, func() { printForecast(prs) })
		if !strings.Contains(out, "Slowing Down") {
			t.Errorf("expected Slowing Down: %s", out)
		}
	})
}

// ---- printHistogram ----

func TestPrintHistogram(t *testing.T) {
	t.Run("all buckets covered", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{CreatedAt: n.Add(-30 * time.Minute), MergedAt: n},
			{CreatedAt: n.Add(-12 * time.Hour), MergedAt: n},
			{CreatedAt: n.Add(-3 * 24 * time.Hour), MergedAt: n},
			{CreatedAt: n.Add(-15 * 24 * time.Hour), MergedAt: n},
			{CreatedAt: n.Add(-40 * 24 * time.Hour), MergedAt: n},
		}
		out := captureStdout(t, func() { printHistogram(prs) })
		for _, label := range []string{"MERGE TIME DISTRIBUTION", "< 1h", "1h - 1d", "1d - 1w", "1w - 1mo", "> 1mo"} {
			if !strings.Contains(out, label) {
				t.Errorf("expected %q in histogram: %s", label, out)
			}
		}
	})
	t.Run("empty still shows header", func(t *testing.T) {
		out := captureStdout(t, func() { printHistogram(nil) })
		if !strings.Contains(out, "MERGE TIME DISTRIBUTION") {
			t.Error("expected header even for empty PRs")
		}
	})
}

// ---- printHeroAnalysis ----

func TestPrintHeroAnalysis(t *testing.T) {
	t.Run("no reviews shows message", func(t *testing.T) {
		out := captureStdout(t, func() { printHeroAnalysis([]PullRequest{{Author: "alice"}}) })
		if !strings.Contains(out, "No reviews found") {
			t.Errorf("expected no reviews message: %s", out)
		}
	})
	t.Run("CRITICAL RISK at 60%", func(t *testing.T) {
		prs := make([]PullRequest, 10)
		for i := 0; i < 6; i++ {
			prs[i].Reviewers = []string{"alice"}
		}
		for i := 6; i < 10; i++ {
			prs[i].Reviewers = []string{"bob"}
		}
		out := captureStdout(t, func() { printHeroAnalysis(prs) })
		if !strings.Contains(out, "CRITICAL RISK") {
			t.Errorf("expected CRITICAL RISK: %s", out)
		}
	})
	t.Run("High Load at 35%", func(t *testing.T) {
		prs := make([]PullRequest, 20)
		for i := 0; i < 7; i++ {
			prs[i].Reviewers = []string{"alice"}
		}
		for i := 7; i < 20; i++ {
			prs[i].Reviewers = []string{"bob"}
		}
		out := captureStdout(t, func() { printHeroAnalysis(prs) })
		if !strings.Contains(out, "High Load") {
			t.Errorf("expected High Load: %s", out)
		}
	})
	t.Run("well distributed shows no risk", func(t *testing.T) {
		prs := make([]PullRequest, 10)
		for i := range prs {
			prs[i].Reviewers = []string{string(rune('a' + i))}
		}
		out := captureStdout(t, func() { printHeroAnalysis(prs) })
		if !strings.Contains(out, "well-distributed") {
			t.Errorf("expected well-distributed: %s", out)
		}
	})
	t.Run("healthy range shows header", func(t *testing.T) {
		// 25% → printed but no risk flag
		prs := make([]PullRequest, 20)
		for i := 0; i < 5; i++ {
			prs[i].Reviewers = []string{"alice"}
		}
		for i := 5; i < 20; i++ {
			prs[i].Reviewers = []string{string(rune('b' + i))}
		}
		out := captureStdout(t, func() { printHeroAnalysis(prs) })
		if !strings.Contains(out, "HERO SYNDROME") {
			t.Error("expected header")
		}
	})
}

// ---- printStaleAnalysis ----

func TestPrintStaleAnalysis(t *testing.T) {
	t.Run("shows stale PR details", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{Number: 1, Title: "old PR", Author: "alice", UpdatedAt: n.Add(-10 * 24 * time.Hour)},
		}
		out := captureStdout(t, func() { printStaleAnalysis(prs) })
		if !strings.Contains(out, "STALE PR DETECTOR") {
			t.Error("expected STALE PR DETECTOR header")
		}
		if !strings.Contains(out, "alice") {
			t.Errorf("expected alice in stale output: %s", out)
		}
		if !strings.Contains(out, "Action") {
			t.Errorf("expected Action prompt: %s", out)
		}
	})
	t.Run("clean board message when no stale", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{Number: 1, Title: "fresh PR", Author: "alice", UpdatedAt: n.Add(-1 * time.Hour)},
		}
		out := captureStdout(t, func() { printStaleAnalysis(prs) })
		if !strings.Contains(out, "Clean board") {
			t.Errorf("expected clean board message: %s", out)
		}
	})
}

// ---- printGhostAnalysis ----

func TestPrintGhostAnalysis(t *testing.T) {
	t.Run("shows ghost reviewer names", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{Number: 1, CreatedAt: n.Add(-72 * time.Hour), Requested: []string{"bob"}},
		}
		out := captureStdout(t, func() { printGhostAnalysis(prs) })
		if !strings.Contains(out, "GHOST REVIEWER") {
			t.Error("expected GHOST REVIEWER header")
		}
		if !strings.Contains(out, "bob") {
			t.Errorf("expected bob in ghost output: %s", out)
		}
	})
	t.Run("no ghosts message", func(t *testing.T) {
		n := time.Now()
		prs := []PullRequest{
			{Number: 1, CreatedAt: n.Add(-1 * time.Hour), Requested: []string{"bob"}},
		}
		out := captureStdout(t, func() { printGhostAnalysis(prs) })
		if !strings.Contains(out, "No ghosts found") {
			t.Errorf("expected no ghosts message: %s", out)
		}
	})
}

// ---- printMemberActivity ----

func TestPrintMemberActivity(t *testing.T) {
	members := map[string]bool{"alice": true, "bob": true}
	out := captureStdout(t, func() { printMemberActivity(prSet(), members) })
	if !strings.Contains(out, "TEAM MEMBER ACTIVITY") {
		t.Error("expected TEAM MEMBER ACTIVITY header")
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected alice in activity output: %s", out)
	}
}

// ---- printServiceAccountActivity ----

func TestPrintServiceAccountActivity(t *testing.T) {
	t.Run("shows bot activity", func(t *testing.T) {
		n := time.Now()
		botMerged := []PullRequest{
			{Author: "bot1", MergedAt: n, CreatedAt: n.Add(-1 * time.Hour), Reviewers: []string{"bot1"}},
		}
		saSet := map[string]bool{"bot1": true}
		out := captureStdout(t, func() { printServiceAccountActivity(botMerged, nil, saSet) })
		if !strings.Contains(out, "SERVICE ACCOUNT ACTIVITY") {
			t.Error("expected SERVICE ACCOUNT ACTIVITY header")
		}
		if !strings.Contains(out, "bot1") {
			t.Errorf("expected bot1: %s", out)
		}
	})
	t.Run("bot reviewer also shown", func(t *testing.T) {
		n := time.Now()
		botMerged := []PullRequest{
			{Author: "bot1", MergedAt: n, CreatedAt: n.Add(-1 * time.Hour), Reviewers: []string{"bot2"}},
		}
		saSet := map[string]bool{"bot1": true, "bot2": true}
		out := captureStdout(t, func() { printServiceAccountActivity(botMerged, nil, saSet) })
		if !strings.Contains(out, "bot2") {
			t.Errorf("expected bot2 reviewer to appear: %s", out)
		}
	})
	t.Run("empty shows no activity message", func(t *testing.T) {
		out := captureStdout(t, func() { printServiceAccountActivity(nil, nil, nil) })
		if !strings.Contains(out, "No service account activity") {
			t.Errorf("expected no activity message: %s", out)
		}
	})
}

// ---- printSingleRepo ----

func TestPrintSingleRepo(t *testing.T) {
	t.Run("with data shows statistics", func(t *testing.T) {
		flags := runFlags{saMode: SAModeSeparate}
		out := captureStdout(t, func() { printSingleRepo(prSet(), openPRSet(), flags) })
		if !strings.Contains(out, "GENERAL STATISTICS") {
			t.Errorf("expected GENERAL STATISTICS: %s", out)
		}
	})
	t.Run("empty does not panic", func(t *testing.T) {
		flags := runFlags{saMode: SAModeSeparate}
		captureStdout(t, func() { printSingleRepo(nil, nil, flags) })
	})
	t.Run("outlier filtering reports reduction", func(t *testing.T) {
		n := time.Now()
		prs := make([]PullRequest, 20)
		for i := range prs {
			prs[i] = PullRequest{
				Author:    "alice",
				CreatedAt: n.Add(-time.Duration(i+1) * 24 * time.Hour),
				MergedAt:  n.Add(-time.Duration(i) * 24 * time.Hour),
				UpdatedAt: n,
			}
		}
		flags := runFlags{saMode: SAModeSeparate, excludeOutliers: true}
		out := captureStdout(t, func() { printSingleRepo(prs, nil, flags) })
		if !strings.Contains(out, "Outlier filtering") {
			t.Errorf("expected outlier filtering message: %s", out)
		}
	})
	t.Run("bot PRs show service account section in separate mode", func(t *testing.T) {
		n := time.Now()
		botPR := PullRequest{
			Author:           "bot1",
			IsServiceAccount: true,
			CreatedAt:        n.Add(-2 * time.Hour),
			MergedAt:         n,
			UpdatedAt:        n,
		}
		flags := runFlags{saMode: SAModeSeparate}
		saSet := map[string]bool{"bot1": true}
		out := captureStdout(t, func() {
			printSingleRepoWithSAMode([]PullRequest{botPR}, nil, flags, saSet)
		})
		if !strings.Contains(out, "SERVICE ACCOUNT") {
			t.Errorf("expected SERVICE ACCOUNT section: %s", out)
		}
	})
}

// ---- printTeamConsolidated ----

func TestPrintTeamConsolidated(t *testing.T) {
	t.Run("with data shows statistics and member activity", func(t *testing.T) {
		cfg := TeamConfig{Name: "test-team"}
		results := []RepoResult{
			{Repo: "org/repo1", MergedPRs: prSet(), OpenPRs: openPRSet()},
		}
		members := map[string]bool{"alice": true}
		flags := runFlags{saMode: SAModeSeparate}
		out := captureStdout(t, func() {
			printTeamConsolidated(cfg, results, members, nil, flags)
		})
		if !strings.Contains(out, "GENERAL STATISTICS") {
			t.Errorf("expected GENERAL STATISTICS: %s", out)
		}
		if !strings.Contains(out, "TEAM MEMBER ACTIVITY") {
			t.Errorf("expected TEAM MEMBER ACTIVITY: %s", out)
		}
	})
	t.Run("with bots shows service account section", func(t *testing.T) {
		n := time.Now()
		botPR := PullRequest{
			Author:           "bot1",
			IsServiceAccount: true,
			CreatedAt:        n.Add(-2 * time.Hour),
			MergedAt:         n,
			UpdatedAt:        n,
		}
		cfg := TeamConfig{Name: "t"}
		results := []RepoResult{{Repo: "r", MergedPRs: []PullRequest{botPR}}}
		saSet := map[string]bool{"bot1": true}
		flags := runFlags{saMode: SAModeSeparate}
		out := captureStdout(t, func() {
			printTeamConsolidated(cfg, results, nil, saSet, flags)
		})
		if !strings.Contains(out, "SERVICE ACCOUNT") {
			t.Errorf("expected SERVICE ACCOUNT in consolidated: %s", out)
		}
	})
	t.Run("exclude outliers still produces statistics", func(t *testing.T) {
		n := time.Now()
		prs := make([]PullRequest, 20)
		for i := range prs {
			prs[i] = PullRequest{
				Author:    "alice",
				CreatedAt: n.Add(-time.Duration(i+1) * 24 * time.Hour),
				MergedAt:  n.Add(-time.Duration(i) * 24 * time.Hour),
				UpdatedAt: n,
			}
		}
		cfg := TeamConfig{Name: "t"}
		results := []RepoResult{{Repo: "r", MergedPRs: prs}}
		flags := runFlags{saMode: SAModeSeparate, excludeOutliers: true}
		out := captureStdout(t, func() {
			printTeamConsolidated(cfg, results, nil, nil, flags)
		})
		if !strings.Contains(out, "GENERAL STATISTICS") {
			t.Errorf("expected stats: %s", out)
		}
	})
	t.Run("empty does not panic", func(t *testing.T) {
		cfg := TeamConfig{Name: "t"}
		results := []RepoResult{{Repo: "r"}}
		flags := runFlags{saMode: SAModeSeparate}
		captureStdout(t, func() {
			printTeamConsolidated(cfg, results, nil, nil, flags)
		})
	})
}

// ---- printTeamTable ----

func TestPrintTeamTable(t *testing.T) {
	t.Run("shows team header and total row", func(t *testing.T) {
		cfg := TeamConfig{Name: "test-team"}
		results := []RepoResult{
			{Repo: "org/repo1", MergedPRs: prSet(), OpenPRs: openPRSet()},
			{Repo: "org/very-long-repository-name-that-exceeds-the-display-limit", MergedPRs: prSet()},
		}
		flags := runFlags{saMode: SAModeSeparate}
		out := captureStdout(t, func() { printTeamTable(cfg, results, nil, nil, flags) })
		if !strings.Contains(out, "TEAM: test-team") {
			t.Errorf("expected TEAM header: %s", out)
		}
		if !strings.Contains(out, "TOTAL") {
			t.Errorf("expected TOTAL row: %s", out)
		}
		if !strings.Contains(out, "...") {
			t.Errorf("expected truncated repo name: %s", out)
		}
	})
	t.Run("exclude outliers still shows total row", func(t *testing.T) {
		n := time.Now()
		prs := make([]PullRequest, 20)
		for i := range prs {
			prs[i] = PullRequest{
				CreatedAt: n.Add(-time.Duration(i+1) * 24 * time.Hour),
				MergedAt:  n.Add(-time.Duration(i) * 24 * time.Hour),
			}
		}
		cfg := TeamConfig{Name: "t"}
		results := []RepoResult{{Repo: "r", MergedPRs: prs}}
		flags := runFlags{saMode: SAModeSeparate, excludeOutliers: true}
		out := captureStdout(t, func() { printTeamTable(cfg, results, nil, nil, flags) })
		if !strings.Contains(out, "TOTAL") {
			t.Errorf("expected TOTAL row: %s", out)
		}
	})
}

// ---- printTeamJSON ----

func TestPrintTeamJSON(t *testing.T) {
	t.Run("valid JSON with consolidated report", func(t *testing.T) {
		cfg := TeamConfig{Name: "test-team"}
		results := []RepoResult{
			{Repo: "org/repo1", MergedPRs: prSet(), OpenPRs: openPRSet()},
		}
		flags := runFlags{saMode: SAModeSeparate}
		out := captureStdout(t, func() { printTeamJSON(cfg, results, nil, nil, flags) })

		var report JSONReport
		if err := json.Unmarshal([]byte(out), &report); err != nil {
			t.Fatalf("invalid JSON output: %v\n%s", err, out)
		}
		if report.TeamName != "test-team" {
			t.Errorf("expected team name 'test-team', got %q", report.TeamName)
		}
		if len(report.Repos) != 1 {
			t.Errorf("expected 1 repo, got %d", len(report.Repos))
		}
		if report.Consolidated == nil {
			t.Error("expected consolidated report")
		}
	})
	t.Run("includes member and SA activity", func(t *testing.T) {
		cfg := TeamConfig{Name: "t"}
		results := []RepoResult{{Repo: "r", MergedPRs: prSet()}}
		members := map[string]bool{"alice": true}
		saSet := map[string]bool{"bot": true}
		flags := runFlags{saMode: SAModeSeparate, excludeOutliers: true}
		out := captureStdout(t, func() { printTeamJSON(cfg, results, members, saSet, flags) })

		var report JSONReport
		if err := json.Unmarshal([]byte(out), &report); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(report.Repos[0].MemberActivity) == 0 {
			t.Error("expected member activity in JSON")
		}
	})
}
