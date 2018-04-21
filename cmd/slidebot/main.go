package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Alexendoo/Slidebot/config"
	"github.com/Alexendoo/Slidebot/lastfm"
	"github.com/Alexendoo/Slidebot/store"
	"github.com/bwmarrin/discordgo"
)

func main() {
	err := config.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = store.Open("bolt.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	dg, err := discordgo.New("Bot " + config.Tokens.Discord)
	if err != nil {
		fmt.Println(err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dg.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Stopping...")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("%s (%s): %s\n", m.Author.Username, m.Author.ID, m.Content)

	if m.Author.ID == s.State.User.ID {
		return
	}

	words := strings.Fields(m.Content)
	if len(words) == 0 {
		return
	}

	switch words[0][1:] {
	case "l":
		lastfm.RecentTrack(words[1:], s, m.Message)
	case "echo":
		s.ChannelMessageSend(m.ChannelID, m.Content)
	}
}
