package commands

import (
	"discord-bot/config"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = make(map[string]Command)
)

type Command struct {
	CommandName string
	Help        string
	Fields      []*discordgo.MessageEmbedField
	Exec        func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

func NewCommand(c *Command) {
	commands[c.CommandName] = *c
}

func ParseCommand(s *discordgo.Session, message *discordgo.MessageCreate) {

	querys := strings.Split(message.Content, " ")

	commandName := strings.ToLower(querys[0])

	if !strings.HasPrefix(commandName, config.BotPrefix) {
		return
	}

	commandName = strings.Split(commandName, "!")[1]

	fmt.Sprintf("%s %s#%s,: %s", message.Author.ID, message.Author.Username,
		message.Author.Discriminator, message.Content)

	switch commandName {
	case "health":
		_, _ = s.ChannelMessageSend(message.ChannelID, "is active")
		return
	case "help":
		_, _ = s.ChannelMessageSend(message.ChannelID, commands[commandName].Help) //todo add toLower
	}

	command, ok := commands[commandName]

	if ok {
		command.Exec(s, message, querys[1:])
	} else {
		_, _ = s.ChannelMessageSend(message.ChannelID, "Command unknown")
	}

}

func SendEmmed(s *discordgo.Session, message *discordgo.MessageCreate,
	uri string, description string, title string, f *[]*discordgo.MessageEmbedField) error {
	_, err := s.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		URL:         uri,
		Type:        discordgo.EmbedTypeArticle,
		Title:       title,
		Description: description,
		Timestamp:   "",
		Color:       0,
		Footer:      &discordgo.MessageEmbedFooter{},
		Image:       &discordgo.MessageEmbedImage{},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{},
		Video:       &discordgo.MessageEmbedVideo{},
		Provider:    &discordgo.MessageEmbedProvider{},
		Author:      &discordgo.MessageEmbedAuthor{},
		Fields:      *f,
	}) // todo edit it
	return err
}
