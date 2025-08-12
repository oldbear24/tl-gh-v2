package main

import (
	"github.com/bwmarrin/discordgo"
)

var serverAdminPermission = int64(discordgo.PermissionAdministrator | discordgo.PermissionManageGuild)
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
			/*{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "delete",
				Description: "Delete an event",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all events",
			},*/
		},
	},
	{
		Name: "dkp-export",
		Type: discordgo.UserApplicationCommand,
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
	{
		Name:        "set-gear",
		Description: "Set your gear and CP",
	},
	{
		Name:        "notifications",
		Description: "Enable or disable notifications",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Type of notification",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "auction",
						Value: "auction",
					},
					{
						Name:  "event",
						Value: "event",
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "enable",
				Description: "Enable or disable notifications",
				Required:    true,
			},
		},
	},
	{
		Name:                     "generate-auth",
		Description:              "Generate an auth credentials",
		DefaultMemberPermissions: &serverAdminPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "guild-id",
				Description: "Guild ID",
				Required:    true,
			},
		},
	},
}
