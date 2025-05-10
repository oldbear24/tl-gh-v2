package main

import (
	"context"
	"fmt"
	"strings"
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
			s.InteractionRespond(i.Interaction, eventEmbed())

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
}
