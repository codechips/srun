package ansi

import (
    "bytes"
    "fmt"
    "regexp"
    "strings"
)

// Common ANSI escape sequences
var (
    // Matches ANSI escape sequences
    ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*[a-zA-Z]")
    
    // Matches carriage return with or without line feed
    crRegex = regexp.MustCompile("\r\n?[^\n]")
)

type ProcessedLine struct {
    Raw       string            // Original line with ANSI codes
    Plain     string            // Line with ANSI codes stripped
    Styles    map[int][]string  // Map of position -> active styles
    Progress  *ProgressInfo     // Progress bar information if detected
}

type ProgressInfo struct {
    Percentage int
    Text       string
}

// Process handles a single line of output containing ANSI escape sequences
func Process(line string) ProcessedLine {
    result := ProcessedLine{
        Raw:    line,
        Styles: make(map[int][]string),
    }

    // Check for progress indicators first
    if strings.Contains(line, "\r") {
        // This might be a progress update
        result.Progress = detectProgress(line)
        // Keep the last line for display
        if idx := strings.LastIndex(line, "\r"); idx >= 0 {
            line = line[idx+1:]
        }
    }

    var plainBuf bytes.Buffer
    var currentPos int
    var activeStyles []string

    // Process the line character by character
    for i := 0; i < len(line); i++ {
        if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
            // Found an escape sequence
            end := i + 2
            for end < len(line) && !isAnsiEnd(line[end]) {
                end++
            }
            if end < len(line) {
                sequence := line[i : end+1]
                styles := parseAnsiSequence(sequence)
                if len(styles) > 0 {
                    activeStyles = append(activeStyles, styles...)
                    result.Styles[currentPos] = append([]string{}, activeStyles...)
                }
                i = end
                continue
            }
        }
        plainBuf.WriteByte(line[i])
        currentPos++
    }

    result.Plain = plainBuf.String()

    // If we haven't detected progress yet, try again with the plain text
    if result.Progress == nil {
        result.Progress = detectProgress(result.Plain)
    }

    return result
}

// isAnsiEnd returns true if the character is a valid ANSI sequence terminator
func isAnsiEnd(c byte) bool {
    return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// parseAnsiSequence converts an ANSI escape sequence to a list of style commands
func parseAnsiSequence(seq string) []string {
    // Strip the ESC[ prefix and end character
    content := seq[2 : len(seq)-1]
    if content == "" {
        return nil
    }

    var styles []string
    for _, code := range strings.Split(content, ";") {
        switch code {
        case "0":
            styles = []string{"reset"}
        case "1":
            styles = append(styles, "bold")
        case "3":
            styles = append(styles, "italic")
        case "4":
            styles = append(styles, "underline")
        // Add more style codes as needed
        default:
            // Handle colors
            if strings.HasPrefix(code, "3") || strings.HasPrefix(code, "4") {
                styles = append(styles, "color-"+code)
            }
        }
    }
    return styles
}

// detectProgress attempts to identify progress bar patterns in the text
func detectProgress(text string) *ProgressInfo {
    // Common progress patterns
    patterns := []struct {
        regex   *regexp.Regexp
        extract func([]string) *ProgressInfo
    }{
        // curl style: ###################################################################  100.0%
        {
            regexp.MustCompile(`([#-]*)[\s>]*\s*(\d+\.?\d*)%`),
            func(matches []string) *ProgressInfo {
                if len(matches) > 2 {
                    var percentage float64
                    fmt.Sscanf(matches[2], "%f", &percentage)
                    return &ProgressInfo{
                        Percentage: int(percentage),
                        Text:      matches[0],
                    }
                }
                return nil
            },
        },
        // pip style: 45% |████████████              | ETA:  00:01
        {
            regexp.MustCompile(`(\d+)%\s*\|[█▇▆▅▄▃▂▁ ]*\|`),
            func(matches []string) *ProgressInfo {
                if len(matches) > 1 {
                    percentage := 0
                    fmt.Sscanf(matches[1], "%d", &percentage)
                    return &ProgressInfo{
                        Percentage: percentage,
                        Text:      matches[0],
                    }
                }
                return nil
            },
        },
        // wget style: 45% [=======>      ]
        {
            regexp.MustCompile(`(\d+)%\s*\[[=>\s]*\]`),
            func(matches []string) *ProgressInfo {
                if len(matches) > 1 {
                    percentage := 0
                    fmt.Sscanf(matches[1], "%d", &percentage)
                    return &ProgressInfo{
                        Percentage: percentage,
                        Text:      matches[0],
                    }
                }
                return nil
            },
        },
    }

    for _, pattern := range patterns {
        if matches := pattern.regex.FindStringSubmatch(text); matches != nil {
            if info := pattern.extract(matches); info != nil {
                return info
            }
        }
    }

    return nil
}
