package commands

import "github.com/bwmarrin/discordgo"

type Command struct{
				Name string
				Help string
				Fields []*discordgo.MessageEmbedField
				Exec func(*discordgo.Session, *discordgo.MessageCreate, []string)
				
}

func SendEmmed (s *discordgo.Session, message *discordgo.MessageCreate,
	uri string, description string, title string, f *[]*discordgo.MessageEmbedField) error{
				_, err := s.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{                                                                         							
					  URL:        uri,
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
						Fields: *f,
				  }) // todo edit it                                                                                                                                      							
					return err 
	}                                                                                                                                                 
	