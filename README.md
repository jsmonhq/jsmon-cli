# JSMon CLI

A command-line tool for interacting with the JSMon API to scan URLs, domains, and upload files.

## Installation

### Using go install (Recommended)

Install directly from the repository:

```bash
go install github.com/jsmonhq/jsmon-cli@latest
```

After installation, make sure `$GOPATH/bin` (or `$HOME/go/bin` by default) is in your `PATH`. The binary will be available as `jsmon-cli`.

### Building from source

For development or custom builds:

```bash
git clone https://github.com/jsmonhq/jsmon-cli.git
cd jsmon-cli
go build -o jsmon-cli
```

## Usage

### Prerequisites

You need a JSMon API key. The priority order is:
1. `-key` flag (highest priority)
2. `~/.jsmon/credentials` file
3. `JSMON_API_KEY` environment variable

For scanning operations, you also need a workspace ID (must be provided via command or env var):
- Pass it via the `-wksp` flag
- Set it as an environment variable: `JSMON_WORKSPACE_ID`

**Note:** Workspace ID is NOT read from the credentials file and must be provided in the command.

#### Credentials File

Create `~/.jsmon/credentials` with your API key written directly in the file:

```
your-api-key-here
```

The file should contain only the API key (first non-empty, non-comment line). Lines starting with `#` are treated as comments and empty lines are ignored.

**Note:** Only the API key is stored in the credentials file. Workspace ID must be provided via `-wksp` flag or environment variable.

### Commands

#### Create a Workspace

```bash
jsmon-cli -cw "My Workspace" -key <your-api-key>
# or
jsmon-cli --create-workspace "My Workspace" -key <your-api-key>
```

#### Upload a URL for Scanning

```bash
jsmon-cli -u "https://example.com/app.js" -wksp <workspace-id> -key <your-api-key>
```

#### Scan a Domain

```bash
jsmon-cli -d "example.com" -wksp <workspace-id> -key <your-api-key>
```

#### Upload a File of URLs

The file should contain one URL per line:

```bash
jsmon-cli -f urls.txt -wksp <workspace-id> -key <your-api-key>
```

### Environment Variables

You can set the API key via environment variable (lowest priority after flag and credentials file):

```bash
export JSMON_API_KEY="your-api-key"
```

For workspace ID, you can set it via environment variable:

```bash
export JSMON_WORKSPACE_ID="your-workspace-id"
```

Then you can use the tool without the `-wksp` flag (API key will be read from credentials file if set):

```bash
jsmon-cli -u "https://example.com/app.js"
jsmon-cli -d "example.com"
jsmon-cli -f urls.txt
```

## Examples

```bash
# Create a workspace
jsmon-cli -cw "My Project" -key abc123

# Upload a single URL
jsmon-cli -u "https://example.com/script.js" -wksp 30319f3d-8d35-42db-afb7-e2bd7f8f7fb1 -key abc123

# Scan a domain
jsmon-cli -d "example.com" -wksp 30319f3d-8d35-42db-afb7-e2bd7f8f7fb1 -key abc123

# Upload multiple URLs from a file
jsmon-cli -f urls.txt -wksp 30319f3d-8d35-42db-afb7-e2bd7f8f7fb1 -key abc123

# Using credentials file (create ~/.jsmon/credentials first)
jsmon-cli -u "https://example.com/script.js"
jsmon-cli -d "example.com"
jsmon-cli -f urls.txt
```

