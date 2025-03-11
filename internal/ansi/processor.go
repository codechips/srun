package ansi

import (
    "regexp"
)

// Matches ANSI escape sequences for stripping them from plain text
var ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")

type ProcessedLine struct {
    Raw     string  // Original line with ANSI codes
    Plain   string  // Line with ANSI codes stripped (for logging)
}

// Process preserves raw ANSI output and provides a clean version for logging
func Process(line string) ProcessedLine {
    // For plain text, replace both ANSI codes and carriage returns
    plain := ansiRegex.ReplaceAllString(line, "")
    plain = regexp.MustCompile("\r").ReplaceAllString(plain, "")

    return ProcessedLine{
        Raw:   line,     // Keep original line with ANSI codes and carriage returns
        Plain: plain,    // Strip both ANSI codes and carriage returns for logging
    }
}
