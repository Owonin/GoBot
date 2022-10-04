package commands

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/Strum355/go-queue/queue"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/kkdai/youtube/v2"
)

const (
	stdURL   = "https://www.youtube.com/watch"
	shortURL = "https://youtu.be/"
	embedURL = "https://www.youtube.com/embed/"
)

var (
	srvr = server{
		//LogChannel:  s.State.Guilds[len(s.State.Guilds)-1].ID,
		Log:         false,
		Nsfw:        false,
		JoinMessage: [3]string{"false", "", ""},
	}

	youtubeCommand = Command{
		CommandName: "youtube",
		Help:        "Воспроизведение видио в аудиочат",
		Exec:        msgYoutube,
	}
)

func init() {
	NewCommand(&youtubeCommand)
}

type voiceInst struct {
	ChannelID string

	Playing bool

	Done chan error

	*sync.RWMutex

	Queue            *queue.Queue
	VoiceCon         *discordgo.VoiceConnection
	StreamingSession *dca.StreamingSession
}

type song struct {
	URL   string `json:"url,omitempty"`
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`

	Duration time.Duration `json:"duration"`
}

func msgYoutube(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {

	fmt.Print(msglist[0])
	if len(msglist) == 1 {
		return
	}

	switch msglist[0] {
	case "play":
		addToQueue(s, m, msglist[1:])
	case "stop":
		stopQueue(s, m)
	case "list", "queue", "songs":
		listQueue(s, m)
	case "pause":
		pauseQueue(s, m)
	case "resume", "unpause":
		unpauseQueue(s, m)
	case "skip", "next":
		skipSong(s, m)
	default:
	}
}

func addToQueue(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {

	fmt.Println(msglist[len(msglist)-1])
	if len(msglist) == 0 {
		return
	}

	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was a problem adding to queue :( please try again")
		return
	}

	if srvr.VoiceInst == nil {
		srvr.newVoiceInstance()
	}

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	url := msglist[0]

	if !strings.HasPrefix(url, stdURL) && !strings.HasPrefix(url, shortURL) && !strings.HasPrefix(url, embedURL) {
		s.ChannelMessageSend(m.ChannelID, "Please make sure the URL is a valid YouTube URL. If I got this wrong, please let my creator know ~owo")
		return
	}

	vid, err := getVideoInfo(url, s, m)
	if err != nil {
		return
	}

	vc, err := createVoiceConnection(s, m, guild, &srvr)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Need to be in a voice channel!")
		return
	}

	srvr.addSong(song{
		URL:      url,
		Name:     vid.Title,
		Duration: vid.Duration,
		//Image: vid.Thumbnails
		//Image:    vid.GetThumbnailURL(ytdl.ThumbnailQualityMedium).String(),
	})

	s.ChannelMessageSend(m.ChannelID, "Добавлено "+vid.Title+" в список воспроизведений!")

	if !srvr.VoiceInst.Playing {
		srvr.VoiceInst.VoiceCon = vc
		srvr.VoiceInst.Playing = true
		srvr.VoiceInst.ChannelID = vc.ChannelID
		go play(s, m, &srvr, vc)
	}

}

func createVoiceConnection(s *discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild, srvr *server) (*discordgo.VoiceConnection, error) {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID && (vs.ChannelID == srvr.VoiceInst.ChannelID || !srvr.VoiceInst.Playing) {
			vc, err := s.ChannelVoiceJoin(guild.ID, vs.ChannelID, false, true)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error joining voice channel")
				log.Panic("error joining voice channel", err)
				return nil, err
			}
			return vc, nil
		}
	}
	return nil, errors.New("not in voice channel")
}

func getVideoInfo(url string, s *discordgo.Session, m *discordgo.MessageCreate) (*youtube.Video, error) {
	client := youtube.Client{}

	vid, err := client.GetVideo(url)
	if err != nil {
		panic(err)
	}

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error getting video info")
		fmt.Print(err)
		return nil, err
	}
	return vid, nil
}

func play(s *discordgo.Session, m *discordgo.MessageCreate, srvr *server, vc *discordgo.VoiceConnection) {
	if srvr.queueLength() == 0 {
		srvr.youtubeCleanup()
		s.ChannelMessageSend(m.ChannelID, "🔇 Список видио закончен!")
		return
	}

	fmt.Println("Downloading video")
	srvr.VoiceInst.Lock()
	vid, err := getVideoInfo(srvr.nextSong().URL, s, m)
	if err != nil {
		srvr.VoiceInst.Unlock()
		return
	}

	client := youtube.Client{}
	format := vid.Formats.WithAudioChannels()

	stream, size, err := client.GetStream(vid, &format[0])
	fmt.Println(size)
	defer stream.Close()

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, " Error downloading the music")
		log.Panic("error downloading YouTube video", err)
		srvr.VoiceInst.Done <- err
		return
	}

	encSesh, err := dca.EncodeMem(stream, dca.StdEncodeOptions)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, " Error starting the stream")
		srvr.youtubeCleanup()
		srvr.VoiceInst.Unlock()
		log.Panic("error validating options", err)
		return
	}
	defer encSesh.Cleanup()

	srvr.VoiceInst.StreamingSession = dca.NewStream(encSesh, vc, srvr.VoiceInst.Done)

	s.ChannelMessageSend(m.ChannelID, "🔊 Играет: "+vid.Title)

	srvr.VoiceInst.Unlock()

Outer:
	for {
		err = <-srvr.VoiceInst.Done

		done, _ := srvr.VoiceInst.StreamingSession.Finished()

		switch {
		case err.Error() == "stop":
			srvr.youtubeCleanup()
			s.ChannelMessageSend(m.ChannelID, "🔇 Остановлено")
			return
		case err.Error() == "skip":
			s.ChannelMessageSend(m.ChannelID, "⏩ Пропуск")
			break Outer
		case !done && err != io.EOF:
			srvr.youtubeCleanup()
			s.ChannelMessageSend(m.ChannelID, "There was an error streaming music :(")
			log.Panic("error streaming music", err)
			return
		case done && err == io.EOF:
			// Remove the currently playing song from the queue and then start the next one
			srvr.finishedSong()
			break Outer
		}
	}

	go play(s, m, srvr, vc)
}

func listQueue(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "There was an issue loading the list :( please try again")
		return
	}

	if srvr.queueLength() == 0 {
		s.ChannelMessageSend(m.ChannelID, "Нету видио в очереди!")
		return
	}

	p := dgwidgets.NewPaginator(s, m.ChannelID)
	p.Add(&discordgo.MessageEmbed{
		Title: guild.Name + "'s queue",

		Fields: func() (out []*discordgo.MessageEmbedField) {
			for i, song := range srvr.iterateQueue() {
				out = append(out, &discordgo.MessageEmbedField{
					Name:  fmt.Sprintf("%d - %s", i, song.Name),
					Value: song.Duration.String(),
				})
			}
			return
		}(),
	})

	for _, song := range srvr.iterateQueue() {
		p.Add(&discordgo.MessageEmbed{
			Title: fmt.Sprintf("Title: %s\nDuration: %s\nURL: %s", song.Name, song.Duration, song.URL),

			Image: &discordgo.MessageEmbedImage{
				URL: song.Image,
			},
		})
	}

	p.SetPageFooters()
	p.Loop = true
	p.ColourWhenDone = 0xff0000
	p.DeleteReactionsWhenDone = true
	p.Widget.Timeout = time.Minute * 2
	p.Spawn()
}

func stopQueue(s *discordgo.Session, m *discordgo.MessageCreate) {

	srvr.VoiceInst.Done <- errors.New("stop")

}

func pauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()

	s.ChannelMessageSend(m.ChannelID, "⏸ Пауза. Для продолжения используйте команду `!yt unpause`")

	srvr.VoiceInst.StreamingSession.SetPaused(true)
}

func unpauseQueue(s *discordgo.Session, m *discordgo.MessageCreate) {

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	srvr.VoiceInst.StreamingSession.SetPaused(false)

}

func skipSong(s *discordgo.Session, m *discordgo.MessageCreate) {

	srvr.VoiceInst.Lock()
	defer srvr.VoiceInst.Unlock()
	srvr.VoiceInst.Done <- errors.New("skip")

}

func (s *server) youtubeCleanup() {
	s.VoiceInst.Lock()
	defer s.VoiceInst.Unlock()
	s.VoiceInst.VoiceCon.Disconnect()
	s.newVoiceInstance()
	//sMap.VoiceInsts--
}

// add to another file

func guildDetails(channelID, guildID string, s *discordgo.Session) (guildDetails *discordgo.Guild, err error) {
	if guildID == "" {
		var channel *discordgo.Channel
		channel, err = channelDetails(channelID, s)
		if err != nil {
			return
		}

		guildID = channel.GuildID
	}

	guildDetails, err = s.State.Guild(guildID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			guildDetails, err = s.Guild(guildID)
			if err != nil {
				fmt.Print(err)
			}
		}
	}
	return
}

func channelDetails(channelID string, s *discordgo.Session) (channelDetails *discordgo.Channel, err error) {
	channelDetails, err = s.State.Channel(channelID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			channelDetails, err = s.Channel(channelID)
			if err != nil {
				fmt.Print(err)
			}
		}
	}
	return
}
