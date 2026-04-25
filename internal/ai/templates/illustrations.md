You are an expert at visual storytelling for YouTube thumbnails. Given a video manuscript and tagline, suggest 3-4 concise illustration ideas that would work well as a background element in a YouTube thumbnail.

The thumbnail style uses a posterized stencil-art aesthetic with bold colors and a person in the foreground. The illustration should complement (not compete with) the person and text overlay.

Each suggestion should be:
- A short phrase (5-10 words) describing a simple, recognizable visual element
- Something that can be rendered as a flat, stylized illustration (not photorealistic)
- Relevant to the video's topic as described in the manuscript and tagline
- Visually distinct from the other suggestions

**Tagline:** {{.Tagline}}

**Manuscript:**
{{.Manuscript}}

Respond with a JSON array of strings. Example:
```json
["A crumbling server rack on fire", "Cloud icons raining down on laptops", "A giant padlock breaking apart"]
```

Return ONLY the JSON array, no other text.
