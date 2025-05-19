package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/jackc/pgx/v5/pgxpool"
	logoutput "github.com/oldbear24/tl-gh-v2/internal/logOutput"
	pgmigrations "github.com/oldbear24/tl-gh-v2/internal/pgMigrations"
)

var log *slog.Logger
var botToken string
var postgreConnString string
var pool *pgxpool.Pool
var logOutput *logoutput.LogOutput
var client *bot.Client

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
	pgmigrations.Init("./migrations/", log)
	mConn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Error("")
		panic(err)
	}
	pgmigrations.RunMigrations(mConn.Conn())
	mConn.Release()

	client, err = disgo.New(botToken,
		// set gateway options
		bot.WithGatewayConfigOpts(
			// set enabled intents
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
			),
		),
		bot.WithEventListenerFunc(commandListener),
		bot.WithEventListeners(registerEventListener()),
		// add event listeners
	)
	log.Info("Bot is starting...") // Log when the bot starts
	if err != nil {
		log.Error("error creating Discord session,", "error", err)
		panic(err)
	}
	defer client.Close(context.TODO())
	registerHooks(client)
	// Open a websocket connection to Discord and begin listening.
	if _, err = client.Rest.SetGlobalCommands(client.ApplicationID, commands); err != nil {
		slog.Error("error while registering commands", slog.Any("err", err))
	}
	if err = client.OpenGateway(context.TODO()); err != nil {
		slog.Error("errors while connecting to gateway", slog.Any("err", err))
		return
	}
	startWorkingThread(client)
	// Wait here until CTRL-C or other term signal is received.
	log.Info("Bot is now running.  Press CTRL-C to exit.") // Log when the bot is running
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Info("Shutting down...") // Log when the bot is shutting down
}

func panicRecover() {
	if r := recover(); r != nil {
		log.Error("Fate has decided to end this program", "error", r)
		panic(r)
	}
}

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
