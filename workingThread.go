package main

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

func startWorkingThread(s *discordgo.Session) {
	go workingThread(s)
}

func workingThread(s *discordgo.Session) {
	workingThreadTimers := make(map[string]time.Time)
	for range time.Tick(time.Second) {
		if checkIfTaskShouldBeRun("syncMembers", workingThreadTimers, time.Hour) {
			go syncMembers(s)
		}
	}

}
func checkIfTaskShouldBeRun(command string, timers map[string]time.Time, runEvery time.Duration) bool {
	currentTime := time.Now()
	runTime, ok := timers[command]
	if !ok || runTime.Before(currentTime) {
		timers[command] = currentTime.Add(runEvery)
		log.Info("Running working thread command", "command", command)
		return true
	}
	return false
}
func syncMembers(s *discordgo.Session) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("", "error", err)
		return
	}
	defer conn.Release()
	guildRows, err := conn.Query(context.Background(), "select id from guilds")
	if err != nil {
		log.Error("Could not retrieve guilds from database", "error", err)
		return
	}
	for guildRows.Next() {
		var id string
		err := guildRows.Scan(&id)
		if err != nil {
			log.Error("Could not get guildId from db row", "error", err)
			continue
		}
		var after string

		tx, err := pool.Begin(context.Background())
		if err != nil {
			log.Error("Could not open transaction", "error", err)
			return
		}
		_, err = tx.Exec(context.Background(), "update players set active = false where guild=$1", id)
		if err != nil {
			log.Error("Could update player active status", "error", err)
			return
		}
		defer tx.Rollback(context.Background())
		for {
			members, err := s.GuildMembers(id, after, 100)
			if err != nil {
				log.Error("Could not retrieve guild members", "error", err)
				break
			}
			if len(members) == 0 {
				break
			}
			after = members[len(members)-1].User.ID
			for _, member := range members {
				if member.User.Bot {
					continue
				}
				if err := CreateOrUpdateMember(tx.Conn(), member, id); err != nil {
					log.Error("Could not update user", "error", err)
				}
			}
		}
		tx.Commit(context.Background())
	}
}
