package lastfm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Alexendoo/Slidebot/config"
	"github.com/bwmarrin/discordgo"
)

type images []struct {
	URL  string `json:"#text"`
	Size string `json:"size"`
}

type trackResponse struct {
	RecentTracks struct {
		Track []struct {
			Artist struct {
				Name   string `json:"name"`
				URL    string `json:"url"`
				Images images `json:"image"`
			} `json:"artist"`
			Loved string `json:"loved"`
			Name  string `json:"name"`
			Album struct {
				Name string `json:"#text"`
			} `json:"album"`
			URL    string `json:"url"`
			Images images `json:"image"`
			Attr   struct {
				Nowplaying string `json:"nowplaying"`
			} `json:"@attr,omitempty"`
		} `json:"track"`
		Attr struct {
			User string `json:"user"`
		} `json:"@attr"`
	} `json:"recenttracks"`
}

func api(method, username string) string {
	target := &url.URL{
		Scheme: "https",
		Host:   "ws.audioscrobbler.com",
		Path:   "/2.0/",
	}

	v := url.Values{}
	v.Set("method", method)
	v.Set("user", username)
	v.Set("api_key", config.Tokens.LastFM)
	v.Set("format", "json")
	v.Set("limit", "1")
	v.Set("extended", "1")

	target.RawQuery = v.Encode()

	return target.String()
}

func RecentTrack(args []string, s *discordgo.Session, m *discordgo.Message) {
	target := api("user.getrecenttracks", args[0])

	resp, err := http.Get(target)
	if err != nil {
		fmt.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var trackJSON trackResponse
	json.Unmarshal(body, &trackJSON)

	embed := buildEmbed(&trackJSON)
	if embed == nil {
		return
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func buildEmbed(trackJSON *trackResponse) *discordgo.MessageEmbed {
	if len(trackJSON.RecentTracks.Track) == 0 {
		return nil
	}
	track := trackJSON.RecentTracks.Track[0]

	embed := &discordgo.MessageEmbed{
		URL:   track.URL,
		Title: track.Name,
		Author: &discordgo.MessageEmbedAuthor{
			Name: track.Artist.Name,
			URL:  track.Artist.URL,
		},
		Color: 0xd50000,
	}

	if len(track.Images) > 0 {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: track.Images[len(track.Images)-1].URL,
		}
	}

	if len(track.Artist.Images) > 0 {
		embed.Author.IconURL = track.Artist.Images[len(track.Artist.Images)-1].URL
	}

	return embed
}
