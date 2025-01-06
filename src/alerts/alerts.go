package alerts

// package main

import (
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	discord_envs map[string]string
	sess         *discordgo.Session
}

var discord Discord

func Init() {
	var discord_envs = make(map[string]string)

	envs := []string{
		"DISCORD_TOKEN",
		"APPLICATION_ID",
		"GUILD_ID",
		"CHANNEL_ID",
		"MONITOR_THREAD",
		"ERROR_THREAD",
	}

	missing := []string{}
	for _, env := range envs {
		v := os.Getenv(env)

		if len(v) == 0 {
			missing = append(missing, env)
		} else {
			discord_envs[env] = v
		}
	}

	if len(missing) != 0 {
		log.Fatalf("Error missing the following env variables: %+v\n", missing)
	}

	sess, err := discordgo.New("Bot" + " " + discord_envs["DISCORD_TOKEN"])
	if err != nil {
		log.Fatal(err)
	}

	discord = Discord{
		sess:         sess,
		discord_envs: discord_envs,
	}
}

func Alert(message string, file_contents string, filetype string) error {
	var msg_send discordgo.MessageSend = discordgo.MessageSend{
    Content: message,
  }

	reader := strings.NewReader(file_contents)

	file := discordgo.File{
		Name:        filetype,
		ContentType: "text/plain",
		Reader:      reader,
	}

	msg_send.File = &file

	_, err := discord.sess.ChannelMessageSendComplex(
		discord.discord_envs["MONITOR_THREAD"],
		&msg_send,
	)

  if err != nil { return err }

  return nil
}
