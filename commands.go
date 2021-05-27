package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strconv"
	"strings"
	"time"
)

// HelpReporter
func HelpReporter(m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'help'")
	help := "```go\n`Standard Commands List`\n```\n" +
		"**`" + o.DiscordPrefix + "help`** or **`" + o.DiscordPrefix + "h`**  ->  show help commands.\n" +
		"**`" + o.DiscordPrefix + "join`** or **`" + o.DiscordPrefix + "j`**  ->  the bot join in to voice channel.\n" +
		"**`" + o.DiscordPrefix + "leave`** or **`" + o.DiscordPrefix + "l`**  ->  the bot leave the voice channel.\n" +
		"**`" + o.DiscordPrefix + "play`**  ->  play and add a one song in the queue.\n" +
		"**`" + o.DiscordPrefix + "radio`**  ->  play a URL radio.\n" +
		"**`" + o.DiscordPrefix + "stop`**  ->  stop the player and remove the queue.\n" +
		"**`" + o.DiscordPrefix + "skip`**  ->  skip the actual song and play the next song of the queue.\n" +
		"**`" + o.DiscordPrefix + "pause`**  ->  pause the player.\n" +
		"**`" + o.DiscordPrefix + "resume`**  ->  resume the player.\n" +
		"**`" + o.DiscordPrefix + "time`**  ->  show the time remaining of song.\n" +
		"**`" + o.DiscordPrefix + "queue list`**  ->  show the list of song in the queue.\n" +
		"**`" + o.DiscordPrefix + "queue remove `**  ->  remove a song of queue indexed for a ***number***, an ***@User*** or the ***last*** song, i.e. ***" + o.DiscordPrefix + "queue remove 2***\n" +
		"**`" + o.DiscordPrefix + "queue clean`**  ->  clean all queue.\n" +
		"**`" + o.DiscordPrefix + "youtube`**  ->  search from youtube.\n\n" +
		"```go\n`Owner Commands List`\n```\n" +
		"**`" + o.DiscordPrefix + "ignore`**  ->  ignore commands of a channel.\n" +
		"**`" + o.DiscordPrefix + "unignore`**  ->  unignore commands of a channel.\n"

	ChMessageSend(m.ChannelID, help)
	//ChMessageSendEmbed(m.ChannelID, "Help", help)
}

// JoinReporter
func JoinReporter(v *VoiceInstance, m *discordgo.MessageCreate, s *discordgo.Session) {
	log.Println("INFO:", m.Author.Username, "send 'join'")
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if voiceChannelID == "" {
		log.Println("ERROR: Voice channel id not found.")
		ChMessageSend(m.ChannelID, "[**Music**] <@"+m.Author.ID+"> You need to join a voice channel!")
		return
	}
	if v != nil {
		log.Println("INFO: Voice Instance already created.")
	} else {
		v = CreateVoiceInstance(m, s)
	}
	var err error
	v.voice, err = dg.ChannelVoiceJoin(v.guildID, voiceChannelID, false, false)
	if err != nil {
		v.Stop()
		log.Println("ERROR: Error to join in a voice channel: ", err)
		return
	}
	v.voice.Speaking(false)
	log.Println("INFO: New Voice Instance created")
	ChMessageSend(m.ChannelID, "[**Music**] I've joined a voice channel!")
}

func CreateVoiceInstance(m *discordgo.MessageCreate, s *discordgo.Session) *VoiceInstance {
	guildID := SearchGuild(m.ChannelID)
	// create new voice instance
	mutex.Lock()
	v := new(VoiceInstance)
	voiceInstances[guildID] = v
	v.guildID = guildID
	v.session = s
	mutex.Unlock()
	//v.InitVoice()
	return v
}

// LeaveReporter
func LeaveReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'leave'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	v.Stop()
	time.Sleep(200 * time.Millisecond)
	v.voice.Disconnect()
	log.Println("INFO: Voice channel destroyed")
	mutex.Lock()
	delete(voiceInstances, v.guildID)
	mutex.Unlock()
	dg.UpdateGameStatus(0, o.DiscordStatus)
	ChMessageSend(m.ChannelID, "[**Music**] I left the voice channel!")
}

