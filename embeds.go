package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

func eventEmbed(conn *pgx.Conn, id int) (*discordgo.MessageEmbed, error) {
	row := conn.QueryRow(context.Background(), "SELECT name,description,date FROM events where id=$1", id)
	var name string
	var description string
	var date time.Time
	err := row.Scan(&name, &description, &date)
	if err != nil {
		return nil, err
	}

	rows, err := conn.Query(context.Background(), `SELECT ep.status,p."role",p.weapon_1,p.weapon_2,p.guild_nick FROM event_participants ep
join players p on ep.guild = p.guild and ep.player = p.id 
WHERE event=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tanks := ""
	dps := ""
	healers := ""
	tentative := ""
	absent := ""
	for rows.Next() {
		var status string
		var role sql.NullInt32
		var wep1 sql.NullInt32
		var wep2 sql.NullInt32
		var guildNick string
		err := rows.Scan(&status, &role, &wep1, &wep2, &guildNick)
		if err != nil {
			return nil, err
		}
		switch status {
		case "going":
			if role.Valid {
				wep1Em := ""
				wep2Em := ""
				if wep1.Valid {
					wep1Em = getWeaponEmote(int(wep1.Int32))
				}
				if wep2.Valid {
					wep2Em = getWeaponEmote(int(wep2.Int32))
				}
				guildNick = fmt.Sprintf("%s%s%s", wep1Em, wep2Em, guildNick)
				roleData := rolesData.GetRole(int(role.Int32))
				switch roleData.Name {
				case "tank":
					tanks += guildNick + "\n"
				case "dps":
					dps += guildNick + "\n"
				case "healer":
					healers += guildNick + "\n"
				}
			} else {
				return nil, fmt.Errorf("role is null for participant %s in event %d", guildNick, id)
			}
		case "tentative":
			tentative += guildNick
		case "not_going":
			absent += guildNick
		}

	}

	embed := &discordgo.MessageEmbed{
		Title:       name,
		Description: description,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Start",
				Value:  date.Format(time.Layout),
				Inline: false,
			},
			{
				Name:   "Tank:",
				Value:  tanks,
				Inline: true,
			},
			{
				Name:   "DPS",
				Value:  dps,
				Inline: true,
			},
			{
				Name:   "Healer",
				Value:  healers,
				Inline: true,
			},
			{
				Name:   "Tentative",
				Value:  tentative,
				Inline: false,
			},
			{
				Name:   "Absence",
				Value:  absent,
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("ID: %d", id),
		},
	}
	return embed, nil

}

func memberGearEmbed(conn *pgx.Conn, i *discordgo.Interaction, member *discordgo.Member) *discordgo.WebhookEdit {
	var cp int
	var role sql.NullInt32
	var weapon1 sql.NullInt32
	var weapon2 sql.NullInt32
	var buildUrl sql.NullString
	guildId, _ := strconv.ParseInt(i.GuildID, 10, 64)
	userId, _ := strconv.ParseInt(member.User.ID, 10, 64)

	weapons := []discordgo.SelectMenuOption{}
	roles := []discordgo.SelectMenuOption{}
	weaponCacheData := weaponsData.GetAllWeapons()
	rolesCacheData := rolesData.GetAllRoles()
	for _, v := range *weaponCacheData {
		weapons = append(weapons, discordgo.SelectMenuOption{Value: v.Name, Label: v.VisibleName, Emoji: &discordgo.ComponentEmoji{ID: v.Emote}})

	}
	for _, v := range *rolesCacheData {
		roles = append(roles, discordgo.SelectMenuOption{Value: v.Name, Label: v.VisibleName, Emoji: &discordgo.ComponentEmoji{ID: v.Emote}})
	}

	err := conn.QueryRow(context.Background(), `select combat_power,role,weapon_1,weapon_2, build_url from players where guild=$1 and id=$2`, guildId, userId).Scan(&cp, &role, &weapon1, &weapon2, &buildUrl)

	if err != nil {
		log.Error("Could not get user stats", "error", err)
	}
	weapon1Field := ""
	weapon2Field := ""
	buildUrlField := ""
	roleField := ""
	if role.Valid {
		if roleData := rolesData.GetRole(int(role.Int32)); roleData.Name != "" {
			roleField = fmt.Sprintf("<:%s:%s> %s", roleData.Name, roleData.Emote, roleData.VisibleName)
		}
	}
	if buildUrl.Valid {
		buildUrlField = fmt.Sprintf("[Here](%s)", buildUrl.String)
	} else {
		buildUrlField = "Not set"
	}

	if weapon1.Valid {

		if weapon1Data := weaponsData.GetWeapon(int(weapon1.Int32)); weapon1Data.Name != "" {
			weapon1Field = fmt.Sprintf("<:%s:%s> %s", weapon1Data.Name, weapon1Data.Emote, weapon1Data.VisibleName)
		}

	}
	if weapon2.Valid {
		if weapon2Data := weaponsData.GetWeapon(int(weapon2.Int32)); weapon2Data.Name != "" {
			weapon2Field = fmt.Sprintf("<:%s:%s> %s", weapon2Data.Name, weapon2Data.Emote, weapon2Data.VisibleName)
		}

	}
	components := []discordgo.MessageComponent{}
	if member.User.ID == i.Member.User.ID {
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "set_weapon_1",
						MenuType:    discordgo.StringSelectMenu,
						Placeholder: "Set weapon 1",
						Options:     weapons,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "set_weapon_2",
						MenuType:    discordgo.StringSelectMenu,
						Placeholder: "Set weapon 2",
						Options:     weapons,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "set_role",
						MenuType:    discordgo.StringSelectMenu,
						Placeholder: "Role",
						Options:     roles,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Set CP",
						CustomID: "set_cp",
					},
					discordgo.Button{
						Label:    "Set build",
						CustomID: "set_build",
					},
				},
			},
		}

	}

	return &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:     getMemberGuildNick(member),
				Thumbnail: &discordgo.MessageEmbedThumbnail{URL: GetMemberAvatarUrl(member, i.GuildID, "256")},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "CP",
						Value:  strconv.Itoa(cp),
						Inline: true,
					},
					{
						Name:   "Role",
						Value:  roleField,
						Inline: true,
					},
					{
						Name:   "Build",
						Value:  buildUrlField,
						Inline: false,
					},
					{
						Name:   "Weapon 1",
						Value:  weapon1Field,
						Inline: true,
					},
					{
						Name:   "Weapon 2",
						Value:  weapon2Field,
						Inline: true,
					},
				},
			},
		},
		Components: &components,
	}
}

func getWeaponEmote(weapon int) string {
	wepData := weaponsData.GetWeapon(int(weapon))
	if wepData == nil || wepData.Emote == "" {
		return ""
	}
	return fmt.Sprintf("<:%s:%s>", wepData.Name, wepData.Emote)
}
