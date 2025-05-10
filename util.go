package main

import (
	"math/rand/v2"

	"github.com/bwmarrin/discordgo"
)

func rollDice() int {
	return rand.IntN(99) + 1
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
