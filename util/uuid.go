package util

import "regexp"

const UUIDRegexp = "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}"

var uuidRegexp = regexp.MustCompile(UUIDRegexp)

func IsUUID(s string) bool {
	return uuidRegexp.MatchString(s)
}
