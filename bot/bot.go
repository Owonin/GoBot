package bot

import (
	"discord-bot/commands"
	"discord-bot/config"
	"fmt"
	"log"
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
		animeList := *commands.GetAnimeList()
		for _, anime := range animeList {

			_, _ = s.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Название: %s Ссылка: ||<%s>||", anime.Title, anime.Url))

		}

	case "!osu":

		// _, _ = s.ChannelMessageSend(message.ChannelID, fmt.Sprintf(` `, ))
		if len(querys) == 1 {
			_, _ = s.ChannelMessageSend(message.ChannelID, "Введите имя пользователя")
		} else {
			var data commands.OsuUserDetail = *commands.GetOsuUser(querys[1])
			//= command.GetOsuUser("amd") //GetOsuUser(querys[1])
			_, err := s.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
				URL:         fmt.Sprintf("https://osu.ppy.sh/users/%d", data.Id),
				Type:        discordgo.EmbedTypeArticle,
				Title:       "Osu! stats",
				Description: fmt.Sprintf("Статистика игрока **%s**", data.Username),
				Timestamp:   "",
				Color:       0,
				Footer:      &discordgo.MessageEmbedFooter{},
				Image:       &discordgo.MessageEmbedImage{},
				Thumbnail:   &discordgo.MessageEmbedThumbnail{},
				Video:       &discordgo.MessageEmbedVideo{},
				Provider:    &discordgo.MessageEmbedProvider{},
				Author:      &discordgo.MessageEmbedAuthor{},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Время в игре",
						Value:  fmt.Sprintf("**%d** часов", data.UserStatistic.Playtime/3600),
						Inline: false,
					},
					{
						Name:   "Ранг",
						Value:  fmt.Sprintf("Мировой ранг - **%d**, Ранг в стране - **%d**.\n", data.UserStatistic.GlobalRank, data.UserStatistic.CountryRank),
						Inline: false,
					},

					{
						Name:   "Статистика",
						Value:  fmt.Sprintf("Точность - **%f**, PP - **%d**", data.UserStatistic.HitAccuracy, int(data.UserStatistic.Pp)),
						Inline: false,
					},

					{
						Name:   "Онлайн",
						Value:  fmt.Sprintf("Онлайн - **%t**,\n Время последнего входа в онлайн - **%s**", data.Is_online, data.Last_visit.Format("2 Jan 2006 15:04:05")),
						Inline: false,
					},
				},
			}) // todo edit it

			if err != nil {
				log.Panic(err)
			}
		}
	}

}