// PlayReporter
func PlayReporter(v *VoiceInstance, m *discordgo.MessageCreate, s *discordgo.Session) {
	log.Println("INFO:", m.Author.Username, "send 'play'")
	if v == nil {
		v = CreateVoiceInstance(m, s)
		JoinReporter(v, m, s)
	}
	if len(strings.Fields(m.Content)) < 2 {
		ChMessageSend(m.ChannelID, "[**Music**] You need specify a name or URL.")
		return
	}
	// if the user is not a voice channel not accept the command
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if v.voice.ChannelID != voiceChannelID {
		ChMessageSend(m.ChannelID, "[**Music**] <@"+m.Author.ID+"> You need to join in my voice channel for send play!")
		return
	}
	// send play my_song_youtube
	command := strings.SplitAfter(m.Content, strings.Fields(m.Content)[0])
	query := strings.TrimSpace(command[1])
	song, err := YoutubeFind(query, v, m)
	if err != nil || song.data.ID == "" {
		log.Println("ERROR: Youtube search: ", err)
		ChMessageSend(m.ChannelID, "[**Music**] I can't found this song!")
		return
	}
	//***`"+ song.data.User +"`***
	ChMessageSend(m.ChannelID, "[**Music**] **`"+song.data.User+"`** has added , **`"+
		song.data.Title+"`** to the queue. **`("+song.data.Duration+")` `["+strconv.Itoa(len(v.queue))+"]`**")
	go func() {
		songSignal <- song
	}()
}

func PlayPlaylistReporter(v *VoiceInstance, m *discordgo.MessageCreate, s *discordgo.Session) {
	log.Println("INFO:", m.Author.Username, "send 'playlist'")
	if v == nil {
		v = CreateVoiceInstance(m, s)
		JoinReporter(v, m, s)
	}
	if len(strings.Fields(m.Content)) < 2 {
		ChMessageSend(m.ChannelID, "[**Music**] You need specify a name or URL.")
		return
	}
	// if the user is not a voice channel not accept the command
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if v.voice.ChannelID != voiceChannelID {
		ChMessageSend(m.ChannelID, "[**Music**] <@"+m.Author.ID+"> You need to join in my voice channel for send play!")
		return
	}
	// send play my_song_youtube
	command := strings.SplitAfter(m.Content, strings.Fields(m.Content)[0])
	query := strings.TrimSpace(command[1])
	videos := LoadPlaylist(query)

	if len(videos) == 0 {
		ChMessageSend(m.ChannelID, "[**Music**] I can't found playlist! ")
		return
	}
	message := ""
	for _, video := range videos {
		song, err := YoutubeFind(video.Title, v, m)
		if err != nil {
			continue
		} else if song.data.ID == "" {
			log.Println("ERROR: Youtube search: ", err)
			message += "\n**" + video.Title + "** - song not found"
			continue
		}
		//***`"+ song.data.User +"`***
		message += "\n** " + strconv.Itoa(len(v.queue)) + ")" + video.Title + "(" + song.data.Duration + ")** - song added"
		go func() {
			songSignal <- song
		}()
		time.Sleep(3 * time.Second)
	}
	ChMessageSend(m.ChannelID, "[**Music**] **` New songs added to queue: `"+message)
}

// ReadioReporter
func RadioReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'radio'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	if len(strings.Fields(m.Content)) < 2 {
		ChMessageSend(m.ChannelID, "[**Music**] You need to specify a url!")
		return
	}
	radio := PkgRadio{"", v}
	radio.data = strings.Fields(m.Content)[1]

	go func() {
		radioSignal <- radio
	}()
	ChMessageSend(m.ChannelID, "[**Music**] **`"+m.Author.Username+"`** I'm playing a radio now!")
}

// StopReporter
func StopReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'stop'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	voiceChannelID := SearchVoiceChannel(m.Author.ID)
	if v.voice.ChannelID != voiceChannelID {
		ChMessageSend(m.ChannelID, "[**Music**] <@"+m.Author.ID+"> You need to join in my voice channel for send stop!")
		return
	}
	v.Stop()
	dg.UpdateGameStatus(0, o.DiscordStatus)
	log.Println("INFO: The bot stop play audio")
	ChMessageSend(m.ChannelID, "[**Music**] I'm stoped now!")
}

// PauseReporter
func PauseReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'pause'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		return
	}
	if !v.speaking {
		ChMessageSend(m.ChannelID, "[**Music**] I'm not playing nothing!")
		return
	}
	if !v.pause {
		v.Pause()
		ChMessageSend(m.ChannelID, "[**Music**] I'm `PAUSED` now!")
	}
}

// ResumeReporter
func ResumeReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'resume'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	if !v.speaking {
		ChMessageSend(m.ChannelID, "[**Music**] I'm not playing nothing!")
		return
	}
	if v.pause {
		v.Resume()
		ChMessageSend(m.ChannelID, "[**Music**] I'm `RESUMED` now!")
	}
}

