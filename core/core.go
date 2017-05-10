package core

import (
	"github.com/pkg/errors"
	"notifier/bot"
	"notifier/config"
	"notifier/gateway"
	"notifier/incoming"
	"notifier/logging"
	"notifier/neo"
	"notifier/sender"
	"notifier/storage"
	"os"
	"os/signal"
	"syscall"
)

var (
	gLogger = logging.WithPackage("core")
)

func Initialization(confPath string) {
	config.Initialization(confPath)
	conf := config.GetInstance()
	logging.PatchStdLog(conf.LogLevel, conf.ServiceName, conf.ServerID)
	gLogger.Info("Environment has been initialized")
}

func Run(confPath string) {
	Initialization(confPath)
	conf := config.GetInstance()
	incomingQueue := incomming.NewMemoryQueue()
	gLogger.Info("Initializing neo client")
	neoDB, err := neo.NewClient(conf.Neo.Host, conf.Neo.Port, conf.Neo.User, conf.Neo.Password, conf.Neo.Timeout,
		conf.Neo.PoolSize)
	if err != nil {
		panic(errors.Wrap(err, "cannot create neo client"))
	}
	gLogger.Info("Initializing messages sender")
	msgsSender, err := sender.NewTelegramSender(conf.Telegram.APIToken, conf.TelegramSender.HttpTimeout)
	if err != nil {
		panic(errors.Wrap(err, "cannot create neo client"))
	}
	dataStorage := storage.NewNeoStorage(neoDB)
	botService := bot.New(incomingQueue, neoDB, msgsSender, dataStorage)
	pollerService := gateway.NewTelegramPoller(incomingQueue)
	gLogger.Info("Starting bot service")
	botService.Start()
	defer botService.Stop()
	gLogger.Info("Starting telegram poller service")
	err = pollerService.Start()
	if err != nil {
		panic(errors.Wrap(err, "cannot start poller"))
	}
	gLogger.Info("Server successfully started")
	waitingForShutdown()
}

func waitingForShutdown() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	gLogger.Infof("Received shutdown signal: %s", <-ch)
}