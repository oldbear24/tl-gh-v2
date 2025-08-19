package main

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var componentsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"set_role": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		roleString := i.MessageComponentData().Values[0]
		roleId := rolesData.GetRoleByName(roleString).Id
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
			return
		}
		defer conn.Release()
		_, err = conn.Exec(context.Background(), `update players set role=$1 where id=$2 and guild=$3`, roleId, i.Member.User.ID, i.GuildID)
		if err != nil {
			log.Error("Could not update user", "error", err)
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})

		_, err = s.InteractionResponseEdit(i.Interaction, memberGearEmbed(conn.Conn(), i.Interaction, i.Interaction.Member))
		if err != nil {
			log.Error("Could not edit interaction message", "error", err)
			return
		}
		log.Debug("Updated player role", "user", i.Member.User.ID, "guild", i.GuildID, "role", roleString)
	},
	"set_cp": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "set_cp_modal",
				Title:    "Set your CP",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: "cp",
								Label:    "Your CP",
								Style:    discordgo.TextInputShort,
								Required: true,
							},
						},
					},
				},
			},
		})
	},
	"set_weapon_1": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		setWeapon(s, i, "1")

	},
	"set_weapon_2": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		setWeapon(s, i, "2")
	},
	"set_build": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "set_build_modal",
				Title:    "Set your build URL",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID: "cp",
								Label:    "Your build URL",
								Style:    discordgo.TextInputShort,
								Required: true,
							},
						},
					},
				},
			},
		})
	},
	"join_event": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		setEventParcitipation(s, i, "going")
	},
	"tentative_event": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		setEventParcitipation(s, i, "tentative")
	},
	"absence_event": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		setEventParcitipation(s, i, "not_going")
	},
}

func setWeapon(s *discordgo.Session, i *discordgo.InteractionCreate, weapon string) {
	weaponString := i.MessageComponentData().Values[0]
	wepId := weaponsData.GetWeaponByName(weaponString).Id
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Could not acquire db connection from pool", "error", err)
		return
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), fmt.Sprintf(`update players set weapon_%s=$1 where id=$2 and guild=$3`, weapon), wepId, i.Member.User.ID, i.GuildID)
	if err != nil {
		log.Error("Could not update user", "error", err)
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})

	_, err = s.InteractionResponseEdit(i.Interaction, memberGearEmbed(conn.Conn(), i.Interaction, i.Interaction.Member))
	if err != nil {
		log.Error("Could not edit interaction message", "error", err)
		return
	}
	log.Debug("Updated player weapon", "user", i.Member.User.ID, "guild", i.GuildID, "slot", weapon, "weapon", weaponString)
}

func setEventParcitipation(s *discordgo.Session, i *discordgo.InteractionCreate, status string) {
	eventMessageId := i.Message.ID
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Could not aquire db connection from pool", "error", err)
		return
	}
	defer conn.Release()
	var eventId int
	var state string
	err = conn.QueryRow(context.Background(), `select id,state from events where guild=$1 and message_id=$2`, i.GuildID, eventMessageId).Scan(&eventId, &state)
	if err != nil {
		log.Error("Could not get event id", "error", err)
		return
	}
	if state != "upcoming" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You can only join upcoming events.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	var roleID int
	err = conn.QueryRow(context.Background(), `select role from players where guild=$1 and id=$2`, i.GuildID, i.Member.User.ID).Scan(&roleID)
	if err != nil {
		log.Error("Could not get user stats", "error", err)
	}
	_, err = conn.Exec(context.Background(), `INSERT INTO event_participants(event, player, guild, status)
VALUES ($1, $2, $3, $4)
ON CONFLICT (event, player,guild)
DO UPDATE SET
    status = EXCLUDED.status
`, eventId, i.Member.User.ID, i.GuildID, status)
	if err != nil {
		log.Error("Could not insert event participant", "error", err)
		return
	}
	log.Debug("Updated event participation", "user", i.Member.User.ID, "guild", i.GuildID, "event", eventId, "status", status)
	updateEventMessage(s, i, conn.Conn(), eventId)
}

func updateEventMessage(s *discordgo.Session, i *discordgo.InteractionCreate, conn *pgx.Conn, eventId int) error {
	embed, err := eventEmbed(conn, eventId)
	if err != nil {
		log.Error("Could not create event embed", "error", err)
	}
	_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
	return err
}
