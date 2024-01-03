package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/maxmwang/pecan/commands"
)

func main() {
	e := loadEnv()

	dg, err := discordgo.New("Bot " + e.botToken)
	if err != nil {
		panic(err)
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	dg.AddHandler(commands.Entry)

	err = dg.Open()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = dg.Close()
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
