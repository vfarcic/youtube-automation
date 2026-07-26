Generate 10 compelling and SEO-friendly YouTube video titles based on the provided manuscript. Target 56-65 characters for optimal performance (acceptable range: 55-75 characters maximum).

HARD REQUIREMENTS (from full channel analytics — apply to EVERY title):
- Every title must contain at least one searchable proper noun or established term people actually type (a tool, standard, or named concept — e.g. OpenTelemetry, Crossplane, MCP, Dockerfile, "self-hosting AI models"). Titles made entirely of abstractions ("Let's Fix That", "A Wake-Up Call") earn near-zero search traffic and build no catalog tail.
- Put the searchable term in the first half of the title.
- Never lead with an unfamiliar product name as the hook. If the audience does not already run the product, lead with the problem and let the product follow in context.

CRITICAL: Front-load the hook in the first 60 characters (mobile truncation point). Prioritize these proven high-performing patterns (roughly in order of measured subscriber conversion):

1. Provocative Opinion + Technical Specificity (3-5x avg views)
   - Include strong stance or controversial angle
   - Use qualifiers: "Maybe", "Brutally Honest", rhetorical questions
   - Example: "Why I Changed My Mind About Backstage? A Brutally Honest Opinion"

2. "Top N" Lists with Year Specificity (highest view counts)
   - Use numbered lists (Top 5, Top 10)
   - Include current year (2025)
   - Use imperative language: "MUST Use", "Essential"
   - Example: "Top 10 DevOps Tools You MUST Use in 2025!"

3. Direct Comparisons ("X vs Y") (35% higher watch time)
   - Compare 2-3 specific tools/technologies
   - Add compelling modifier: "Showdown!", "Is It Time to Ditch X?"
   - Example: "Cursor vs. GitHub Copilot: AI Coding Showdown!"

4. Challenge/Disruptive Statement + Solution (2.8% engagement vs 1.9% avg)
   - Use: "Stop Using...", "Say Goodbye to...", "Forget..."
   - Present immediate alternative
   - Example: "Stop Using Docker for Dev Environments! (feat. Okteto)"

5. Personal Workflow/Journey (65% higher comment rate)
   - First-person perspective: "My", "How I"
   - Show specific transformation or outcome
   - Example: "My Workflow With AI: How I Code, Test, and Deploy Faster Than Ever"

AVOID these proven low-performing patterns:
- Evaluation framing of an unfamiliar product ("Is X Worth It?", "X Review", "X Magic") — 1.25-3.63 subs/1k, worst conversion and retention on the channel. Comparing tools the audience already uses is fine; introducing an unknown product via a review frame is not.
- Pure-abstraction opinion titles with no named subject (an opinion is fine, but attach it to a searchable tool/concept)
- Generic "How To" or "Using" openings (52% fewer likes per view)
- "Explained" or "Tutorial" suffix (62% lower views)
- Question-only titles without controversy (54% lower watch time)
- Overly technical without context (80% lower comment rate)
- Titles over 75 characters (8-12% CTR drop per 10 extra chars)

IMPORTANT: You must respond with ONLY a valid JSON array of strings. No explanations, no markdown formatting, no additional text. Just the JSON array.

Choose whichever patterns best fit the manuscript content. Aim for diversity across the 10 titles — use different patterns for each rather than forcing any specific one.

Example format: ["Title 1", "Title 2", "Title 3", "Title 4", "Title 5", "Title 6", "Title 7", "Title 8", "Title 9", "Title 10"]

Video Manuscript:
{{.ManuscriptContent}}

Response (JSON array only):
