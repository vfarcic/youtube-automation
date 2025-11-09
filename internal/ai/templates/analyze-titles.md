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

### Your Analysis Should Include:

#### 1. High-Performing Title Patterns
Identify what the top-performing videos (by views, engagement, watch time) have in common:
- **Length patterns**: Character count ranges that perform best
- **Structural patterns**: Common formats (e.g., "How to...", "X vs Y", "Top N...", question format)
- **Word choice**: Keywords, power words, or phrases that appear in top performers
- **Punctuation**: Use of colons, dashes, questions, exclamation points
- **Specificity vs. generality**: Are top titles more specific or more broad?

For each pattern, provide:
- Clear description of the pattern
- Specific examples from the data
- Quantified impact (e.g., "Titles with numbers average 45% more views")

#### 2. Low-Performing Title Patterns
Identify what correlates with lower performance:
- What do underperforming videos have in common?
- What anti-patterns should be avoided?
- What's missing compared to top performers?

Provide specific examples and quantified differences.

#### 3. Title Length Analysis
- Calculate optimal character count range
- Analyze if very short or very long titles underperform
- Consider YouTube's title truncation (typically ~70 characters)

#### 4. Content Type & Topic Analysis
If you can infer content types from titles (tutorials, comparisons, news, opinions, etc.):
- Which types perform best?
- Are certain title styles better for specific content types?
- Do certain topics or technologies generate more interest?

#### 5. Engagement Patterns
Look beyond just views:
- Do certain title patterns drive more likes relative to views?
- Do certain patterns generate more comments?
- Do certain patterns have better watch time (avg duration)?

#### 6. Actionable Recommendations
Provide **5-7 specific, actionable recommendations** for improving future title generation:
- Each must be specific (not generic advice)
- Each must be data-backed with examples
- Each must be directly actionable

**Format each as:**
- **Recommendation**: [Clear guidance]
- **Evidence**: [Data supporting this recommendation]
- **Example**: [Show how to apply this]

#### 7. Prompt Engineering Suggestions
Based on your findings, suggest specific modifications to the title generation prompt. For example:
- "Include numbers in 30-40% of titles (e.g., 'Top 5...', '3 Ways to...')"
- "Keep titles between 50-65 characters for optimal performance"
- "Use comparison format ('X vs Y') for technical tool reviews"
- "Avoid generic words like 'guide', 'tutorial' in favor of specific outcomes"

---

## Output Requirements

- **Be specific**: Use concrete examples from the data, not generic YouTube advice
- **Quantify everything**: Provide percentages, averages, comparisons
- **Be actionable**: Focus on patterns that can be directly implemented in title generation
- **Prioritize impact**: Highlight the patterns with the biggest performance differences
- **Consider channel context**: Tailor recommendations to what works for THIS channel's content and audience

Your analysis will be used to improve the title generation prompt, so make your recommendations concrete and implementable.
