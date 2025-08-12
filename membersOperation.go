package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

func CreateOrUpdateMember(conn *pgx.Conn, m *discordgo.Member, guildId string) error {
	_, err := conn.Exec(context.Background(), `INSERT INTO players (id, name, guild, guild_nick, active) 
VALUES ($1, $2, $3, $4, true)
ON CONFLICT (id, guild) DO UPDATE 
  SET name = EXCLUDED.name, 
      guild_nick = EXCLUDED.guild_nick,
      active = EXCLUDED.active;
`, m.User.ID, m.User.Username, guildId, m.DisplayName())
	return err
}

func SetMemberInactive(m *discordgo.Member, guildId string) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Could not acquire db connection from pool", "error", err)
		return
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), "UPDATE players SET active = false WHERE discord_id = $1 AND guild = $2", m.User.ID, guildId)
	if err != nil {
		log.Error("Could not set player inactive", "error", err)
		return
	}
}
func GetMemberAvatarUrl(m *discordgo.Member, guildId string, size string) string {
	if m.Avatar != "" {
		if member, err := s.GuildMember(guildId, m.User.ID); err != nil {
			log.Error("Could not get member", "error", err)
		} else {
			return member.AvatarURL(size)
		}

	}
	return m.User.AvatarURL(size)
}

// GetMemberAvatarUrlWithDefaultSize returns the avatar URL with a default size.
func GetMemberAvatarUrlWithDefaultSize(m *discordgo.Member, guildId string) string {
	return GetMemberAvatarUrl(m, guildId, "")
}
