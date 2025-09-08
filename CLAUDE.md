# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Fastmask is a Go CLI tool that generates masked email addresses using the Fastmail API. It creates random, pronounceable email prefixes and interfaces with Fastmail's JMAP protocol to create masked emails for specific domains.

## Key Commands

### Build
```bash
# Build with Make (includes linting and formatting)
make

# Watch and rebuild on changes
make watch-build
```

### Test
# Run with Make (includes all quality checks)
make
```

### Quality Checks
```bash
# Run all quality checks (including vulnerability scanning, static analysis, and formatting)
make

# For deep static analysis (expensive, used in CI)
make deepcheck=t
```

### Development
```bash
# Watch and run fastmask with example.com
# Avoid running this as it will actually create a masked email every time it runs.
make watch-fastmask
```

# Format code
```bash
make
```

# Run specific quality checkers
# Just run them all every time
```bash
make
```

### Installation
```bash
GOEXPERIMENT=jsonv2 go install github.com/kensmith/fastmask/fastmask@latest
```

## Architecture

### Core Components

1. **Main Application** (`fastmask/main.go`):
   - Entry point that handles CLI arguments
   - Manages authentication with Fastmail API
   - Creates masked emails via JMAP protocol
   - Configuration loading from `~/.config/fastmask/config.json`

2. **Prefix Generation** (`fastmask/prefix.go`):
   - Generates random 5-character pronounceable prefixes
   - Uses crypto/rand for secure randomness
   - Lexicon of consonants only (bcdfghjkmnpqrstvwxyz)

3. **Build System** (`GNUmakefile` + `mk/`):
   - Complex Make-based build with parallel execution
   - Automatic tool installation and management
   - Quality checking pipeline with multiple static analyzers
   - Watch mode for development

### Key Design Decisions

- **JSON v2 Experimental**: Uses Go's experimental JSON v2 library (requires Go 1.25+)
- **JMAP Protocol**: Implements Fastmail's JMAP API for masked email creation
- **Security**: Token stored in XDG config directory with recommended permissions
- **Error Handling**: Comprehensive error messages with actionable instructions

### API Flow

1. Load token from `~/.config/fastmask/config.json`
2. Authenticate with Fastmail to get account ID and API URL
3. Generate random prefix using secure random generator
4. Create masked email via JMAP MaskedEmail/set method
5. Return formatted JSON response with prefix, domain, and email

## Important Notes

- Requires `GOEXPERIMENT=jsonv2` for all Go commands due to experimental JSON v2 usage
- Token must have masked email capability enabled in Fastmail
- The build system uses a sophisticated Make structure with automatic dependency management
- Quality checks include: govulncheck, staticcheck, errcheck, nilaway (deep mode), deadcode (deep mode)
