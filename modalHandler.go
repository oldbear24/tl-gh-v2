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
	"create_event_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		data := i.ModalSubmitData()
		name := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		date := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		time := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Cannot acquire DB connection", "error", err)
			return
		}
		startDate, err := parseDateTime(date, time)
		if err != nil {
			log.Error("Could not parse date and time", "date", date, "time_", time)
			return
		}
		tx, err := conn.Begin(context.Background())
		if err != nil {
			log.Error("Could not open transcation", "error", err)
			return
		}
		defer tx.Rollback(context.Background())
		row := conn.QueryRow(context.Background(), "INSERT INTO events(guild,channel,name,description,date) VALUES($1,$2,$3,$4,$5) RETURNING id;", i.GuildID, i.ChannelID, name, description, startDate)
		var id int
		err = row.Scan(&id)
		if err != nil {
			log.Error("Could not get id from event record", "error", err)
			return
		}
		embed, err := eventEmbed(conn.Conn(), id)
		if err != nil {
			log.Error("Could not create event embed", "error", err)
		}

		mess, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Join",
							Style:    discordgo.PrimaryButton,
							CustomID: "join_event",
						},
						discordgo.Button{
							Label:    "Tentative",
							Style:    discordgo.SecondaryButton,
							CustomID: "tentative_event",
						},
						discordgo.Button{
							Label:    "Absence",
							Style:    discordgo.DangerButton,
							CustomID: "absence_event",
						},
					},
				},
			},
		})
		if err != nil {
			log.Error("Could not send message", "error", err)
			return
		}
		_, err = tx.Exec(context.Background(), "update events set message_id = $1 where id=$2", mess.ID, id)
		if err != nil {
			log.Error("Could update event", "error", err)
			s.ChannelMessageDelete(i.ChannelID, mess.ID)
			return
		}
		s.InteractionResponseDelete(i.Interaction)
		tx.Commit(context.Background())
	},
}
