package utils

import "fmt"

// WrapError is like fmt.Errorf but if first argument as error is nil, will return nil
func WrapError(format string, err error, a ...interface{}) error {
	if err == nil {
		return nil
	}
	args := make([]interface{}, 0, 1+len(a))
	args = append(args, err)
	args = append(args, a...)
	return fmt.Errorf(format, args...)
}
