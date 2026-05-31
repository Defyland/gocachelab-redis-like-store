package protocol

import "testing"

func TestResponseRendering(t *testing.T) {
	tests := map[string][]byte{
		"simple": SimpleString("OK"),
		"error":  Error("bad\nline"),
		"int":    Integer(42),
		"bulk":   BulkString("Ada"),
		"null":   NullBulkString(),
		"array":  Array([]string{"a", "bb"}),
	}

	assertBytes(t, tests["simple"], "+OK\r\n")
	assertBytes(t, tests["error"], "-ERR bad line\r\n")
	assertBytes(t, tests["int"], ":42\r\n")
	assertBytes(t, tests["bulk"], "$3\r\nAda\r\n")
	assertBytes(t, tests["null"], "$-1\r\n")
	assertBytes(t, tests["array"], "*2\r\n$1\r\na\r\n$2\r\nbb\r\n")
}

func assertBytes(t *testing.T, got []byte, want string) {
	t.Helper()
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
