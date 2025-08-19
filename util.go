package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/crypto/bcrypt"
)

// rollDice returns a random number between 1 and 100 inclusive.
func rollDice() int {
	return rand.IntN(100) + 1
}

// getMemberGuildNick retrieves the member's display name, falling back to their
// username if a guild nickname is not set.
func getMemberGuildNick(m *discordgo.Member) string {
	name := m.DisplayName()
	if name == "" {
		name = m.User.Username
	}
	return name
}

// GetOptions converts a slice of command options into a map keyed by option
// name for easier lookup.
func GetOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap
}

// generatePasswordAndHash returns a randomly generated password and its bcrypt
// hash.
func generatePasswordAndHash() (string, string, error) {
	password := randomString(40)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", "", err
	}
	return password, string(hash), nil
}

// randomString generates a random string of the specified length using a set of
// alphanumeric and symbol characters.
func randomString(i int) string {
	// Generate a random string of length i
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#&$*"
	b := make([]byte, i)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}

// parseDateTime parses the provided date (DD-MM-YYYY) and time (HH:MM) strings
// into a time.Time value in the Europe/Prague location.
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

// stringPtr returns a pointer to the provided string.
func stringPtr(s string) *string {
	return &s
}

// contains reports whether item is present in slice.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
