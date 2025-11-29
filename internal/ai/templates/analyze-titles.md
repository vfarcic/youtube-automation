# YouTube Title Analysis Task

You are analyzing YouTube video performance data to identify patterns in title effectiveness. Your goal is to provide **specific, actionable, data-backed recommendations** that can be used to improve future title generation.

## Dataset Overview

- **Total Videos**: {{len .Videos}}
- **Date Range**: {{.StartDate}} to {{.EndDate}}

## Video Performance Data

{{range .Videos}}
**"{{.Title}}"**
- Views: {{.Views}}
- Avg View Duration: {{printf "%.0f" .AverageViewDuration}} seconds
- Likes: {{.Likes}}
- Comments: {{.Comments}}
- Published: {{.PublishedAt.Format "2006-01-02"}}

{{end}}

---

## Your Analysis Task

Analyze the video performance data above to identify what makes titles successful for this YouTube channel.

### IMPORTANT: Account for Video Age

When analyzing performance, **always consider the publish date**. Older videos naturally accumulate more views over time, so a video from 2 years ago with 50K views may actually be underperforming compared to a 3-month-old video with 30K views.

**Strategies to account for age:**
- Compare videos from similar time periods
- Look at views-per-day or views-per-month rates when possible
- Focus more on engagement metrics (likes/views ratio, comments/views ratio, avg view duration) which are less affected by age
- When identifying patterns, check if they hold true across both old and recent videos

---

## Output Requirements

Return your analysis as a **valid JSON object** with the following structure:

```json
{
  "highPerformingPatterns": [
    {
      "pattern": "Pattern name (e.g., 'Titles with numbers')",
      "description": "Clear description of the pattern",
      "impact": "Quantified impact (e.g., '45% more views on average')",
      "examples": ["Example title 1", "Example title 2"]
    }
  ],
  "lowPerformingPatterns": [
    {
      "pattern": "Anti-pattern name",
      "description": "What correlates with lower performance",
      "impact": "Quantified negative impact",
      "examples": ["Example title 1", "Example title 2"]
    }
  ],
  "titleLengthAnalysis": {
    "optimalRange": "Character count range (e.g., '50-65 characters')",
    "finding": "Description of length impact on performance",
    "data": "Supporting statistics"
  },
  "contentTypeAnalysis": {
    "finding": "Which content types/topics perform best",
    "topPerformers": ["Content type 1", "Content type 2"],
    "data": "Supporting statistics"
  },
  "engagementPatterns": {
    "finding": "Title patterns that drive engagement beyond just views",
    "likesPattern": "Patterns that drive likes",
    "commentsPattern": "Patterns that drive comments",
    "watchTimePattern": "Patterns that drive longer watch time"
  },
  "recommendations": [
    {
      "recommendation": "Clear, actionable guidance",
      "evidence": "Data supporting this recommendation with specific metrics",
      "example": "How to apply this (before/after example or specific approach)"
    }
  ],
  "promptSuggestions": [
    "Specific modification to title generation prompt (e.g., 'Include numbers in 30-40% of titles')",
    "Another specific suggestion (e.g., 'Keep titles between 50-65 characters')",
    "Use comparison format ('X vs Y') for technical tool reviews"
  ]
}
```

**Critical Requirements:**
- **Return ONLY valid JSON** - no markdown code blocks, no extra text
- **Be specific**: Use concrete examples from the data, not generic advice
- **Quantify everything**: Provide percentages, averages, comparisons in impact fields
- **Be actionable**: Focus on patterns that can be directly implemented
- **Prioritize impact**: Highlight patterns with biggest performance differences (5-7 recommendations max)
- **Consider channel context**: Tailor to THIS channel's content and audience

Your JSON response will be parsed programmatically, so ensure it's valid and follows the exact structure above.
