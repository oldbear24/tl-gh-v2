package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5/pgxpool"

	logoutput "github.com/oldbear24/openobserve-go-slog-writer"
	pgmigrations "github.com/oldbear24/tl-gh-v2/internal/pgMigrations"
)

var log *slog.Logger
var botToken string
var postgreConnString string
var s *discordgo.Session
var pool *pgxpool.Pool
var logOutput *logoutput.LogOutput

func main() {
	initApp()
	var err error
	defer logOutput.Close()
	defer panicRecover()
	config, err := pgxpool.ParseConfig(postgreConnString)
	if err != nil {
		log.Error("Failed to parse config", "error", err)
		return
	}
	config.ConnConfig.Tracer = &AppQueryTracer{}
	//	config.ConnConfig.Tracer. = &pgx.QueryTracer{}

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Error("Unable to connect to database", "error", err)
		panic(err)
	}
	exePath, _ := os.Executable()
	exrDir := filepath.Dir(exePath)
	pgmigrations.Init(filepath.Join(exrDir, "migrations"), log)
	mConn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("")
		panic(err)
	}
	pgmigrations.RunMigrations(mConn.Conn())
	mConn.Release()
	s, err = discordgo.New("Bot " + botToken)
	log.Info("Bot is starting...") // Log when the bot starts
	if err != nil {
		log.Error("error creating Discord session,", "error", err)
		panic(err)
	}
	s.Identify.Intents = discordgo.IntentsAll
	registerHooks(s)
	// Open a websocket connection to Discord and begin listening.
	err = s.Open()
	if err != nil {
		log.Error("error opening connection,", "error", err)
		panic(err)
	}
	defer s.Close()
	log.Info("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Error("Cannot create command,", "error", err, "command", v.Name)
			panic(err)
		}
		log.Info("Registered command", "command", cmd.Name)
		registeredCommands[i] = cmd
	}

	// Delete unused commands
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		log.Error("Cannot fetch existing commands,", "error", err)
		panic(err)
	}
	for _, cmd := range existingCommands {
		found := false
		for _, regCmd := range registeredCommands {
			if cmd.ID == regCmd.ID {
				found = true
				break
			}
		}
		if !found {
			err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Error("Cannot delete command,", "error", err, "command", cmd.Name)
				panic(err)
			}
			log.Info("Deleted unused command", "command", cmd.Name)
		}
	}
	startWorkingThread(s)
	// Wait here until CTRL-C or other term signal is received.
	log.Info("Bot is now running.  Press CTRL-C to exit.") // Log when the bot is running
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Info("Shutting down...") // Log when the bot is shutting down
}

// panicRecover logs and re-panics if a panic occurs during execution.
func panicRecover() {
	if r := recover(); r != nil {
		log.Error("Fate has decided to end this program", "error", r)
		panic(r)
	}
}

// initApp reads configuration from environment variables and command-line
// flags, configuring logging and application-wide settings.
func initApp() {
	var enableExtenal bool
	var externalLogUrl string
	var externalLogToken string
	var externalLogOrg string
	var externalLogStream string
	var logLevelInt int
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key, value := pair[0], pair[1]
		if strings.HasPrefix(key, "TLGH_") {
			switch key {
			case "TLGH_BOT_TOKEN":
				botToken = value
			case "TLGH_POSTGRE_CONN_STRING":
				postgreConnString = value
			case "TLGH_EXTLOG_ENABLE":
				enableExtenal, _ = strconv.ParseBool(value)
			case "TLGH_LOG_URL":
				externalLogUrl = value
			case "TLGH_LOG_TOKEN":
				externalLogToken = value
			case "TLGH_LOG_ORG":
				externalLogOrg = value
			case "TLGH_LOG_STREAM":
				externalLogStream = value
			case "TLGH_LOG_LEVEL":
				logLevelInt, _ = strconv.Atoi(value)
			}
		}
	}
	flag.StringVar(&botToken, "t", "", "Bot Token")
	flag.StringVar(&postgreConnString, "db", "", "PostgreSQL Connection String")
	flag.BoolVar(&enableExtenal, "extlogenable", false, "Enable External Log")
	flag.StringVar(&externalLogUrl, "logurl", "", "External Log URL")

	flag.StringVar(&externalLogToken, "logtoken", "", "External Log Token")
	flag.StringVar(&externalLogOrg, "logorg", "", "External Log Organization")
	flag.StringVar(&externalLogStream, "logstream", "", "External Log Stream")
	flag.IntVar(&logLevelInt, "loglevel", 0, "Log Level")
	flag.Parse()

	var logLevel slog.Level
	switch logLevelInt {
	case -4:
		logLevel = slog.LevelDebug
	case 0:
		logLevel = slog.LevelInfo
	case 4:
		logLevel = slog.LevelWarn
	case 8:
		logLevel = slog.LevelError
	}
	logOutput = logoutput.New(enableExtenal, externalLogUrl, externalLogToken, externalLogOrg, externalLogStream)
	logHandler := slog.NewJSONHandler(logOutput, &slog.HandlerOptions{Level: logLevel, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			a.Key = "_timestamp"
			a.Value = slog.Int64Value(time.Now().UnixMicro())
		}
		return a
	}})
	log = slog.New(logHandler)
}
