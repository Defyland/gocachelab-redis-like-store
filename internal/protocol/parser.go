package protocol

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Command struct {
	Name string
	Args []string
}

var (
	ErrEmptyCommand = errors.New("empty command")
	ErrRESPCommand  = errors.New("RESP protocol is not supported by this build")
)

func ParseLine(line string) (Command, error) {
	line = strings.TrimRight(line, "\r\n")
	if strings.TrimSpace(line) == "" {
		return Command{}, ErrEmptyCommand
	}
	if strings.HasPrefix(strings.TrimLeftFunc(line, unicode.IsSpace), "*") {
		return Command{}, ErrRESPCommand
	}

	parts, err := parseInlineFields(line)
	if err != nil {
		return Command{}, err
	}
	if len(parts) == 0 {
		return Command{}, ErrEmptyCommand
	}

	return Command{
		Name: strings.ToUpper(parts[0]),
		Args: parts[1:],
	}, nil
}

func parseInlineFields(line string) ([]string, error) {
	fields := make([]string, 0, 4)
	for i := 0; i < len(line); {
		for i < len(line) && isSpace(line[i]) {
			i++
		}
		if i >= len(line) {
			break
		}

		if line[i] == '"' {
			value, next, err := parseQuotedField(line, i+1)
			if err != nil {
				return nil, err
			}
			fields = append(fields, value)
			i = next
			if i < len(line) && !isSpace(line[i]) {
				return nil, fmt.Errorf("quoted argument must be followed by whitespace")
			}
			continue
		}

		start := i
		for i < len(line) && !isSpace(line[i]) {
			if line[i] == '"' {
				return nil, fmt.Errorf("unexpected quote in unquoted argument")
			}
			i++
		}
		fields = append(fields, line[start:i])
	}
	return fields, nil
}

func parseQuotedField(line string, start int) (string, int, error) {
	var b strings.Builder
	for i := start; i < len(line); i++ {
		switch line[i] {
		case '"':
			return b.String(), i + 1, nil
		case '\\':
			if i+1 >= len(line) {
				return "", 0, fmt.Errorf("unfinished escape sequence")
			}
			i++
			switch line[i] {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '\\', '"':
				b.WriteByte(line[i])
			default:
				return "", 0, fmt.Errorf("unsupported escape sequence \\%c", line[i])
			}
		default:
			b.WriteByte(line[i])
		}
	}
	return "", 0, fmt.Errorf("unterminated quoted argument")
}

func EncodeCommand(name string, args ...string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, strings.ToUpper(name))
	for _, arg := range args {
		parts = append(parts, encodeArg(arg))
	}
	return strings.Join(parts, " ")
}

func encodeArg(arg string) string {
	if arg == "" {
		return `""`
	}
	needsQuote := false
	for i := 0; i < len(arg); i++ {
		if isSpace(arg[i]) || arg[i] == '"' || arg[i] == '\\' || arg[i] < 0x20 || arg[i] == 0x7f {
			needsQuote = true
			break
		}
	}
	if !needsQuote {
		return arg
	}

	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(arg); i++ {
		switch arg[i] {
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		default:
			b.WriteByte(arg[i])
		}
	}
	b.WriteByte('"')
	return b.String()
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}
