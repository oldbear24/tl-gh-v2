package main

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/crypto/bcrypt"
)

func rollDice() int {
	return rand.IntN(100) + 1
}

func getMemberGuildNick(m *discordgo.Member) string {
	name := m.DisplayName()
	if name == "" {
		name = m.User.Username
	}
	return name
}

func GetOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap
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
