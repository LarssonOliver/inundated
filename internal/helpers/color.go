package helpers

import "regexp"

var colorRegexp = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)

func IsValidColor(color string) bool {
	return colorRegexp.MatchString(color)
}
