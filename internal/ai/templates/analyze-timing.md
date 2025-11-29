# YouTube Timing Analysis & Recommendation Generation

You are analyzing a YouTube channel's publishing schedule and performance data to generate timing recommendations.

**CRITICAL: All times must be in UTC timezone format (HH:MM, 24-hour).**

## Dataset Overview

- **Total Videos**: {{.TotalVideos}}
- **Analysis Period**: Videos published with various timing patterns

## Current Publishing Pattern

{{range .CurrentPattern}}
- **{{.DayOfWeek}} {{.TimeOfDay}} UTC**: {{.Count}} videos ({{printf "%.1f" .Percentage}}%)
{{end}}

## Performance Data by Time Slot

{{range .PerformanceBySlot}}
**{{.Slot.DayOfWeek}} {{.Slot.TimeOfDay}} UTC** ({{.VideoCount}} videos)
- Avg First-Week Views: {{printf "%.0f" .AvgFirstWeekViews}}
- Avg First-Week CTR: {{printf "%.2f" .AvgFirstWeekCTR}}%
- Avg First-Week Engagement: {{printf "%.2f" .AvgFirstWeekEngagement}}%

{{end}}

---

## Task: Generate 6-8 Timing Recommendations

Your goal is **iterative improvement through data-driven experimentation**.

### Strategy

1. **If time slot has sufficient performance data** (3+ videos):
   - High first-week views/CTR/engagement → **KEEP** in recommendations
   - Low first-week metrics compared to other slots → **REPLACE** with new alternative

2. **If time slot has limited or no data**:
   - Generate new recommendations to test
   - Focus on creating **diverse experimental coverage**

3. **Ensure diversity in recommendations**:
   - Spread across **different days of the week** (Monday through Sunday)
   - Spread across **different UTC hours** (aim for 12+ hour range across recommendations)
   - Avoid clustering all recommendations around similar times

4. **Output 6-8 total recommendations** with mix of:
   - Proven performers (if data exists)
   - New time slots to test (for experimental coverage)

### Reasoning Requirements

Each recommendation must include **substantive reasoning** that explains the strategic value of that time slot. Your reasoning should:

- **Be specific and analytical**: Don't just state what the slot is, explain WHY it's valuable for experimentation or performance
- **Reference data when available**: If performance data exists, cite specific metrics and comparisons
- **Articulate hypotheses for new slots**: When suggesting untested times, explain what pattern or audience behavior you're testing
- **Show diverse strategic thinking**: Each recommendation should test different hypotheses (temporal patterns, day-of-week effects, hour-of-day variations, etc.)
- **Consider experimental design**: Explain how this slot complements other recommendations to create comprehensive coverage
- **Be 2-4 sentences**: Provide depth, not just surface-level descriptions

**What makes good reasoning:**
- ✅ "Testing early morning slot (09:00 UTC) to identify if there's an audience segment that consumes content before mid-day. This complements existing 16:00 data by exploring 7-hour temporal separation. If performance matches or exceeds 16:00 baseline, suggests audience availability isn't time-constrained."
- ❌ "Morning slot to test different times."

### Constraints

- **All times MUST be UTC** (format: "HH:MM" 24-hour, e.g., "09:00", "16:00", "21:00")
- **All days MUST be full English weekday names** (e.g., "Monday", "Tuesday", "Saturday")
- **Maximize experimental diversity**: Recommendations should span multiple days and time ranges
- **Iterative approach**: Keep what works based on data, explore new possibilities where data is missing

### Output Format

Return a **valid JSON array** with 6-8 recommendations. Each recommendation must have:
- `day`: Full weekday name
- `time`: UTC time in HH:MM 24-hour format
- `reasoning`: Substantive strategic explanation (2-4 sentences following requirements above)

**Structure:**
```json
[
  {
    "day": "DayName",
    "time": "HH:MM",
    "reasoning": "Your detailed analytical reasoning here explaining the strategic value of this slot."
  }
]
```

**IMPORTANT**: Return ONLY the JSON array. No markdown code blocks, no extra text, no formatting. Just the raw JSON array starting with `[` and ending with `]`.
