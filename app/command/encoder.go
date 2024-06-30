package command

import (
	"fmt"
	"strings"
)

type Encoder struct {
	UseBulkStrings bool
}

// TODO: As this grows in complexity, it may be worth thinking about restructing these encoder/decoder bits
func (e Encoder) Encode(data any) (string, error) {
	switch typedData := data.(type) {
	// case []byte: // TODO: Not sure if this needs to be handled serperately (as a bulk string)
	// 	return encodeBulkString(typedData)
	case []any:
		return e.encodeArray(typedData)
	default:
		return e.encodePrimitive(typedData)
	}
}

// MustEncode calls encode, but panics if the error is not nil
func (e Encoder) MustEncode(data any) string {
	res, err := e.Encode(data)
	if err != nil {
		panic(err)
	}
	return res
}

func (e Encoder) encodeArray(arrayData []any) (string, error) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("*%d%s", len(arrayData), Delimeter))
	for _, data := range arrayData {
		res, err := e.Encode(data)
		if err != nil {
			return "", fmt.Errorf("failed to encode list element: %w", err)
		}
		builder.WriteString(res)
	}

	return builder.String(), nil
}

func (e Encoder) encodePrimitive(data any) (string, error) {
	var result string
	var err error
	switch typedData := data.(type) {
	case int:
		result, err = encodeInt(typedData)
	case string:
		if e.UseBulkStrings {
			result, err = encodeBulkString(typedData)
		} else {
			result, err = encodeString(typedData)
		}
	case bool:
		result, err = encodeBool(typedData)
	default:
		return "", fmt.Errorf("tried to encode primitive data of an unknown type %[1]T: %[1]v", data)
	}
	if err != nil {
		return "", err
	}

	return fmt.Sprint(result, Delimeter), nil
}

func encodeInt(data int) (string, error) {
	return fmt.Sprintf(":%d", data), nil
}

func encodeBulkString(data string) (string, error) {
	return fmt.Sprintf("$%d\r\n%s", len(data), data), nil
}

func encodeString(data string) (string, error) {
	return fmt.Sprintf("+%s", data), nil
}

func encodeBool(data bool) (string, error) {
	if data {
		return "#t", nil
	}
	return "#f", nil
}

//  case '+':
// 		return nil, errors.New("TODO: Implement parser.parseSimpleString()")
// 	case '-':
// 		return nil, errors.New("TODO: Implement parser.parseSimpleError()")
// 	case ':':
// 		return parser.parseInt()
// 	case '$':
// 		return parser.parseBulkString()
// 	case '*':
// 		return parser.parseArray()
// 	case '_':
// 		return nil, errors.New("TODO: Implement parser.parseNull()")
// 	case '#':
// 		return nil, errors.New("TODO: Implement parser.parseBool()")
// 	case ',':
// 		return nil, errors.New("TODO: Implement parser.parseDouble()")
// 	case '(':
// 		return nil, errors.New("TODO: Implement parser.parseBigNumber()")
// 	case '!':
// 		return nil, errors.New("TODO: Implement parser.parseBulkError()")
// 	case '=':
// 		return nil, errors.New("TODO: Implement parser.parseVerbatimString()")
// 	case '%':
// 		return nil, errors.New("TODO: Implement parser.parseMap()")
// 	case '~':
// 		return nil, errors.New("TODO: Implement parser.parseSet()")
// 	case '>':
// 		return nil, errors.New("TODO: Implement parser.parsePush()")
// 	}
