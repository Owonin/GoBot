package bot

import (
	"discord-bot/commands"
	"discord-bot/config"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	BotId string
	goBot *discordgo.Session

	helpMessage = "Help message is not avalible yet((("
)

func Start() {

	//creating new bot session
	goBot, err := discordgo.New("Bot " + config.Token)

	//Handling error
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Making our bot a user using User function .
	u, err := goBot.User("@me")
	//Handlinf error
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Storing our id from u to BotId .
	BotId = u.ID

	// Adding handler function to handle our messages using AddHandler from discordgo package. We will declare messageHandler function later.
	goBot.AddHandler(messageHandler)

	err = goBot.Open()
	//Error handling
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer goBot.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Terminating bot")

}

//Definition of messageHandler function it takes two arguments first one is discordgo.Session which is s , second one is discordgo.MessageCreate which is m.
func messageHandler(s *discordgo.Session, message *discordgo.MessageCreate) {
	//Bot musn't reply to it's own messages , to confirm it we perform this check.
	if message.Author.ID == BotId {
		return
	}
	querys := strings.Split(strings.ToLower(message.Content), " ")
	//command := strings.TrimSpace(message.Content)

	//command = strings.ToLower(command)

	//TODO Make message filter

	switch querys[0] {
	case "!health":
		_, _ = s.ChannelMessageSend(message.ChannelID, "is active")
	case "!ping":
		_, _ = s.ChannelMessageSend(message.ChannelID, "pong")
	case "!help":
		_, _ = s.ChannelMessageSend(message.ChannelID, helpMessage)
	case "!anime":
				commands.AnimeCommand.Exec(s, message, nil)
	case "!osu":
					commands.OsuCommand.Exec(s,message, querys[1:])
	case "!youtube":
					commands.YoutubeCommand.Exec(s, message, querys)
	}

	
}

