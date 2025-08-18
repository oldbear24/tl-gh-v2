package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

func registerHooks(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentsHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			if h, ok := modalHandlers[i.ModalSubmitData().CustomID]; ok {
				h(s, i)
			}
		default:
			break
		}

	})
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Info("logged in as", "username", s.State.User.Username, "discriminator", s.State.User.Discriminator)
	})
	s.AddHandler(func(s *discordgo.Session, i *discordgo.GuildCreate) {
		log.Info("joined guild", "guild", i.Guild.Name)
		UpdateGuildRecord(i.Guild)
	})
	s.AddHandler(func(s *discordgo.Session, i *discordgo.GuildMemberAdd) {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
		}
		defer conn.Release()
		CreateOrUpdateMember(conn.Conn(), i.Member, i.GuildID)

	})
	s.AddHandler(func(s *discordgo.Session, i *discordgo.GuildMemberRemove) {
		SetMemberInactive(i.Member, i.GuildID)
	})
	s.AddHandler(func(s *discordgo.Session, i *discordgo.GuildMemberUpdate) {
		conn, err := pool.Acquire(context.Background())
		if err != nil {
			log.Error("Could not acquire db connection from pool", "error", err)
		}
		defer conn.Release()
		CreateOrUpdateMember(conn.Conn(), i.Member, i.GuildID)
	})
}
