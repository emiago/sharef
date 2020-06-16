package errx

import "fmt"

func Wrapf(err error, format string, args ...interface{}) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
