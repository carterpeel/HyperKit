package errorTypes

import "strings"

func IsBtDevDown(err error) bool {
	return strings.HasSuffix(err.Error(), "Host is down")
}
