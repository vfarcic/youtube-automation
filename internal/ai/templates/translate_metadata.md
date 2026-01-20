Translate the following YouTube video metadata from English to {{.TargetLanguage}}.

REQUIREMENTS:

1. TECHNICAL TERMS: Keep DevOps, cloud, and programming terms in English where that's common practice in the {{.TargetLanguage}}-speaking tech community (e.g., Kubernetes, Docker, CI/CD, cluster, deployment, pipeline).

2. TITLE: Translate naturally while preserving emotional tone and impact. Target 56-65 characters.

3. DESCRIPTION: Translate naturally. Preserve all URLs and hashtags exactly as-is.

4. TAGS: Translate non-technical tags, keep technical tags in English. Maintain comma-separated format within 450 characters.

5. TIMECODES: Keep timestamps exactly as-is (e.g., "0:00", "2:30"). Translate only the label text after each timestamp.

6. SHORT TITLES: If shortTitles array is present, translate each title naturally while keeping them catchy and short (max 100 characters each). Maintain the same array order.

RESPONSE FORMAT:
Return ONLY valid JSON matching the input structure. No markdown, no explanation.

INPUT:
{{.InputJSON}}

OUTPUT (JSON only):
