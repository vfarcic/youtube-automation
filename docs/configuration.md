# Configuring a Video Upload

This document outlines the various configuration options available for customizing video uploads.

## Global Default Settings

You can set global default languages for your videos. These will be applied if no specific language is set for an individual video.

These settings can be configured in two ways:

1.  **Via `settings.yaml` file:**
    Add or modify the `videoDefaults` section in your `settings.yaml` file located in the project root.

    ```yaml
    # settings.yaml
    videoDefaults:
      language: "en"         # Default language for video metadata (e.g., title, description)
      audioLanguage: "es"    # Default language for the audio track
    ```

2.  **Via Command-Line Flags:**
    You can override the `settings.yaml` values or set them if the file doesn't exist using command-line flags when running the application:

    *   `--video-defaults-language <ISO_CODE>`: Sets the default video language.
        *Example: `--video-defaults-language fr`*
    *   `--video-defaults-audio-language <ISO_CODE>`: Sets the default video audio language.
        *Example: `--video-defaults-audio-language de`*

    If neither `settings.yaml` nor command-line flags specify these values, they will default to `"en"` (English).

## Per-Video Language Overrides

You can specify the language and audio language for each video individually by adding fields to your video metadata file (e.g., `video.yaml`). These per-video settings will always take precedence over the global defaults.

```yaml
# Example: my_video_metadata.yaml
title: "My Awesome Video"
description: "A great video about coding."
# ... other metadata ...
language: "ja"           # Specific language for this video's metadata
audioLanguage: "ko"      # Specific audio language for this video's track
```

If `language` or `audioLanguage` are omitted or left empty in the video's metadata, the globally configured default values (from `settings.yaml` or flags) will be used.

## Language Codes

The language settings (both global and per-video) affect the `defaultLanguage` and `defaultAudioLanguage` properties of your video on YouTube. This helps YouTube categorize and recommend your content appropriately.

You should use **ISO 639-1 language codes** (e.g., "en" for English, "es" for Spanish, "fr" for French).
For a comprehensive list of ISO 639-1 codes, you can refer to resources like the [Library of Congress ISO 639-1 Registration Authority](https://www.loc.gov/standards/iso639-2/php/code_list.php) or other official sources. 