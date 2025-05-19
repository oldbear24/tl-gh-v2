package main

import "github.com/disgoorg/disgo/discord"

// var serverAdminPermission = int64(discordgo.PermissionAdministrator | discordgo.PermissionManageServer)
var commands = []discord.ApplicationCommandCreate{

	discord.SlashCommandCreate{
		Name:        "ping",
		Description: "Sends a pong message",
	},
	discord.SlashCommandCreate{
		Name:        "roll",
		Description: "Rolls a dice",
	},
	discord.SlashCommandCreate{
		Name:        "events",
		Description: "Events",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "create",
				Description: "Create an event",
			},
		},
	},
	discord.UserCommandCreate{
		Name: "dkp-export",
	},
	discord.SlashCommandCreate{
		Name:        "gear",
		Description: "Check your gear and CP",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				Name:        "user",
				Description: "User to check",
				Required:    false,
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "notifications",
		Description: "Enable or disable notifications",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "type",
				Description: "Type of notification",
				Required:    true,
				Choices: []discord.ApplicationCommandOptionChoiceString{
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
		},
	},
	discord.SlashCommandCreate{

		Name:        "generate-auth",
		Description: "Generate an auth credentials",
		//TODO: add perms
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "guild-id",
				Description: "Guild ID",
				Required:    true,
			},
		},
	},
}