// TimeReporter
func TimeReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'time'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	if v.speaking == true && v.radioFlag == false {
		var duration TimeDuration
		var message string
		if v.stream != nil {
			d := v.stream.PlaybackPosition()
			duration.Second = int(d.Seconds())
			t := AddTimeDuration(duration)

			if len(strings.Split(v.nowPlaying.Duration, ":")) == 2 {
				message = fmt.Sprintf("[**Music**] The playback time of **`%s`**  is  **`(%d:%02d)`**  of  **`(%s)`**  -  **`%s`**",
					v.nowPlaying.Title, t.Minute, t.Second, v.nowPlaying.Duration, v.nowPlaying.User)
			} else if len(strings.Split(v.nowPlaying.Duration, ":")) == 3 {
				message = fmt.Sprintf("[**Music**] The playback time of **`%s`**  is  **`(%d:%02d:%02d)`**  of  **`(%s)`**  -  **`%s`**",
					v.nowPlaying.Title, t.Hour, t.Minute, t.Second, v.nowPlaying.Duration, v.nowPlaying.User)
			} else if len(strings.Split(v.nowPlaying.Duration, ":")) == 4 {
				message = fmt.Sprintf("[**Music**] The playback time of **`%s`**  is  **`(%d:%02d:%02d:%02d)`**  of  **`(%s)`**  -  **`%s`**",
					v.nowPlaying.Title, t.Day, t.Hour, t.Minute, t.Second, v.nowPlaying.Duration, v.nowPlaying.User)
			}
			ChMessageSend(m.ChannelID, message)
			return
		}
	}
}

// QueueReporter
func QueueReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'queue'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	if len(v.queue) == 0 {
		log.Println("INFO: The queue is empty.")
		ChMessageSend(m.ChannelID, "[**Music**] The song queue is empty!")
		return
	}
	if len(strings.Fields(m.Content)) < 2 {
		ChMessageSend(m.ChannelID, "[**Music**] You need specify a `sub-command`!")
		return
	}
	if strings.HasSuffix(m.Content, "queue clean") {
		log.Println("INFO:", m.Author.Username, "send 'queue clean'")
		v.QueueClean()
		ChMessageSend(m.ChannelID, "[**Music**] Queue cleaned")
		return
	}
	if strings.Contains(m.Content, "queue remove") {
		voiceChannelID := SearchVoiceChannel(m.Author.ID)
		if v.voice.ChannelID != voiceChannelID {
			ChMessageSend(m.ChannelID, "[**Music**] <@"+m.Author.ID+"> You need to join in my voice channel to remove of queue!")
			return
		}
		log.Println("INFO:", m.Author.Username, "send 'queue remove'")
		if len(strings.Fields(m.Content)) != 3 {
			ChMessageSend(m.ChannelID, "[**Music**] You need define a `number`, an `@User` or `last` command")
			return
		}
		// is a number?
		if k, err := strconv.Atoi(strings.Fields(m.Content)[2]); err == nil {
			if k < len(v.queue) && k != 0 {
				song := v.queue[k]
				v.QueueRemoveIndex(k)
				ChMessageSend(m.ChannelID, "[**Music**] The songs  **`["+strconv.Itoa(k)+"]`  -  `"+song.Title+"`**  was removed of queue!")
				return
			} else {
				ChMessageSend(m.ChannelID, "[**Music**] The songs **`["+strconv.Itoa(k)+"]`** not exist!")
				return
			}
		}
		// is an user?
		if len(m.Mentions) != 0 {
			v.QueueRemoveUser(m.Mentions[0].Username)
			ChMessageSend(m.ChannelID, "[**Music**] The songs indexed by **`"+m.Mentions[0].Username+"`** was removed of queue!")
			return
		}
		// the `last` song?
		if strings.HasSuffix(m.Content, "queue remove last") {
			log.Println("INFO:", m.Author.Username, "send 'queue remove last'")
			if len(v.queue) > 1 {
				v.QueueRemoveLast()
				ChMessageSend(m.ChannelID, "[**Music**] The last songs indexed was removed of queue!")
				return
			}
			ChMessageSend(m.ChannelID, "[**Music**] No more songs in the queue!")
			return
		}

	}
	// queue list
	if strings.HasSuffix(m.Content, "queue list") {
		log.Println("INFO:", m.Author.Username, "send 'queue list'")
		message := "[**Music**] My songs are:\n\nNow Playing: **`" + v.nowPlaying.Title + "`  -  `(" +
			v.nowPlaying.Duration + ")`  -  " + v.nowPlaying.User + "**\n"

		queue := v.queue[1:]
		if len(queue) != 0 {
			var duration TimeDuration
			for i, q := range queue {
				message = message + "\n**`[" + strconv.Itoa(i+1) + "]`  -  `" + q.Title + "`  -  `(" + q.Duration + ")`  -  " + q.User + "**"
				d := strings.Split(q.Duration, ":")

				switch len(d) {
				case 2:
					// mm:ss
					ss, _ := strconv.Atoi(d[1])
					duration.Second = duration.Second + ss
					mm, _ := strconv.Atoi(d[0])
					duration.Minute = duration.Minute + mm
				case 3:
					// hh:mm:ss
					ss, _ := strconv.Atoi(d[2])
					duration.Second = duration.Second + ss
					mm, _ := strconv.Atoi(d[1])
					duration.Minute = duration.Minute + mm
					hh, _ := strconv.Atoi(d[0])
					duration.Hour = duration.Hour + hh
				case 4:
					// dd:hh:mm:ss
					ss, _ := strconv.Atoi(d[3])
					duration.Second = duration.Second + ss
					mm, _ := strconv.Atoi(d[2])
					duration.Minute = duration.Minute + mm
					hh, _ := strconv.Atoi(d[1])
					duration.Hour = duration.Hour + hh
					dd, _ := strconv.Atoi(d[0])
					duration.Day = duration.Day + dd
				}
			}
			t := AddTimeDuration(duration)
			message = message + "\n\nThe total duration: **`" +
				strconv.Itoa(t.Day) + "d` `" +
				strconv.Itoa(t.Hour) + "h` `" +
				strconv.Itoa(t.Minute) + "m` `" +
				strconv.Itoa(t.Second) + "s`**"
		}
		ChMessageSend(m.ChannelID, message)
		return
	}
}

