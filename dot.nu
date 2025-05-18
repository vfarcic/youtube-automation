#!/usr/bin/env nu

source scripts/mcp.nu

def main [] {}

def --env "main setup" [] {

    let anthropic_key_env = ($env | get --ignore-errors ANTHROPIC_API_KEY | default "")
    let anthropic_key = if ($anthropic_key_env | is-not-empty) {
        $anthropic_key_env
    } else {
        input $"(ansi yellow_bold)Anthropic API Key: (ansi reset)"
    }

    let github_token_env = ($env | get --ignore-errors GITHUB_TOKEN | default "")
    let github_token = if ($github_token_env | is-not-empty) {
        $github_token_env
    } else {
        input $"(ansi yellow_bold)GitHub Token: (ansi reset)"
    }

    main apply mcp --anthropic-api-key $anthropic_key --github-token $github_token
    
}
