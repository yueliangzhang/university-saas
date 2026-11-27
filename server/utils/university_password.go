package utils

import (
	"errors"
	"unicode/utf8"
)

var ErrUniversityAccountPasswordLength = errors.New("密码须为6至12个字符")

func ValidateUniversityAccountPassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < 6 || length > 12 {
		return ErrUniversityAccountPasswordLength
	}
	return nil
}
