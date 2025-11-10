# Review YouTube Title Analytics

Review AI-generated title performance insights and suggest improvements to the title generation prompt template.

## Usage
```
/analyze-titles
```

## Description
This command helps developers apply insights from YouTube title analytics to improve title generation. It:
- Reads the latest AI analysis from `./tmp/title-analysis-*.md`
- Reviews the current title generation template in `internal/ai/templates/titles.md`
- Suggests specific, actionable improvements based on data-driven insights
- Shows before/after examples of proposed changes
- Guides you through applying improvements

## Workflow

### Step 1: Find Latest Analysis
- Search `./tmp/` directory for `title-analysis-*.md` files
- Identify the most recent analysis (by date in filename: `title-analysis-YYYY-MM-DD.md`)
- If no analysis found, inform user to run: **Analyze → Titles** from the app menu first

### Step 2: Read Analysis Data
- Read the latest `title-analysis-*.md` file
- Extract key AI recommendations about:
  - High-performing title patterns
  - Power words that increase CTR
  - Optimal title length observations
  - Topic/keyword performance insights
  - Successful title structures

### Step 3: Review Current Template
- Read `internal/ai/templates/titles.md`
- Understand current prompt instructions
- Identify areas where AI recommendations could improve the prompt

### Step 4: Suggest Specific Improvements
For each actionable recommendation from the analysis:
- Propose specific text additions/modifications to the template
- Show **before** (current template section) and **after** (proposed change)
- Explain the rationale (link back to data in analysis)
- Prioritize changes by potential impact (based on CTR, views, etc.)

### Step 5: Present Recommendations Summary
Create a structured summary:
```markdown
## Title Template Improvement Recommendations

Based on analysis from: [date]
Videos analyzed: [count]

### High Priority Changes
1. **[Recommendation Title]**
   - **Finding**: [What the data shows]
   - **Current template**: [Relevant section]
   - **Proposed change**: [Specific edit]
   - **Expected impact**: [Why this will help]

### Medium Priority Changes
[Similar structure]

### Low Priority Changes
[Similar structure]
```

### Step 6: Guide Implementation
Ask the user:
```
Would you like me to apply these improvements to the template?
Options:
1. Apply all high-priority changes
2. Apply specific changes (I'll ask which ones)
3. Show me the full updated template first (no changes yet)
4. I'll make the changes manually
```

Based on user choice:
- **Option 1**: Apply all high-priority changes using the Edit tool
- **Option 2**: Present numbered list, let user select which to apply
- **Option 3**: Show complete updated template content for review
- **Option 4**: Provide summary and exit

### Step 7: Validate Changes
If changes were applied:
- Show the updated template content
- Confirm the changes align with AI recommendations
- Suggest testing: Generate titles for a recent video to validate improvements

## Expected Outcomes

✅ **Developer understands** what the data reveals about title performance
✅ **Template improvements are specific** and traceable to data insights
✅ **Changes are easy to apply** with clear before/after examples
✅ **Future title generation** automatically benefits from insights
✅ **Process is iterative** - run analysis periodically to refine further

## Notes

- **Data-Driven**: Every suggestion should reference specific findings from the analysis
- **Actionable**: Focus on concrete changes, not generic advice like "make titles more engaging"
- **Traceable**: Link each recommendation back to the analytics data (CTR, views, etc.)
- **Conservative**: Don't over-optimize; preserve flexibility for different video types
- **Iterative**: This process repeats monthly/quarterly as more data accumulates

## Example Interaction Flow

```
User: /analyze-titles

Claude: Found latest analysis from 2025-11-09 (200 videos analyzed)

Key Findings:
- Videos with numbers in titles averaged 23% higher CTR
- Titles 50-60 characters performed best (avg CTR: 8.2%)
- Question-format titles underperformed (avg CTR: 4.1%)

I've identified 3 high-priority improvements to the template:

1. **Encourage Number-Based Titles**
   Finding: Titles with numbers had 23% higher CTR
   Current: "Generate 5 compelling and SEO-friendly YouTube video titles"
   Proposed: "Generate 5 compelling and SEO-friendly YouTube video titles. Prioritize including specific numbers, metrics, or list formats (e.g., '5 Ways...', '10 Tips...') as they increase CTR by 23% based on historical data."
   Impact: Higher clickthrough rates on future videos

2. **Optimize Title Length**
   Finding: 50-60 character titles had highest CTR (8.2%)
   Current: "Each title must be 70 characters or less"
   Proposed: "Each title should be 50-60 characters (performs best based on channel data), with a maximum of 70 characters."
   Impact: Better optimized title lengths

3. **Reduce Question Formats**
   Finding: Question titles underperformed (4.1% CTR vs 7.8% avg)
   Current: No specific guidance
   Proposed: Add note: "Avoid question-format titles (e.g., 'How to...?') as they underperform on this channel."
   Impact: Fewer low-performing title structures

Would you like me to apply these improvements? [Options 1-4]
```

## Requirements

- Analysis file must exist in `./tmp/title-analysis-*.md`
- Template file exists at `internal/ai/templates/titles.md`
- User has reviewed analysis at least once before running this command
