package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func commandListener(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	if h, ok := cmdHandler[data.CommandName()]; ok {
		h(event)
	}
}

/*
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

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
				Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})
			conn, err := pool.Acquire(context.Background())
			if err != nil {
				log.Error("Could not aquire db connection from pool", "error", err)
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

			reader := strings.NewReader(csvContent)
			fileNamePart := time.Now().UTC().Format("20060102150405")
			_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
				Content: fmt.Sprintf("DKP Export %s", fileNamePart),
				Files: []*discordgo.File{
					{
						Name:        fmt.Sprintf("%s_dkp_export.csv", fileNamePart),
						ContentType: "text/csv",
						Reader:      reader,
					},
				}})
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
				log.Error("Could not aquire db connection from pool", "error", err)
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
				log.Error("Could not aquire db connection from pool", "error", err)
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
	}
*/
var cmdHandler = map[string]func(event *events.ApplicationCommandInteractionCreate){
	"ping": func(event *events.ApplicationCommandInteractionCreate) {
		err := event.CreateMessage(discord.NewMessageCreateBuilder().SetContent("Pong").SetEphemeral(true).Build())
		if err != nil {
			log.Error("error on sending response", slog.Any("err", err))

		}
	},
	"roll": func(event *events.ApplicationCommandInteractionCreate) {
		number := rollDice()
		responseText := fmt.Sprintf("> <@%s> dice result: %d", event.Member().User.ID, number)
		event.CreateMessage(discord.NewMessageCreateBuilder().SetContent(responseText).Build())
		log.Info("Rolled dice", "user", event.Member().User.ID, "result", number)
	},
	"events": func(event *events.ApplicationCommandInteractionCreate) {
		sub := *event.SlashCommandInteractionData().SubCommandName
		switch sub {
		case "create":
			// Handle the create subcommand
			createEvent(event)
			/*case "delete":
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
				})*/
		}
	},
	"gear": func(event *events.ApplicationCommandInteractionCreate) {
		event.DeferCreateMessage(true)

		var memberID snowflake.ID
		if optionMember, ok := event.SlashCommandInteractionData().Options["user"]; ok {
			memberID = optionMember.Snowflake()
		} else {
			memberID = event.Member().User.ID
		}
		member, err := client.Rest.GetMember(*event.GuildID(), memberID)
		if err != nil {
			log.Error("Could not get guild member", "memberId", memberID, "guildId", *event.GuildID())
			return
		}
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not aquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()

		log.Info("Sent gear embed", "user", member.User.ID, "guild", *event.GuildID())

	},
}

func createEvent(event *events.ApplicationCommandInteractionCreate) {
	minLen5 := 5
	minLen10 := 10
	modal := discord.NewModalCreateBuilder().
		SetCustomID("create_event_modal").
		SetTitle("Create event").
		SetComponents(discord.NewActionRow(
			discord.TextInputComponent{
				CustomID:    "event_name",
				Label:       "Event name",
				Placeholder: "Enter event name",
				MaxLength:   100,
				Required:    true,
				Style:       discord.TextInputStyleShort,
			},
		),
			discord.NewActionRow(discord.TextInputComponent{
				CustomID:    "event_description",
				Label:       "Event description",
				Placeholder: "Enter event description",
				MaxLength:   1000,
				Required:    true,
				Style:       discord.TextInputStyleParagraph,
			}),
			discord.NewActionRow(discord.TextInputComponent{
				CustomID:    "event_date",
				Label:       "Event date",
				Placeholder: "Enter event date (DD-MM-YYY)",
				MinLength:   &minLen10,
				MaxLength:   10,
				Required:    true,
				Style:       discord.TextInputStyleShort,
			}),
			discord.NewActionRow(discord.TextInputComponent{
				CustomID:    "event_time",
				Label:       "Event time",
				Placeholder: "Enter event time (HH:MM)",
				MinLength:   &minLen5,
				MaxLength:   5,
				Required:    true,
				Style:       discord.TextInputStyleShort,
			}),
		).
		Build()
	err := event.Modal(modal)

	if err != nil {
		log.Error("Could not send modal", "error", err)
		return
	}

}
