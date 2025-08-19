package main

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
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
		/*	if checkIfTaskShouldBeRun("proccesNotifications", workingThreadTimers, time.Second*30) {
			go proccesNotifications(s)
		}*/
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
		log.Debug("Syncing members for guild", "guild", id)
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
		log.Debug("Finished syncing members", "guild", id)
	}
}

func proccesNotifications(s *discordgo.Session) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("Failed to obtain connection from pool", "error", err)
		return
	}
	defer conn.Release()
	tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		log.Error("Could not open transaction", "error", err)
		return
	}
	defer tx.Rollback(context.Background())
	rows, err := tx.Query(context.Background(), "select id,player_id,notification_text from notification_queue FOR UPDATE")
	if err != nil {
		log.Error("Could not retrieve notifications from database", "error", err)
		return
	}
	idsToDelete := []int{}

	defer rows.Close()
	for rows.Next() {
		var id int
		var playerId string
		var notificationText string
		err := rows.Scan(&id, &playerId, &notificationText)
		if err != nil {
			log.Error("Could not get notification from db row", "error", err)
			continue
		}
		idsToDelete = append(idsToDelete, id)

		channel, err := s.UserChannelCreate(playerId)
		if err != nil {
			log.Error("Could not create user channel", "error", err)
			continue
		}
		_, err = s.ChannelMessageSend(channel.ID, notificationText)
		if err != nil {
			log.Error("Could not send notification", "error", err)
			continue
		}
		log.Debug("Sent notification", "userId", playerId, "notificationText", notificationText)
	}
	if len(idsToDelete) > 0 {
		_, err = tx.Exec(context.Background(), "delete from notification_queue where id = ANY($1)", idsToDelete)
		if err != nil {
			log.Error("Could not delete notification from db", "error", err)
			return
		}
	}
	tx.Commit(context.Background())
}
