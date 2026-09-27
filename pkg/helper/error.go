package helper

import "fmt"

func CompositeError(text string, err error) error {
	return fmt.Errorf(text+": %w", err)
}
