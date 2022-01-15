package util

import (
	"fmt"
	"strconv"
	"strings"
)

func PinArrayToString(pinArray [8]uint) (string, error) {
	str := strings.Builder{}
	for _, digit := range pinArray {
		if digit > 9 {
			return "", fmt.Errorf("each pin digit must be greater than 0 and less than 9")
		}
		_, _ = str.WriteString(strconv.Itoa(int(digit)))
	}
	return str.String(), nil
}
