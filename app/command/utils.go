package command

import (
	"fmt"
	"strconv"
	"strings"
)

func parseIntWithPrefix(intStr string, prefix string) (int, error) {
	trimmedStr, found := strings.CutPrefix(intStr, prefix)
	if !found {
		return 0, fmt.Errorf("expected to find prefix %q when parsing string %q", prefix, intStr)
	}
	return strconv.Atoi(trimmedStr)
}
