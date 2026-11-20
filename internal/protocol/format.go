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

	err := ValidateLine(data)
	if err != nil {
		return "", err
	}

	s := fmt.Sprintf("OK %s", data)
	err = ValidateLine(s)
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

	s := fmt.Sprintf("OK %s", buffer.String())
	ValidateLine(s + "\n")
	// if len(s) > 1025 {
	// 	return "", ErrLineTooLong
	// }

	return s, nil
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

	for _, s := range newSlc {
		err := ValidateLine(s)
		if err != nil {
			return "", err
		}
	}

	s := strings.Join(newSlc, " ")
	s = fmt.Sprintf("EVT %s\n", s)
	ValidateLine(s + "\n")
	// if len(s) > 1025 {
	// 	return "", ErrLineTooLong
	// }

	return s, nil
}
