Analyze the following manuscript and identify {{.CandidateCount}} segments that would make excellent YouTube Shorts.

REQUIREMENTS FOR EACH SEGMENT:
1. **Self-contained**: Must make sense without additional context
2. **High-impact**: Contains a surprising insight, strong opinion, practical tip, or memorable statement
3. **Word count**: Maximum {{.MaxWords}} words
4. **Standalone value**: Viewer should gain something even without watching the full video

IDEAL SHORT TYPES (prioritize these):
- **Hot takes**: Controversial or surprising opinions
- **Aha moments**: Key insights that challenge conventional thinking
- **Practical tips**: Actionable advice viewers can use immediately
- **Myth busters**: Correcting common misconceptions

AVOID THESE:
- Segments that require prior context from the video
- Long explanations or tutorials
- Content that references "earlier" or "later" in the video
- Introductions or conclusions

RESPONSE FORMAT:
Return a JSON array with exactly {{.CandidateCount}} candidates:
- "id": Unique identifier (short1, short2, etc.)
- "title": Catchy title (max 50 characters)
- "text": The exact text segment from the manuscript (copy verbatim)
- "rationale": Why this makes a good Short (1-2 sentences)

IMPORTANT:
- Extract the EXACT text from the manuscript - do not rewrite
- Order by quality (best candidate first)

MANUSCRIPT:
{{.ManuscriptContent}}

Respond with ONLY the JSON array:
