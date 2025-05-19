package main

import (
	"context"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func registerHooks(c *bot.Client) {
	/*c.AddHandler()
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
	*/
}

func registerEventListener() *events.ListenerAdapter {
	return &events.ListenerAdapter{
		OnReady:             onReadyListener,
		OnGuildJoin:         onGuildJoin,
		OnGuildMemberJoin:   onGuildMemberJoin,
		OnGuildMemberLeave:  onGuildMemberLeave,
		OnGuildMemberUpdate: onGuildMemberUpdate,
	}

}

func onReadyListener(event *events.Ready) {
	log.Info("logged in as", "username", event.User.Username, "discriminator", event.User.Discriminator)

}
func onGuildJoin(event *events.GuildJoin) {
	log.Info("joined guild", "guild", event.Guild.Name)
	UpdateGuildRecord(event.Guild)
}
func onGuildMemberJoin(event *events.GuildMemberJoin) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Could not aquire db connection from pool", "error", err)
	}
	defer conn.Release()
	CreateOrUpdateMember(conn.Conn(), event.Member, int64(event.GuildID))

}
func onGuildMemberLeave(event *events.GuildMemberLeave) {
	SetMemberInactive(event.Member, int64(event.GuildID))
}

func onGuildMemberUpdate(event *events.GuildMemberUpdate) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Could not aquire db connection from pool", "error", err)
	}
	defer conn.Release()
	CreateOrUpdateMember(conn.Conn(), event.Member, int64(event.GuildID))
}
