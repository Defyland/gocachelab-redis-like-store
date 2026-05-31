package protocol

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseLineSupportsInlineCommandsAndQuotes(t *testing.T) {
	cmd, err := ParseLine("set user:1 \"Ada Lovelace\"\r\n")
	if err != nil {
		t.Fatalf("ParseLine returned error: %v", err)
	}
	if cmd.Name != "SET" {
		t.Fatalf("command name = %q, want SET", cmd.Name)
	}
	if !reflect.DeepEqual(cmd.Args, []string{"user:1", "Ada Lovelace"}) {
		t.Fatalf("args = %#v", cmd.Args)
	}
}

func TestEncodeCommandRoundTripsEscapedArguments(t *testing.T) {
	line := EncodeCommand("set", "key with spaces", "line\nwith\tchars")
	cmd, err := ParseLine(line)
	if err != nil {
		t.Fatalf("ParseLine returned error: %v", err)
	}
	if cmd.Name != "SET" {
		t.Fatalf("command = %q", cmd.Name)
	}
	if !reflect.DeepEqual(cmd.Args, []string{"key with spaces", "line\nwith\tchars"}) {
		t.Fatalf("args = %#v", cmd.Args)
	}
}

func TestParseLineRejectsMalformedInput(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"SET \"unterminated",
		"SET \"quoted\"tail",
	}

	for _, input := range tests {
		if _, err := ParseLine(input); err == nil {
			t.Fatalf("ParseLine(%q) returned nil error", input)
		}
	}
}

func TestParseLineRejectsRESPForNow(t *testing.T) {
	_, err := ParseLine("*1\r\n$4\r\nPING\r\n")
	if !errors.Is(err, ErrRESPCommand) {
		t.Fatalf("error = %v, want ErrRESPCommand", err)
	}
}
