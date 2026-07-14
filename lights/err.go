package lights

import (
	"errors"
	"fmt"
	"log"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func LogError(source string, err error) error {
	if err != nil {
		log.Printf("ERROR - %s: %v", source, err)
	}
	return err
}

var (
	NoBytes  error = errors.New("no bytes to show")
	NoColors error = errors.New("no colors to show")
)

func LogNoColors(source string) error {
	return LogError(source, NoColors)
}
func LogNoBytes(source string) error {
	return LogError(source, NoBytes)
}

func LogLengthMisMatch(source string, expected int, got int) error {
	err := fmt.Errorf("frame size mismatch expected (%d) got (%d)",
		expected, got)
	return LogError(source, err)
}
