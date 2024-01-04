package commands

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

var languageToImage = map[string]string{
	"py":         "python",
	"python":     "python",
	"js":         "node",
	"javascript": "node",
}

var imageToCmdPrefix = map[string][]string{
	"python": {"python", "-c"},
	"node":   {"node", "-e"},
}

func exec(s *discordgo.Session, m *discordgo.MessageCreate, c string) {
	if !strings.HasPrefix(c, "```") || !strings.HasSuffix(c, "```") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Invalid syntax. Use ``````<lang>\n<code>``````")
		return
	}

	raw0 := strings.TrimPrefix(c, "```")
	raw0 = strings.TrimSuffix(raw0, "```")
	raw1 := strings.SplitN(raw0, "\n", 2)
	lang := raw1[0]
	code := raw1[1]
	image, ok := languageToImage[lang]
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Invalid or unsupported language: "+lang)
		return
	}
	cmdPrefix := imageToCmdPrefix[image]

	logrus.WithFields(logrus.Fields{
		"time":   m.Timestamp,
		"author": m.Author.ID,
		"lang":   lang,
	}).Info("exec")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		reportErr(s, m, err)
		return
	}
	defer func() {
		_ = cli.Close()
	}()

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		reportErr(s, m, err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   append(cmdPrefix, code),
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

	_ = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"time":      m.Timestamp,
			"author":    m.Author.ID,
			"lang":      lang,
			"container": resp.ID,
		}).Error("failed to remove container")
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, out)
	if err != nil {
		reportErr(s, m, err)
		return
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "```"+buf.String()+"```")
}