// SkipReporter
func SkipReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'skip'")
	if v == nil {
		log.Println("INFO: The bot is not joined in a voice channel")
		ChMessageSend(m.ChannelID, "[**Music**] I need join in a voice channel!")
		return
	}
	if len(v.queue) == 0 {
		log.Println("INFO: The queue is empty.")
		ChMessageSend(m.ChannelID, "[**Music**] Currently there's no music playing, add some? ;)")
		return
	}
	if v.Skip() {
		ChMessageSend(m.ChannelID, "[**Music**] I'm `PAUSED`, please `resume` first.")
	}
}

// YoutubeReporter
func YoutubeReporter(v *VoiceInstance, m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'youtube'")
	command := strings.SplitAfter(m.Content, strings.Fields(m.Content)[0])
	query := strings.TrimSpace(command[1])
	song, err := YoutubeFind(query, v, m)
	if err != nil || song.data.ID == "" {
		log.Println("ERROR: Youtube search: ", err)
		ChMessageSend(m.ChannelID, "[**Music**] I can't found this song!")
		return
	}
	ChMessageSendHold(m.ChannelID, "[**Music**] **`"+song.data.User+"`**, Youtube URL: https://www.youtube.com/watch?v="+song.data.VidID)
}

// ShowRolesWithIdAndName
func ShowRolesWithIdAndName(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildID := SearchGuild(m.ChannelID)
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return
	}
	result := ""
	for _, role := range roles {
		if role.Name == "@everyone" {
			continue
		}
		result += fmt.Sprintf("%v -> %v\n", role.ID, role.Name)
	}
	ChMessageSendHold(m.ChannelID, result)
}

// Not used for now
// StatusReporter
func StatusReporter(m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'status'")
	if len(strings.Fields(m.Content)) < 2 {
		ChMessageSend(m.ChannelID, "[**Music**] You need to specify a status!")
		return
	}
	command := strings.SplitAfter(m.Content, "status")
	status := strings.TrimSpace(command[1])
	dg.UpdateGameStatus(0, status)
	ChMessageSend(m.ChannelID, "[**Music**] Status: `"+status+"`")
}

// StatusCleanReporter
func StatusCleanReporter(m *discordgo.MessageCreate) {
	log.Println("INFO:", m.Author.Username, "send 'statusclean'")
	dg.UpdateGameStatus(0, "")
}

func AddMessageForSelectRoles(m *discordgo.MessageCreate) {
	message := "Ролёо абхын тулоо эмодзи дарагты / For selecting your role click on an emoji / Для выбора роли выберите эмодзи: \nБуряад <:buryadtug:827123053799931934>\nХалха <:halhatug:847545621606563870>\nХальмаг <:halmagtug:847545511270547468>\nОрод <:orodtug:847558068283375647>\nХужаа <:huzhaatug:847558624711540756>"
	msg := ChMessageSendWithoutPurge(m.ChannelID, message)
	if msg == nil {
		return
	}
	log.Println("Message ID:", msg.ID)
}
