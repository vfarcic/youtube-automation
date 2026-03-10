# YouTube Title Analysis Task (A/B Test Data)

You are analyzing YouTube A/B test data to identify patterns in title effectiveness. YouTube runs A/B tests on title variants and reports watch-time share percentages — higher share means that variant kept viewers watching longer compared to alternatives in the same test. This is the primary quality signal.

## Dataset Overview

- **Total Videos with A/B Data**: {{.VideoCount}}

{{.ABData}}

---

## Your Analysis Task

Analyze the A/B test data above to identify what makes titles successful for this YouTube channel. Focus on the **share percentages** as the primary signal — they directly measure which title variant performs better in head-to-head tests, eliminating confounds like video topic or publish timing.

### Key Analysis Principles

- **Share is the primary signal**: A title with 60%+ share clearly outperformed its variants
- **Cross-reference with first-week metrics**: Use views, CTR, likes, and engagement as secondary signals
- **Look for patterns across winning variants**: What do high-share titles have in common?
- **Identify anti-patterns from losing variants**: What do low-share titles share?

---

## Output Requirements

Return your analysis as a **valid JSON object** with the following structure:

```json
{
  "highPerformingPatterns": [
    {
      "pattern": "Pattern name (e.g., 'Titles with numbers')",
      "description": "Clear description of the pattern",
      "impact": "Quantified impact using A/B share data (e.g., 'Variants with numbers averaged 58% share vs 42% without')",
      "examples": ["Winning title 1 (share: 65%)", "Winning title 2 (share: 70%)"]
    }
  ],
  "lowPerformingPatterns": [
    {
      "pattern": "Anti-pattern name",
      "description": "What correlates with lower A/B share",
      "impact": "Quantified negative impact from share data",
      "examples": ["Losing title 1 (share: 25%)", "Losing title 2 (share: 30%)"]
    }
  ],
  "recommendations": [
    {
      "recommendation": "Clear, actionable guidance for title creation",
      "evidence": "A/B test data supporting this (cite specific share percentages)",
      "example": "Before/after example showing how to apply this"
    }
  ],
  "titlesMdContent": "Complete replacement content for the titles.md prompt file. This should be a markdown document with sections for patterns to use, patterns to avoid, and specific guidelines. Include {{"{{.ManuscriptContent}}"}} as a placeholder where the manuscript content will be inserted. Base all guidance on the A/B test evidence above."
}
```

**Critical Requirements:**
- **Return ONLY valid JSON** - no markdown code blocks, no extra text
- **Be specific**: Use concrete A/B share data from the dataset, not generic advice
- **Quantify with share data**: Always cite share percentages as evidence
- **Be actionable**: Focus on patterns that can be directly implemented in title writing
- **Prioritize by share impact**: Highlight patterns with biggest share differences (5-7 recommendations max)
- **titlesMdContent**: Write a complete, self-contained prompt document that can replace the existing titles.md file. Include the literal text `{{"{{.ManuscriptContent}}"}}` as a placeholder. Structure it with clear sections for high-performing patterns, anti-patterns, and actionable rules derived from the A/B data.

Your JSON response will be parsed programmatically, so ensure it's valid and follows the exact structure above.
