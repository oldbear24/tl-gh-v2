package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

func UpdateGuildRecord(g *discordgo.Guild) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Unable to acquire connection", "error", err)
		return
	}
	defer conn.Release()
	var exists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS( SELECT 1 FROM guilds WHERE id = $1 LIMIT 1)", g.ID).Scan(&exists)
	if err != nil {
		log.Error("Error checking guild exists", "error", err)
		return
	}
	if exists {
		_, err = conn.Exec(context.Background(), "UPDATE guilds SET name = $1 WHERE id = $2", g.Name, g.ID)
		if err != nil {
			log.Error("Error updating guild record", "error", err)
			return
		}
	} else {
		_, err = conn.Exec(context.Background(), "INSERT INTO guilds (id, name) VALUES ($1, $2)", g.ID, g.Name)
		if err != nil {
			log.Error("Error inserting guild record", "error", err)
			return
		}
	}
}
