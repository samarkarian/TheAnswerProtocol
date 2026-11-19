package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
)

func FormatOK(data string) (string, error) {

	if data == "" {
		return "OK\n", nil
	}

	if strings.Contains(data, "\r") || strings.Contains(data, "\n") {
		return "", ValidateLine(data)
	}

	return fmt.Sprintf("OK %s\n", data), nil
}

func FormatOKJSON(v any) (string, error) {

	jsonFmt, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	s := string(jsonFmt)

	return fmt.Sprintf("OK %s\n", s), nil
}

func FormatEvent(cat, typ, payload string) (string, error) {

	slc := []string{cat, typ, payload}
	newSlc := []string{}

	for _, s := range slc {
		if s != "" {
			newSlc = append(newSlc, s)
		}
	}

	for _, s := range slc {
		if strings.Contains(s, "\r") || strings.Contains(s, "\n") {
			return "", ValidateLine(s)
		}
	}

	s := strings.Join(newSlc, " ")
	return fmt.Sprintf("EVT %s\n", s), nil
}
