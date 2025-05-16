package validators

import "regexp"

func IsUsernameValid(username string) bool {
	var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_\-]{3,30}$`)
	return validUsername.MatchString(username)
}
