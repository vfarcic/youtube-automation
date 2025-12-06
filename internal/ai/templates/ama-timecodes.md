Analyze this AMA (Ask Me Anything) livestream transcript and generate YouTube timecodes for each question/topic segment.

CRITICAL REQUIREMENTS:
1. The FIRST entry MUST be "00:00 Intro (skip to first question)" - AMA streams have intro music/animation
2. Identify each distinct question or topic change in the transcript
3. Use the timestamp from when the question/topic STARTS
4. Summarize each question concisely (5-10 words)
5. Format timestamps as MM:SS for videos under 1 hour, HH:MM:SS for longer

OUTPUT FORMAT (plain text, one entry per line):
00:00 Intro (skip to first question)
02:15 How do you handle secrets in GitOps?
08:42 Kubernetes vs Nomad for small teams
15:30 Best practices for multi-cluster management
...

GUIDELINES:
- Look for question indicators: "question from", "asks", "what about", "how do you", "can you explain"
- Each timecode should represent a meaningful topic shift
- Keep summaries clear and searchable (good for SEO)
- Aim for 5-20 timecodes depending on video length
- Round timestamps to the nearest logical start point

Return ONLY the timecode list, no additional commentary.

TRANSCRIPT:
{{.Transcript}}

TIMECODES:
