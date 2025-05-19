package main

import (
	"context"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5"
)

func startWorkingThread(c *bot.Client) {
	go workingThread(c)
}

func workingThread(c *bot.Client) {
	workingThreadTimers := make(map[string]time.Time)
	for range time.Tick(time.Second) {
		if checkIfTaskShouldBeRun("syncMembers", workingThreadTimers, time.Hour) {
			go syncMembers(c)
		}
		if checkIfTaskShouldBeRun("proccesNotifications", workingThreadTimers, time.Second*30) {
			go proccesNotifications(c)
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

func syncMembers(c *bot.Client) {
	s := *c
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
		var id int64
		err := guildRows.Scan(&id)
		if err != nil {
			log.Error("Could not get guildId from db row", "error", err)
			continue
		}
		var after snowflake.ID

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

			members, err := s.Rest.GetMembers(snowflake.ID(id), 100, after)
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

func proccesNotifications(c *bot.Client) {
	s := *c
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
		var playerId int64
		var notificationText string
		err := rows.Scan(&id, &playerId, &notificationText)
		if err != nil {
			log.Error("Could not get notification from db row", "error", err)
			continue
		}
		idsToDelete = append(idsToDelete, id)

		channel, err := s.Rest.CreateDMChannel(snowflake.ID(playerId))
		if err != nil {
			log.Error("Could not create user channel", "error", err)
			continue
		}
		_, err = s.Rest.CreateMessage(channel.ID(), discord.NewMessageCreateBuilder().SetContent(notificationText).Build())
		if err != nil {
			log.Error("Could not send notification", "error", err)
			continue
		}
		log.Debug("Sent notification", "userId", playerId, "notificationText", notificationText)
	}
	_, err = tx.Exec(context.Background(), "delete from notification_queue where id = ANY($1)", idsToDelete)
	if err != nil {
		log.Error("Could not delete notification from db", "error", err)
		return
	}
	tx.Commit(context.Background())
}
