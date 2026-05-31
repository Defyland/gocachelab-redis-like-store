package protocol

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func SimpleString(value string) []byte {
	return []byte("+" + sanitizeLine(value) + "\r\n")
}

func Error(message string) []byte {
	return []byte("-ERR " + sanitizeLine(message) + "\r\n")
}

func Integer(value int64) []byte {
	return []byte(":" + strconv.FormatInt(value, 10) + "\r\n")
}

func BulkString(value string) []byte {
	return []byte("$" + strconv.Itoa(len(value)) + "\r\n" + value + "\r\n")
}

func NullBulkString() []byte {
	return []byte("$-1\r\n")
}

func Array(values []string) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "*%d\r\n", len(values))
	for _, value := range values {
		b.Write(BulkString(value))
	}
	return b.Bytes()
}

func sanitizeLine(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
