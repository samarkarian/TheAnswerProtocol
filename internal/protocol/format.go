package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func FormatOK(data string) (string, error) {

	if data == "" {
		return "OK\n", nil
	}

	s := fmt.Sprintf("OK %s", data)
	err := ValidateLine(s)
	if err != nil {
		return "", err
	}

	return s + "\n", nil
}

func FormatOKJSON(v any) (string, error) {

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(v)
	if err != nil {
		return "", err
	}

	s := "OK " + strings.TrimSuffix(buffer.String(), "\n")
	err = ValidateLine(s)
	if err != nil {
		return "", err
	}

	return s + "\n", nil
}

func FormatErr(e ErrorCode) string {

	return fmt.Sprintf("ERR %d %s\n", e.Code, e.Message)
}

func FormatEvent(cat, typ, payload string) (string, error) {

	slc := []string{cat, typ, payload}
	newSlc := []string{}

	if cat == "" || payload == "" {
		return "", ErrMalformedEvent
	}
	if cat == ScopeStats && typ != "" {
		return "", ErrMalformedEvent
	}
	if cat != ScopeStats && typ == "" {
		return "", ErrMalformedEvent
	}
	if strings.Contains(cat, " ") || strings.Contains(typ, " ") {
		return "", ErrMalformedEvent
	}

	for _, s := range slc {
		if s != "" {
			newSlc = append(newSlc, s)
		}
	}

	s := strings.Join(newSlc, " ")
	s = fmt.Sprintf("EVT %s", s)
	err := ValidateLine(s)
	if err != nil {
		return "", err
	}

	return s + "\n", nil
}
