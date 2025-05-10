package main

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var componentsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"set_role": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		roleString := i.MessageComponentData().Values[0]
		roleId := rolesData.GetRoleByName(roleString).Id
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not aquire db connection from pool", "error", err)
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
}

func setWeapon(s *discordgo.Session, i *discordgo.InteractionCreate, weapon string) {
	weaponString := i.MessageComponentData().Values[0]
	wepId := weaponsData.GetWeaponByName(weaponString).Id
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Could not aquire db connection from pool", "error", err)
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
}
