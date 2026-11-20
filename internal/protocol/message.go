package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Errors returned by parse functions
var (
	ErrUnknownMessage = errors.New("protocol: unknown message type")
	ErrMalformedError = errors.New("protocol: malformed ERR line")
	ErrMalformedEvent = errors.New("protocol: malformed EVT line")
)

// Message is a parsed server line. Only the fields matching Kind are set.
type Message struct {
	Kind Kind

	Data string // OK: everything that follows the OK (may be empty)

	Code   int    // ERR: e.g. 404
	ErrMsg string // ERR: e.g. "NAME_IN_USE"

	Category string // EVT: e.g. GLOBAL
	Type     string // EVT: e.g. CHAT
	Payload  string // EVT: the rest of the line, e.g. "alice hello"
}

// ParseLine parses a line received from the server into a Message.
// It returns the error from ValidateLine, or ErrUnknownMessage, ErrMalformedError or ErrMalformedEvent.
func ParseLine(line string) (Message, error) {
	line = strings.TrimRight(line, "\n\r")

	if err := ValidateLine(line); err != nil {
		return Message{}, err
	}

	// "OK ...", "ERR <code> <msg>", "EVT <category> [<type>] <payload>"
	head, data, _ := strings.Cut(line, " ")
	switch Kind(head) {
	case KindOK:
		return Message{
			Kind: KindOK,
			Data: data,
		}, nil

	case KindEvent:
		category, rest, found := strings.Cut(data, " ")
		if !found || category == "" {
			return Message{}, ErrMalformedEvent
		}
		if category == ScopeStats {
			if rest == "" {
				return Message{}, ErrMalformedEvent
			}
			return Message{
				Kind:     KindEvent,
				Category: category,
				Payload:  rest,
			}, nil
		}
		typ, payload, found := strings.Cut(rest, " ")
		if !found || typ == "" || payload == "" {
			return Message{}, ErrMalformedEvent
		}
		return Message{
			Kind:     KindEvent,
			Category: category,
			Type:     typ,
			Payload:  payload,
		}, nil

	case KindErr:
		codeStr, msg, found := strings.Cut(data, " ")
		if !found || len(codeStr) != 3 || msg == "" {
			return Message{}, ErrMalformedError
		}
		code, err := strconv.Atoi(codeStr)
		// a sign ("+12") uses one of the 3 chars
		// so signed values are < 100 and rejected
		if err != nil || code < 100 {
			return Message{}, ErrMalformedError
		}
		return Message{
			Kind:   KindErr,
			Code:   code,
			ErrMsg: msg,
		}, nil

	default:
		return Message{}, ErrUnknownMessage
	}
}

// ParseKV parses space-separated key=value pairs, as found in OK responses
// and STATS events, into a map. Words without "=" and pairs with an empty
// key are skipped. An empty value is kept, and for a duplicate key the last
// value wins. It never returns nil.
//
//	ParseKV("hello proto=1") // {"proto": "1"}
//	ParseKV("a=1 a=2")       // {"a": "2"}
func ParseKV(payload string) map[string]string {
	kv := make(map[string]string)

	for word := range strings.FieldsSeq(payload) {
		key, value, found := strings.Cut(word, "=")
		// skip words without "=" and pairs with an empty key
		if !found || key == "" {
			continue
		}
		kv[key] = value
	}
	return kv
}

// ParsePresence splits a PRESENCE payload such as "ENTER alice" into its action and username.
// It returns an error if:
//   - action or username is empty
//   - username is invalid
//   - action is unknown
func ParsePresence(payload string) (action, name string, err error) {
	action, name, _ = strings.Cut(payload, " ")
	if err := ValidateUsername(name); err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrMalformedEvent, err)
	}
	switch action {
	case PresenceEnter, PresenceLeave:
		return action, name, nil
	default:
		return "", "", ErrMalformedEvent
	}
}
