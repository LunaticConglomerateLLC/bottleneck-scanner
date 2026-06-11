# Bottleneck

> A fork of [josephgoksu/bottleneck](https://github.com/josephgoksu/bottleneck) with multi-repo team analysis, open PR visibility, service account handling, and flexible output modes.

Bottleneck is a CLI tool for Engineering Managers and Platform Engineers to analyze GitHub Pull Request velocity. It goes beyond simple averages to identify _where_ your development process is getting stuck, providing actionable insights into your team's workflow and codebase health.

## What this fork adds

| Feature | Original | This fork |
|---|---|---|
| Single-repo analysis | ✅ | ✅ |
| Multi-repo team analysis (`--team`) | — | ✅ |
| Open PR queue (stale PR tracking) | — | ✅ |
| Service account / bot separation (`--service-accounts`) | — | ✅ |
| Output modes: summary, table, JSON (`--report`) | — | ✅ |
| Team config via YAML | — | ✅ |

## Key Features

- **True Velocity Stats:** Detailed breakdown of Time to Merge (PR Creation → Merge), including Median, Average, and Percentiles.
- **Size vs Speed Analysis:** Calculates the correlation between PR size (lines of code changed) and merge time.
- **Directory Hotspots:** Identifies which parts of your codebase are "swamps" associated with the slowest average merge times.
- **Long Tail Contributors:** Highlights authors most frequently involved in the slowest 10% of PRs.
- **Review Efficiency:** Splits merge time into Triage Time (Created → First Review) and Review Time (First Review → Merged).
- **Monthly Trends:** Visual indicators (🚀/🐢) to see if velocity is improving or degrading month-over-month.
- **Forecast:** Moving average prediction for the next 30 days based on recent trends.
- **Merge Distribution:** Histogram visualizing the distribution of merge times.
- **Leaderboard:** Most active and fastest contributors by average merge time.
- **Smart Filtering:** Exclude statistical outliers (top/bottom 5%) and fetch large datasets with automatic pagination.

## Installation

### Prerequisites

- [Go](https://go.dev/) (1.21+)
- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated (`gh auth login`)

### Build

```bash
git clone <this-repo>
cd bottleneck
go build -o bottleneck main.go
go install
```

## Usage

### Single repo

```bash
bottleneck [flags] <owner/repo>
```

### Team (multi-repo)

```bash
bottleneck --team team.yml [flags]
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--limit <n>` | `100` | Max merged PRs to fetch (supports pagination) |
| `--exclude-outliers` | `false` | Exclude fastest/slowest 5% from analysis |
| `--timeout <duration>` | `30s` | Timeout per GitHub API request |
| `--delay <duration>` | `200ms` | Delay between requests (increase to avoid rate limits) |
| `--team <file>` | — | Path to team YAML config (mutually exclusive with positional arg) |
| `--report <mode>` | `full` | Output mode: `full`, `summary`, `table`, `json` |
| `--service-accounts <mode>` | `separate` | Bot handling: `separate`, `exclude`, `label` |

### Team config YAML

```yaml
name: my-team

repos:
  - org/repo-a
  - org/repo-b

members:
  - alice
  - bob

service_accounts:
  - renovate-bot
  - github-actions[bot]
```

### Example commands

```bash
# Single repo, 200 PRs, no outliers
bottleneck --limit 200 --exclude-outliers lancedb/lancedb

# Multi-repo team report as JSON
bottleneck --team team.yml --report json

# Team report, bots excluded, summary mode
bottleneck --team team.yml --service-accounts exclude --report summary
```

## Sample Output

```text
🔍 Fetching merged PRs for lancedb/lancedb (limit 200)...
✂️  Outlier filtering active. Reduced from 200 to 180 PRs.
------------------------------------------------------------
📊 GENERAL STATISTICS
   (Time from PR Creation -> Merge)
   Count:   180
   Average: 1d 20h
   Median:  9h 39m
   Min:     6m 45s
   Max:     17d 4h
------------------------------------------------------------
🚦 REVIEW EFFICIENCY
   Avg Time to First Review:   20h 15m (Triage Speed)
   Avg Review to Merge:        1d 0h (Coding/Fixing Speed)
------------------------------------------------------------
📐 SIZE vs SPEED ANALYSIS
   Correlation Coeff: 0.11  (Range: -1.0 to +1.0)
   ✅ RESULT: Weak/No Correlation (< 0.3)
      Insight: Small PRs are getting stuck just as often as huge ones.
      Action:  Your bottleneck is likely PROCESS (Triage/CI/Availability), not code size.
------------------------------------------------------------
🔥 DIRECTORY HOTSPOTS (Avg Merge Time)
   nodejs              : 3d 21h (avg over 40 PRs)
   java                : 3d 14h (avg over 5 PRs)
   docs                : 2d 19h (avg over 20 PRs)
   python              : 2d 10h (avg over 70 PRs)
   rust                : 2d 4h (avg over 57 PRs)
------------------------------------------------------------
🐌 LONG TAIL CONTRIBUTORS (Handling the Slowest 10%)
   westonpace     : 3 slow PRs
   naaa760        : 3 slow PRs
   wjones127      : 2 slow PRs
   (Note: These authors might be tackling the hardest complexity, not working slowly.)
------------------------------------------------------------
📈 MONTHLY TRENDS
   2025-10: 2d 3h           (29 PRs) 🐢
   2025-11: 1d 4h           (30 PRs) 🚀
   2025-12: 1d 15h          ( 5 PRs) 🐢
------------------------------------------------------------
🔮 FORECAST (Next 30 Days)
   🎯 PREDICTION: ~1d 15h / PR
   🏁 TREND:      📈 Speeding Up
------------------------------------------------------------
📊 MERGE TIME DISTRIBUTION
   < 1h       : ■■                   (15)
   1h - 1d    : ■■■■■■■■■■■■■■■■■■■■ (106)
   1d - 1w    : ■■■■■■■■             (44)
   1w - 1mo   : ■■                   (15)
   > 1mo      :                      (0)
```

## Interpreting the Data

1. **"Why does it feel slow?"** — Check General Statistics. If Average is significantly higher than Median, a few nightmare PRs are dragging the average down. Your typical flow may be fine.

2. **"Are we ignoring PRs, or is review taking too long?"** — Examine Review Efficiency.
   - High **Triage Time** (>24h to first review): PRs are sitting unnoticed. Consider a reviewer rotation.
   - High **Review to Merge**: PRs are complex or CI is slow. Encourage smaller PRs or invest in faster pipelines.

3. **"Is the codebase itself causing bottlenecks?"** — Look at Directory Hotspots. Directories with consistently high merge times may need refactoring or dedicated reviewers.

4. **"Is PR size impacting velocity?"** — Consult Size vs Speed Analysis.
   - Strong positive correlation: break work into smaller PRs.
   - Weak/no correlation: the bottleneck is process (triage, review, CI), not PR size.

## Contributing

Contributions are welcome. Open an issue or submit a pull request.

## License

MIT

---

Original project by [@josephgoksu](https://x.com/josephgoksu) — [josephgoksu/bottleneck](https://github.com/josephgoksu/bottleneck)
