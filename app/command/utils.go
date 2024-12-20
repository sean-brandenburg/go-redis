package command

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseIntWithPrefix(intStr string, prefix string) (int64, error) {
	trimmedStr, found := strings.CutPrefix(intStr, prefix)
	if !found {
		return 0, fmt.Errorf("expected to find prefix %q when parsing string %q", prefix, intStr)
	}
	return strconv.ParseInt(trimmedStr, 10, 64)
}
