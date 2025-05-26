package main

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func rollDice() int {
	return rand.IntN(99) + 1
}

func generatePasswordAndHash() (string, string, error) {
	password := randomString(40)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", "", err
	}
	return password, string(hash), nil
}

func randomString(i int) string {
	// Generate a random string of length i
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#&$*"
	b := make([]byte, i)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}
func parseDateTime(dateStr, timeStr string) (time.Time, error) {
	combined := fmt.Sprintf("%s %s", dateStr, timeStr)
	loc, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		return time.Time{}, err
	}

	// Define layout according to the format of combined string
	layout := "02-01-2006 15:04"
	return time.ParseInLocation(layout, combined, loc)
}

func prepareTable(data [][]string) string {

	// Find longest text per column
	colWidths := make([]int, len(data[0]))
	for _, row := range data {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	for _, row := range data {
		for i, cell := range row {
			format := fmt.Sprintf("%%-%ds", colWidths[i]+5) // longest cell + 5 spaces
			buf.WriteString(fmt.Sprintf(format, cell))
		}
		buf.WriteByte('\n')
	}

	result := buf.String()
	return result

}
