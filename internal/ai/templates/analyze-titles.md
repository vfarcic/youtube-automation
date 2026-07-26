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
  "titlesMdContent": "Complete replacement content for the titles.md prompt file. Must instruct the AI to generate exactly 10 titles as a JSON array of strings. Include {{"{{.ManuscriptContent}}"}} as a placeholder. Structure with high-performing patterns, anti-patterns, and guidelines based on A/B evidence. End with: Response (JSON array only):"
}
```

**Critical Requirements:**
- **Return ONLY valid JSON** - no markdown code blocks, no extra text
- **Be specific**: Use concrete A/B share data from the dataset, not generic advice
- **Quantify with share data**: Always cite share percentages as evidence
- **Be actionable**: Focus on patterns that can be directly implemented in title writing
- **Prioritize by share impact**: Highlight patterns with biggest share differences (5-7 recommendations max)
- **titlesMdContent**: Write a complete, self-contained prompt document that can replace the existing titles.md file. Follow these constraints EXACTLY:
  - Include the literal text `{{"{{.ManuscriptContent}}"}}` as a placeholder where the manuscript will be inserted
  - Structure it with clear sections for high-performing patterns, anti-patterns, and actionable guidelines derived from the A/B data
  - The prompt MUST instruct the AI to generate exactly **10** title suggestions (the user picks from these)
  - The prompt MUST instruct the AI to respond with ONLY a valid JSON array of strings (e.g., `["Title 1", "Title 2", ...]`) — no markdown, no explanations, no annotations
  - Do NOT ask for rule numbers, pattern references, or any metadata alongside titles — just the titles themselves
  - Describe patterns by name (e.g., "Bold Opinionated Claim") not by number (e.g., "Rule 1")
  - Do NOT force specific patterns (e.g., "ensure at least one uses X"). Instead, instruct the AI to choose whichever patterns best fit the manuscript content and aim for diversity across the 10 titles
  - **You MUST include the following two sections VERBATIM in titlesMdContent, in addition to whatever A/B-derived patterns you discover.** These encode search-tail and subscriber-conversion findings from full YouTube Analytics that the A/B watch-time-share data above cannot see, so you cannot derive or omit them — copy them exactly:

    ```
    ## Hard Requirements (Search Tail & Conversion — Non-Negotiable)

    These come from full YouTube Analytics for this channel (search views and subscriber conversion over the days-29-to-120 tail), not from A/B watch-time share. They apply to EVERY one of the 10 titles.

    1. **Every title must contain at least one searchable proper noun or established term** — a tool, standard, or named concept people actually type (`OpenTelemetry`, `Crossplane`, `MCP`, `Dockerfile`, `self-hosting AI models`). Titles built entirely from abstractions ("Let's Fix That", "A Wake-Up Call") earn near-zero search traffic and build no catalog tail — a ~300x gap in measured search views.
    2. **Put the searchable term in the first half of the title.** Front-load the noun people type, not just the most dramatic word.
    3. **Never lead with an unfamiliar product name as the hook.** If the audience does not already run the product, lead with the problem and let the product follow in context. A product the audience doesn't know does not become interesting by being reviewed — only as the vehicle for a problem they already have.

    ## Ranked Format Preferences (Highest Measured Conversion First)

    When multiple patterns fit the manuscript, prefer them in this order (subscribers per 1k views, first 28 days):

    1. **Year-anchored list** — `Top N <domain> Tools You MUST Use in <year>` (14.22 subs/1k, by far the strongest format).
    2. **Contrarian / death-claim on a NAMED subject** (9.91) — e.g. "The End of Infrastructure-as-Code", "Why Self-Hosting AI Models Is a Bad Idea".
    3. **Comparison between tools the audience already uses** (5.7–8.4) — e.g. "Terraform vs. Crossplane vs. Ansible".
    4. **Plain explainer / guide framing** (6.95) — acceptable, but the weakest of the viable patterns.
    ```

  - Also add an anti-pattern to titlesMdContent warning against **evaluation framing of an unfamiliar product** ("Is X Worth It?", "X Review", "X Magic") — 1.25–3.63 subs/1k, worst conversion and retention on the channel — while noting comparisons of tools the audience already uses are fine.

Your JSON response will be parsed programmatically, so ensure it's valid and follows the exact structure above.
