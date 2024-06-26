package command

import (
	"fmt"
	"strings"
)

// TODO: As this grows in complexity, it may be worth thinking about restructing these encoder/decoder bits
func Encode(data any) (string, error) {
	switch typedData := data.(type) {
	// case []byte: // TODO: Not sure if this needs to be handled
	// 	return encodeBulkString(typedData)
	case []any:
		return encodeArray(typedData)
	default:
		return encodePrimitive(typedData)
	}
}

func encodeArray(arrayData []any) (string, error) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("*%d%s", len(arrayData), Delimeter))
	for _, data := range arrayData {
		res, err := Encode(data)
		if err != nil {
			return "", fmt.Errorf("failed to encode list element: %w", err)
		}
		builder.WriteString(res)
		builder.WriteString(Delimeter)
	}
	builder.WriteString(Delimeter)

	return builder.String(), nil
}

func encodePrimitive(data any) (string, error) {
	var result string
	var err error
	switch typedData := data.(type) {
	case int64:
		result, err = encodeInt(typedData)
	case string:
		result, err = encodeString(typedData)
	default:
		return "", fmt.Errorf("tried to encode primitive data of an unknown type %[1]T: %[1]v", data)
	}
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", result, Delimeter), nil
}

func encodeInt(data int64) (string, error) {
	return fmt.Sprintf(":%d", data), nil
}

func encodeString(data string) (string, error) {
	return fmt.Sprintf("+%s", data), nil
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
