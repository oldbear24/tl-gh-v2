package main

import (
	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Sends a pong message",
	},
	{
		Name:        "roll",
		Description: "Rolls a dice",
	},
	{
		Name:        "events",
		Description: "Events",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "create",
				Description: "Create an event",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "delete",
				Description: "Delete an event",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all events",
			},
		},
	},
	{
		Name:        "dkp-export",
		Type:        discordgo.UserApplicationCommand,
	},
	{
		Name:        "gear",
		Description: "Check your gear and CP",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "User to check",
				Required:    false,
			},
		},
	},
}
