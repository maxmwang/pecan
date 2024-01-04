package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

const commandPrefix = ">"

func Entry(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, commandPrefix) {
		return
	}

	if strings.HasPrefix(m.Content, commandPrefix+"exec") {
		c := strings.TrimPrefix(m.Content, commandPrefix+"exec\n")
		exec(s, m, c)
	}

	if strings.HasPrefix(m.Content, ">ping") {
		_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			return
		}
	}
}

func reportErr(s *discordgo.Session, m *discordgo.MessageCreate, err error) {
	_, err = s.ChannelMessageSend(m.ChannelID, err.Error())
	if err != nil {
		return
	}
}
