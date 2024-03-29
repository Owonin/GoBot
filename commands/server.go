package commands

import (
	"sync"

	"github.com/Strum355/go-queue/queue"
)
type servers struct {
	Count      int `json:"-"`
	VoiceInsts int `json:"-"`

	Mutex sync.RWMutex `json:"-"`

	serverMap map[string]*server
}

func (s *servers) server(id string) (val *server, ok bool) {
	val, ok = s.serverMap[id]
	return
}

func (s *servers) setServer(id string, serv server) {
	s.serverMap[id] = &serv
}
type server struct {
	LogChannel string `json:"log_channel"`
	Prefix     string `json:"server_prefix,omitempty"`

	Log    bool `json:"log_active"`
	Kicked bool `json:"kicked"`
	Nsfw   bool `json:"nsfw"`

	//Enabled, Message, Channel
	JoinMessage [3]string `json:"join"`

	VoiceInst *voiceInst `json:"-"`

	Playlists map[string][]song `json:"playlists"`
}


func (s *server) newVoiceInstance() {
	s.VoiceInst = &voiceInst{
		Queue:   queue.New(),
		Done:    make(chan error),
		RWMutex: new(sync.RWMutex),
	}
}

func (s server) nextSong() song {
	return s.VoiceInst.Queue.Front().(song)
}

func (s server) finishedSong() {
	s.VoiceInst.Queue.PopFront()
}

func (s server) addSong(song song) {
	s.VoiceInst.Queue.PushBack(song)
}

func (s server) queueLength() int {
	s.VoiceInst.RLock()
	defer s.VoiceInst.RUnlock()
	return s.VoiceInst.Queue.Len()
}

func (s server) iterateQueue() []song {
	s.VoiceInst.RLock()
	defer s.VoiceInst.RUnlock()
	ret := make([]song, s.VoiceInst.Queue.Len())
	for i, val := range s.VoiceInst.Queue.List() {
		ret[i] = val.(song)
	}
	return ret
}

