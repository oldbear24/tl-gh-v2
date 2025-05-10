package main

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

var modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"set_cp_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ModalSubmitData()
		var cp = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		cpInt, err := strconv.Atoi(cp)
		if err != nil {
			log.Error("Could not parse CP", "error", err)
			return
		}
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not get connection from pool", "error", err)
			return
		}
		defer conn.Release()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})
		_, err = conn.Exec(context.Background(), "update players set combat_power=$1 where id = $2 and guild = $3", cpInt, i.Member.User.ID, i.GuildID)
		if err != nil {
			log.Error("Could not update players CP", "error", err)
			return
		}
		s.InteractionResponseEdit(i.Interaction, memberGearEmbed(conn.Conn(), i.Interaction, i.Interaction.Member))
	},
	"set_build_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ModalSubmitData()
		var buildUrl = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not get connection from pool", "error", err)
			return
		}
		defer conn.Release()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})
		_, err = conn.Exec(context.Background(), "update players set build_url=$1 where id = $2 and guild = $3", buildUrl, i.Member.User.ID, i.GuildID)
		if err != nil {
			log.Error("Could not update players Build URL", "error", err)
			return
		}
		s.InteractionResponseEdit(i.Interaction, memberGearEmbed(conn.Conn(), i.Interaction, i.Interaction.Member))
	},
	"setup_event_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		log.Error("No handler for modal", "id", i.ID)
	},
}
