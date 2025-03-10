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
    return ProcessedLine{
        Raw:   line,
        Plain: ansiRegex.ReplaceAllString(line, ""), // Strip ANSI codes for plain text logging
    }
}
