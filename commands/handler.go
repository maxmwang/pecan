package commands

import (
	"context"
	"io"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const commandPrefix = ">"

func Entry(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, commandPrefix) {
		return
	}

	if strings.HasPrefix(m.Content, ">exec") {
		ctx := context.Background()
		code := strings.TrimPrefix(m.Content, ">exec\n```\n")
		code = strings.TrimSuffix(code, "\n```")

		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			reportErr(s, m, err)
			return
		}
		defer func() {
			err = cli.Close()
			if err != nil {
				reportErr(s, m, err)
				return
			}
		}()

		reader, err := cli.ImagePull(ctx, "python", types.ImagePullOptions{})
		if err != nil {
			reportErr(s, m, err)
			return
		}
		defer func() {
			err = reader.Close()
			if err != nil {
				reportErr(s, m, err)
				return
			}
		}()

		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: "python",
			Cmd:   []string{"python", "-c", code},
			Tty:   false,
		}, nil, nil, nil, "")
		if err != nil {
			reportErr(s, m, err)
			return
		}

		if err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			reportErr(s, m, err)
			return
		}

		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err = <-errCh:
			if err != nil {
				reportErr(s, m, err)
				return
			}
		case <-statusCh:
		}

		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			reportErr(s, m, err)
			return
		}

		buf := new(strings.Builder)
		_, err = io.Copy(buf, out)
		if err != nil {
			reportErr(s, m, err)
			return
		}
		_, err = s.ChannelMessageSend(m.ChannelID, "```"+buf.String()+"```")
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
