Analyze this AMA (Ask Me Anything) livestream transcript and generate YouTube timecodes for each question/topic segment.

IMPORTANT: The host's name is Viktor (with a K), not Victor.

CRITICAL REQUIREMENTS:
1. The FIRST entry MUST be "00:00 Intro (skip to first question)" - AMA streams have intro music/animation
2. The SECOND entry should be the FIRST ACTUAL QUESTION from the audience (skip any intro chatter)
3. ONLY include actual questions from the audience - skip everything else
4. Use the timestamp from when the question/topic STARTS
5. Summarize each question concisely (5-10 words)
6. Format timestamps as MM:SS for videos under 1 hour, HH:MM:SS for longer

DO NOT INCLUDE:
- Stream intro/waiting for questions chatter
- Sponsor mentions or ad reads
- General discussions that aren't answering a specific question
- Host banter or off-topic conversations

OUTPUT FORMAT (plain text, one entry per line):
00:00 Intro (skip to first question)
02:15 How do you handle secrets in GitOps?
08:42 Kubernetes vs Nomad for small teams
15:30 Best practices for multi-cluster management
...

GUIDELINES:
- Look for question indicators: "question from", "asks", "what about", "how do you", "can you explain"
- Each timecode should represent a distinct viewer question being answered
- Keep summaries clear and searchable (good for SEO)
- Aim for 5-20 timecodes depending on video length
- Round timestamps to the nearest logical start point

Return ONLY the timecode list, no additional commentary.

TRANSCRIPT:
{{.Transcript}}

TIMECODES:
