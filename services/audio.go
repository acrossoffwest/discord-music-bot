package services

import (
	"github.com/jonas747/dca"
	"io"
	"log"
	"time"
)

func GlobalPlay(songSig chan PkgSong) {
	for {
		select {
		case song := <-songSig:
			if song.v.radioFlag {
				song.v.Stop()
				time.Sleep(200 * time.Millisecond)
			}
			go song.v.PlayQueue(song.data)
		}
	}
}

func GlobalRadio(radioSig chan PkgRadio) {
	for {
		select {
		case radio := <-radioSig:
			radio.v.Stop()
			time.Sleep(200 * time.Millisecond)
			go radio.v.Radio(radio.data)
		}
	}
}

func (v *main.VoiceInstance) PlayQueue(song main.Song) {
	// add song to queue
	v.QueueAdd(song)
	if v.speaking {
		// the bot is playing
		return
	}
	go func() {
		v.audioMutex.Lock()
		defer v.audioMutex.Unlock()
		for {
			if len(v.queue) == 0 {
				main.dg.UpdateGameStatus(0, main.o.DiscordStatus)
				ChMessageSend(v.nowPlaying.ChannelID, "[**Music**] End of queue!")
				return
			}
			v.nowPlaying = v.QueueGetSong()
			go ChMessageSend(v.nowPlaying.ChannelID, "[**Music**] Playing, **`"+
				v.nowPlaying.Title+"`  -  `("+v.nowPlaying.Duration+")`  -  **<@"+v.nowPlaying.ID+">\n") //*`"+ v.nowPlaying.User +"`***")
			// If monoserver
			if main.o.DiscordPlayStatus {
				main.dg.UpdateGameStatus(0, v.nowPlaying.Title)
			}
			v.stop = false
			v.skip = false
			v.speaking = true
			v.pause = false
			v.voice.Speaking(true)

			v.DCA(v.nowPlaying.VideoURL)

			v.QueueRemoveFisrt()
			if v.stop {
				v.QueueRemove()
			}
			v.stop = false
			v.skip = false
			v.speaking = false
			v.voice.Speaking(false)
		}
	}()
}

func (v *main.VoiceInstance) Radio(url string) {
	v.audioMutex.Lock()
	defer v.audioMutex.Unlock()
	if main.o.DiscordPlayStatus {
		main.dg.UpdateGameStatus(0, "Radio")
	}
	v.radioFlag = true
	v.stop = false
	v.speaking = true
	v.pause = false
	v.voice.Speaking(true)

	v.DCA(url)

	main.dg.UpdateGameStatus(0, main.o.DiscordStatus)
	v.radioFlag = false
	v.stop = false
	v.speaking = false
	v.voice.Speaking(false)
}

// DCA
func (v *main.VoiceInstance) DCA(url string) {
	opts := dca.StdEncodeOptions
	opts.RawOutput = true
	opts.Bitrate = 128
	opts.Application = "lowdelay"

	encodeSession, err := dca.EncodeFile(url, opts)
	if err != nil {
		log.Println("FATA: Failed creating an encoding session: ", err)
	}
	v.encoder = encodeSession
	done := make(chan error)
	stream := dca.NewStream(encodeSession, v.voice, done)
	v.stream = stream
	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				log.Println("FATA: An error occured", err)
			}
			// Clean up incase something happened and ffmpeg is still running
			encodeSession.Cleanup()
			return
		}
	}
}

// Stop stop the audio
func (v *main.VoiceInstance) Stop() {
	v.stop = true
	if v.encoder != nil {
		v.encoder.Cleanup()
	}
}

func (v *main.VoiceInstance) Skip() bool {
	if v.speaking {
		if v.pause {
			return true
		} else {
			if v.encoder != nil {
				v.encoder.Cleanup()
			}
		}
	}
	return false
}

// Pause pause the audio
func (v *main.VoiceInstance) Pause() {
	v.pause = true
	if v.stream != nil {
		v.stream.SetPaused(true)
	}
}

// Resume resume the audio
func (v *main.VoiceInstance) Resume() {
	v.pause = false
	if v.stream != nil {
		v.stream.SetPaused(false)
	}
}
