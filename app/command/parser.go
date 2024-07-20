package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/log"
	"go.uber.org/zap"
)

const (
	Delimeter      = "\r\n"
	NullBulkString = "$-1\r\n"
	OKString       = "+OK\r\n"
)

type CommandParser struct {
	tokens []string
	curIdx int
	logger log.Logger
}

func NewParser(cmd string, logger log.Logger) (CommandParser, error) {
	cmdTokens := strings.Split(
		strings.TrimSuffix(cmd, Delimeter),
		Delimeter,
	)

	return CommandParser{
		logger: logger,
		curIdx: 0,
		tokens: cmdTokens,
	}, nil
}

func (parser *CommandParser) Parse() ([]Command, error) {
	parsedArrays := [][]any{}
	for {
		if parser.remainingTokens() == 0 {
			break
		}

		parsedArray, err := parser.parseArray()
		if err != nil {
			return nil, fmt.Errorf("error parsing command to array: %w", err)
		}
		parsedArrays = append(parsedArrays, parsedArray)
	}

	commands := make([]Command, 0, len(parsedArrays))
	for _, arr := range parsedArrays {
		command, err := ToCommand(arr)
		if err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}

	return commands, nil
}

func (parser *CommandParser) parseNext() (any, error) {
	token, err := parser.peekNextToken()
	if err != nil {
		return nil, err
	}
	parser.logger.Debug("parsing token", zap.String("token", token))

	if len(token) == 0 {
		// TODO: This might actualy be legal in some cases in which case maybe make this a err var
		return nil, fmt.Errorf("found an empty token while parsing command")
	}

	switch []rune(token)[0] {
	case '+':
		return parser.parseSimpleString()
	case '-':
		return nil, errors.New("TODO: Implement parser.parseSimpleError()")
	case ':':
		return parser.parseInt()
	case '$':
		return parser.parseBulkString()
	case '*':
		return parser.parseArray()
	case '_':
		return nil, errors.New("TODO: Implement parser.parseNull()")
	case '#':
		return nil, errors.New("TODO: Implement parser.parseBool()")
	case ',':
		return nil, errors.New("TODO: Implement parser.parseDouble()")
	case '(':
		return nil, errors.New("TODO: Implement parser.parseBigNumber()")
	case '!':
		return nil, errors.New("TODO: Implement parser.parseBulkError()")
	case '=':
		return nil, errors.New("TODO: Implement parser.parseVerbatimString()")
	case '%':
		return nil, errors.New("TODO: Implement parser.parseMap()")
	case '~':
		return nil, errors.New("TODO: Implement parser.parseSet()")
	case '>':
		return nil, errors.New("TODO: Implement parser.parsePush()")
	}

	return nil, fmt.Errorf("expected to parse an identified element, but got %q", token)
}

var OutOfTokensErr = errors.New("tried to access the next token, but no more tokens remain")

func (parser CommandParser) peekNextToken() (string, error) {
	parser.logger.Debug(
		"peeking token while parsing command",
		zap.Any("tokens", parser.tokens),
		zap.Int("index", parser.curIdx),
		zap.Int("totalSize", len(parser.tokens)),
	)
	if parser.remainingTokens() == 0 {
		return "", fmt.Errorf("tried to get element %d with %d tokens: %w", parser.curIdx+1, len(parser.tokens), OutOfTokensErr)
	}
	token := parser.tokens[parser.curIdx]
	return token, nil
}

func (parser *CommandParser) popNextToken() (string, error) {
	parser.logger.Debug(
		"popping token while parsing command",
		zap.Any("tokens", parser.tokens),
		zap.Int("index", parser.curIdx),
		zap.Int("totalSize", len(parser.tokens)),
	)
	token, err := parser.peekNextToken()
	if err != nil {
		return "", err
	}

	parser.curIdx++
	return token, nil
}

func (parser CommandParser) remainingTokens() int {
	return len(parser.tokens) - parser.curIdx
}

func (parser *CommandParser) parseSimpleString() (string, error) {
	token, err := parser.popNextToken()
	if err != nil {
		return "", fmt.Errorf("error getting next token while parsing simple string: %w", err)
	}
	parsedString, ok := strings.CutPrefix(token, "+")
	if !ok {
		return "", fmt.Errorf("next token did not have the expected prefix '+' while parsing simple string: %w", err)
	}

	return parsedString, err
}

func (parser *CommandParser) parseInt() (int64, error) {
	token, err := parser.popNextToken()
	if err != nil {
		return 0, fmt.Errorf("error parsing integer: %w", err)
	}
	parsedInt, err := parseIntWithPrefix(token, ":")

	return parsedInt, err
}

func (parser *CommandParser) parseBulkString() (string, error) {
	lengthToken, err := parser.popNextToken()
	if err != nil {
		return "", fmt.Errorf("error parsing size of bulk string: %w", err)
	}
	dataToken, err := parser.popNextToken()
	if err != nil {
		return "", fmt.Errorf("error getting next token from bulk string: %w", err)
	}

	length, err := parseIntWithPrefix(lengthToken, "$")
	if err != nil {
		return "", err
	}

	// If the length of the data is less than the expected length, it's possible we split at a delimiter that was part
	// of the string so add it back and add the next element until we have reached or exceeded the size
	for int64(len(dataToken)) < length {
		additionalData, err := parser.popNextToken()
		if err != nil {
			// There was some data, but not enough to make the full length string
			return "", fmt.Errorf("length of %d does not match the length of the provided data %q", length, dataToken)
		}
		dataToken = fmt.Sprint(dataToken, Delimeter, additionalData)
	}

	if int64(len(dataToken)) > length {
		return "", fmt.Errorf("length of %d does not match the length of the provided data %q", length, dataToken)
	}

	return dataToken, nil
}

func (parser *CommandParser) parseArray() ([]any, error) {
	token, err := parser.popNextToken()
	if err != nil {
		return nil, fmt.Errorf("error parsing array size: %w", err)
	}
	numElems, err := parseIntWithPrefix(token, "*")
	if err != nil {
		return nil, fmt.Errorf("error parsing array size to an integer: %w", err)
	}

	parsedArray := make([]any, 0, numElems)
	for range numElems {
		nextElem, err := parser.parseNext()
		if err != nil {
			return nil, fmt.Errorf("error parsing array element: %w", err)
		}
		parsedArray = append(parsedArray, nextElem)
	}

	return parsedArray, nil
}
