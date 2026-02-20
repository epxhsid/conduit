package protoc

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Protocol struct{}

func Write(w io.Writer, domain string, port int) error {
	_, err := fmt.Fprintf(w, "REGISTER %s %d\n", domain, port)
	return err
}

func Read(r io.Reader) (domain string, port int, err error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return "", 0, fmt.Errorf("failed to read from reader: %v", scanner.Err())
	}

	parts := strings.Fields(scanner.Text())
	if len(parts) != 3 || parts[0] != "REGISTER" {
		return "", 0, fmt.Errorf("invalid protocol message: %s", scanner.Text())
	}

	port, err = strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port number: %v", err)
	}

	return parts[1], port, nil
}

func OK(w io.Writer) error {
	_, err := fmt.Fprintln(w, "OK")
	return err
}

func Error(w io.Writer, err error) error {
	_, writeErr := fmt.Fprintf(w, "ERROR %v\n", err)
	return writeErr
}
