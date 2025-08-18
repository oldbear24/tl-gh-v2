package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Pong!",
			},
		})
		log.Info("Ping command executed", "user", i.Member.User.ID)
	},
	"roll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		number := rollDice()
		responseText := fmt.Sprintf("> <@%s> dice result: %d", i.Member.User.ID, number)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{
			Content: responseText,
		}})
		log.Info("Rolled dice", "user", i.Member.User.ID, "result", number)
	},
	"events": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Handle the events command
		if len(i.ApplicationCommandData().Options) == 0 {
			return
		}
		switch i.ApplicationCommandData().Options[0].Name {
		case "create":
			// Handle the create subcommand
			createEvent(s, i)
		case "delete":
			// Handle the delete subcommand
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "Delete event",
				},
			})
		case "list":
			// Handle the list subcommand
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: "List events",
				},
			})
		}

	},
	"gear": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		optMap := GetOptions(i.ApplicationCommandData().Options)
		member := i.Interaction.Member
		if optMember, ok := optMap["user"]; ok {
			if gMember, err := s.GuildMember(i.GuildID, optMember.UserValue(s).ID); err == nil {
				member = gMember
			}

		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2}})
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()

		_, err = s.InteractionResponseEdit(i.Interaction, memberGearEmbed(conn.Conn(), i.Interaction, member))
		if err != nil {
			log.Error("Could not edit interaction message", "error", err)
			return
		}
		log.Info("Sent gear embed", "user", member.User.ID, "guild", i.GuildID)
	},
	"dkp-export": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Export will be sent to your DMs",
				Flags:   discordgo.MessageFlagsSuppressNotifications | discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Error("Could not send message", "error", err)
			return
		}
		guild, err := s.State.Guild(i.GuildID)
		if err != nil {
			log.Error("Could not get guild from state", "error", err)
			return
		}
		csvContent := ""
		for _, voiceState := range guild.VoiceStates {
			if voiceState.ChannelID == i.ChannelID {
				csvContent += fmt.Sprintf("%s\n", voiceState.UserID)
			}
		}
		channel, err := s.UserChannelCreate(i.Member.User.ID)
		if err != nil {
			log.Error("Could not create user channel", "error", err)
			return
		}

		fileNamePart := time.Now().UTC().Format("20060102150405")
		filename := fmt.Sprintf("%s_dkp_export.csv", fileNamePart)
		spacing := discordgo.SeparatorSpacingSizeLarge
		r := bytes.NewReader([]byte(csvContent))
		color := 48895
		s.InteractionResponseDelete(i.Interaction)
		_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
			Flags: discordgo.MessageFlagsIsComponentsV2,
			Components: []discordgo.MessageComponent{
				discordgo.Container{
					AccentColor: &color,
					Components: []discordgo.MessageComponent{
						discordgo.TextDisplay{
							Content: fmt.Sprintf("# DKP Export %s", fileNamePart),
						},
						discordgo.Separator{
							Spacing: &spacing,
						},
						discordgo.FileComponent{
							File: discordgo.UnfurledMediaItem{
								URL: fmt.Sprintf("attachment://%s", filename),
							},
						},
					},
				},
			},
			Files: []*discordgo.File{
				{
					Name:   filename,
					Reader: r,
				},
			},
		})
		if err != nil {
			log.Error("Could not send message", "error", err)
			return
		}
		log.Info("Sent DKP export", "user", i.Member.User.ID, "channel", i.ChannelID, "guild", i.GuildID)
	},
	"generate-auth": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		options := GetOptions(i.ApplicationCommandData().Options)
		guildId := options["guild-id"].StringValue()
		if guildId != i.GuildID {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid guild ID",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			if err != nil {
				log.Error("Could not send message", "error", err)
			}
			return
		}
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Error("Could not send message", "error", err)
			return
		}
		// Generate auth credentials
		user := randomString(20)

		password, passwordHash, err := generatePasswordAndHash()
		if err != nil {
			log.Error("Could not generate password and hash", "error", err)
			return
		}
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()
		tx, err := conn.Begin(context.Background())
		if err != nil {
			log.Error("Could not open transaction", "error", err)
			return
		}
		defer tx.Rollback(context.Background())

		_, err = tx.Exec(context.Background(), "update guilds set api_user=$1, api_key=$2 where id =$3 ", user, passwordHash, guildId)
		if err != nil {
			log.Error("Could not update guilds", "error", err)
			return
		}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				{
					Title: "Auth credentials",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "User",
							Value: user,
						},
						{
							Name:  "Password",
							Value: password,
						},
					},
				},
			},
		})

		if err != nil {
			log.Error("Could not edit interaction message", "error", err)
			return
		}

		log.Info("Generated auth credentials", "user", i.Member.User.ID, "guild", guildId)
		tx.Commit(context.Background())
	},
	"notifications": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		opt := GetOptions(i.ApplicationCommandData().Options)
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()
		columnName := ""
		typeConfig := opt["type"].StringValue()
		switch typeConfig {
		case "auction":
			columnName = "auctions"
		case "event":
			columnName = "events"
		}
		enable := opt["enable"].BoolValue()
		_, err = conn.Exec(context.Background(), fmt.Sprintf(`update player_configs set %s=$1 where player=$2 and guild=$3`, columnName), enable, i.Member.User.ID, i.GuildID)
		if err != nil {
			log.Error("Could not update player config", "error", err)
			return
		}
		log.Debug("Updated player config", "user", i.Member.User.ID, "guild", i.GuildID, "type", typeConfig, "notifications", enable)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: fmt.Sprintf("Notifications for %s set to %t", typeConfig, enable),
			},
		})
	},
	"set_game_roles": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		opt := GetOptions(i.ApplicationCommandData().Options)
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()

		_, err = conn.Exec(context.Background(), `update guilds set game_role=$1, game_leader_role=$2 where id=$3`, opt["role"].RoleValue(s, i.GuildID).ID, opt["leader_role"].RoleValue(s, i.GuildID).ID, i.GuildID)
		if err != nil {
			log.Error("Could not update guilds", "error", err)
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Game roles updated"}})

	},
	"add-game-role": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()
		var gameRole string
		var leaderRole string
		err = conn.QueryRow(context.Background(), `select game_role, game_leader_role from guilds where id=$1`, i.GuildID).Scan(&gameRole, &leaderRole)
		if err != nil {
			log.Error("Could not get game roles", "error", err)
			return
		}
		if gameRole == "" {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("Game role is not set for this server."),
			})
			return
		}
		if leaderRole == "" {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("Game leader role is not set for this server."),
			})
			return
		}

		if i.ApplicationCommandData().TargetID == "" {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("You need to select a user to add the game role."),
			})
			return
		}
		if i.Member.Roles == nil || !contains(i.Member.Roles, leaderRole) {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("You need to be a game leader to add game roles."),
			})
			return
		}
		// Add game role to user
		err = s.GuildMemberRoleAdd(i.GuildID, i.ApplicationCommandData().
			TargetID, gameRole)
		if err != nil {
			log.Error("Could not add game role", "error", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("Could not add game role. Please try again later."),
			})
			return
		}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr("Game role added successfully."),
		})

		log.Info("Added game role", "user", i.ApplicationCommandData().TargetID, "guild", i.GuildID, "role", gameRole)
	},
	"remove-game-role": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})

		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()
		var gameRole string
		var leaderRole string
		err = conn.QueryRow(context.Background(), `select game_role, game_leader_role from guilds where id=$1`, i.GuildID).Scan(&gameRole, &leaderRole)
		if err != nil {
			log.Error("Could not get game roles", "error", err)
			return
		}
		if gameRole == "" {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("Game role is not set for this server."),
			})
			return
		}
		if leaderRole == "" {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("Game leader role is not set for this server."),
			})
			return
		}
		if i.ApplicationCommandData().TargetID == "" {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("You need to select a user to remove the game role."),
			})
			return
		}
		if i.Member.Roles == nil || !contains(i.Member.Roles, leaderRole) {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("You need to be a game leader to remove game roles."),
			})
			return
		}
		// Remove game role from user
		err = s.GuildMemberRoleRemove(i.GuildID, i.ApplicationCommandData().
			TargetID, gameRole)
		if err != nil {
			log.Error("Could not remove game role", "error", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: stringPtr("Could not remove game role. Please try again later."),
			})
			return
		}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr("Game role removed successfully."),
		})
		log.Info("Removed game role", "user", i.Member.User.ID, "guild", i.GuildID, "role", gameRole)
	},
}



func createEvent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Create event",
			CustomID: "create_event_modal",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{

						discordgo.TextInput{
							CustomID:    "event_name",
							Label:       "Event name",
							Style:       discordgo.TextInputShort,
							Placeholder: "Enter event name",
							MinLength:   1,
							MaxLength:   100,
							Required:    true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "event_description",
							Label:       "Event description",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Enter event description",
							MinLength:   1,
							MaxLength:   1000,
							Required:    true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "event_date",
							Label:       "Event date",
							Style:       discordgo.TextInputShort,
							Placeholder: "Enter event date (DD-MM-YYYY)",
							MinLength:   10,
							MaxLength:   10,
							Required:    true,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "event_time",
							Label:       "Event time",
							Style:       discordgo.TextInputShort,
							Placeholder: "Enter event time (HH:MM)",
							MinLength:   5,
							MaxLength:   5,
							Required:    true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Error("Could not send modal", "error", err)
		return
	}

}
