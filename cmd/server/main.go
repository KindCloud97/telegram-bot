package main

import (
	"fmt"
	"net"
	"os"

	"github.com/KindCloud97/telegram-bot/model"
	"github.com/KindCloud97/telegram-bot/queue"
	"github.com/KindCloud97/telegram-bot/service"
	proto "github.com/KindCloud97/telegram-bot/usersvc"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("some error occured:", err)
		os.Exit(1)
	}

	botToken := os.Getenv("TOKEN")
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println("could not create bot:", err)
		os.Exit(1)
	}

	s := service.NewService(bot, queue.NewQueue())

	var eg errgroup.Group
	eg.Go(func() error { return startServer(bot, s) })
	eg.Go(func() error { return startBot(bot, s) })
	if err := eg.Wait(); err != nil {
		fmt.Println("error occured", err)
		os.Exit(1)
	}
}

func startServer(bot *telego.Bot, srv proto.UserServiceServer) error {
	server := grpc.NewServer()
	proto.RegisterUserServiceServer(server, srv)

	r, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	err = server.Serve(r)
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func startBot(bot *telego.Bot, s *service.Service) error {
	updates, err := bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return fmt.Errorf("start long polling: %w", err)
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return fmt.Errorf("new bot handler: %w", err)
	}

	defer bh.Stop()
	defer bot.StopLongPolling()

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, err = bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			"Hello, I'm Allmight Bot!\nYou can send me text or image messages!\n\n Use /help for all available commands.",
		))
		if err != nil {
			fmt.Println("ERROR! could not send message:", err)
			return
		}
	}, th.CommandEqual("start"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, err = bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			"/connect to chat with operator.\n",
		))
		if err != nil {
			fmt.Println("ERROR! could not send message:", err)
			return
		}
	}, th.CommandEqual("help"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		s.AddToQueue(model.User{
			ChatId:  update.Message.Chat.ID,
			Name:    update.Message.From.FirstName,
			Surname: update.Message.From.LastName,
		})
	}, th.CommandEqual("connect"))

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		s.HandleMessage(update.Message)
	}, th.AnyMessage())

	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, err = bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			"Unknown command, use /help to list all available commands",
		))
		if err != nil {
			fmt.Println("ERROR! could not send message:", err)
			return
		}
	}, th.AnyCommand())

	bh.Start()

	return nil
}
