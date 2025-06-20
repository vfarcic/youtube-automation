#!/usr/bin/env nu

# Creates the MCP servers configuration file.
#
# Usage:
# > main apply mcp
# > main apply mcp --location my-custom-path.json
# > main apply mcp --location [ my-custom-path.json, another-path.json ]
# > main apply mcp --memory-file-path /custom/memory.json --anthropic-api-key XYZ --github-token ABC
# > main apply mcp --enable-playwright
#
def --env "main apply mcp" [
    --location: list<string> = [".cursor/mcp.json", ".roo/mcp.json", ".vscode/mcp.json"], # Path(s) where the MCP servers configuration file will be created.
    --memory-file-path: string = "",         # Path to the memory file for the memory MCP server. If empty, defaults to an absolute path for 'memory.json' in CWD.
    --anthropic-api-key: string = "",        # Anthropic API key for the taskmaster-ai MCP server. If empty, $env.ANTHROPIC_API_KEY is used if set.
    --github-token: string = "",             # GitHub Personal Access Token for the github MCP server. If empty, $env.GITHUB_TOKEN is used if set.
    --enable-playwright = false              # Enable Playwright MCP server for browser automation
] {
    let resolved_memory_file_path = if $memory_file_path == "" {
        (pwd) | path join "memory.json" | path expand
    } else {
        $memory_file_path
    }

    let resolved_anthropic_api_key = if $anthropic_api_key != "" {
        $anthropic_api_key
    } else if ("ANTHROPIC_API_KEY" in $env) {
        $env.ANTHROPIC_API_KEY
    } else {
        ""
    }

    let resolved_github_token = if $github_token != "" {
        $github_token
    } else if ("GITHUB_TOKEN" in $env) {
        $env.GITHUB_TOKEN
    } else {
        ""
    }

    mut mcp_servers_map = {}

    $mcp_servers_map = $mcp_servers_map | upsert "memory" {
        command: "npx",
        args: ["-y", "@modelcontextprotocol/server-memory"],
        env: {
            MEMORY_FILE_PATH: $resolved_memory_file_path
        }
    }

    $mcp_servers_map = $mcp_servers_map | upsert "context7" {
        command: "npx",
        args: ["-y", "@upstash/context7-mcp"]
    }

    if $resolved_anthropic_api_key != "" {
        $mcp_servers_map = $mcp_servers_map | upsert "taskmaster-ai" {
            command: "npx",
            args: ["-y", "--package=task-master-ai", "task-master-ai"],
            env: {
                ANTHROPIC_API_KEY: $resolved_anthropic_api_key
            }
        }
    }

    if $resolved_github_token != "" {
        $mcp_servers_map = $mcp_servers_map | upsert "github" {
            url: "https://api.githubcopilot.com/mcp/",
            headers: {
                Authorization: $"Bearer ($resolved_github_token)"
            }
        }
    }

    if $enable_playwright {
        $mcp_servers_map = $mcp_servers_map | upsert "playwright" {
            command: "npx",
            args: ["-y", "@playwright/mcp@latest"]
        }
    }

    let config_record = { mcpServers: $mcp_servers_map }

    for $output_location in $location {
        let parent_dir = $output_location | path dirname
        if not ($parent_dir | path exists) {
            mkdir $parent_dir
            print $"Created directory: ($parent_dir)"
        }
        $config_record | to json --indent 2 | save -f $output_location
        print $"MCP servers configuration file created at: ($output_location)"
    }
} 