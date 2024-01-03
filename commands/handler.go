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

	if strings.HasPrefix(m.Content, ">ping") {
		_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			return
		}
	}
}
