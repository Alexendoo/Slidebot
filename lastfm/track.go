package lastfm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Alexendoo/Slidebot/config"
	"github.com/Alexendoo/Slidebot/markdown"
	"github.com/Alexendoo/Slidebot/store"
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

func RecentTrack(args []string, s *discordgo.Session, m *discordgo.Message) {
	fmt.Println(args)

	username := getUsername(args, m)
	target := api("user.getrecenttracks", username)

	fmt.Println(target)

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
	err = json.Unmarshal(body, &trackJSON)
	if err != nil {
		fmt.Println(err)
		return
	}

	embed := buildEmbed(&trackJSON)
	if embed == nil {
		fmt.Println("nil embed")
		return
	}

	s.ChannelMessageSendEmbed(m.ChannelID, embed)
}

func getUsername(args []string, m *discordgo.Message) string {
	if len(args) > 0 {
		username := args[0]

		err := store.Set(store.BucketLastFM, m.Author.ID, username)
		if err != nil {
			fmt.Println(err)
		}

		return username
	}

	username, err := store.Get(store.BucketLastFM, m.Author.ID)
	if err != nil {
		fmt.Println(err)
	}

	// TODO: if username == ""

	return username
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

func buildEmbed(trackJSON *trackResponse) *discordgo.MessageEmbed {
	if len(trackJSON.RecentTracks.Track) == 0 {
		return nil
	}
	track := trackJSON.RecentTracks.Track[0]

	embed := &discordgo.MessageEmbed{

		Author: &discordgo.MessageEmbedAuthor{
			Name: markdown.Escape(track.Name),
			URL:  track.URL,
		},

		Title: fmt.Sprintf("By **%s**", markdown.Escape(track.Artist.Name)),
		URL:   track.Artist.URL,

		Description: fmt.Sprintf("On the album **%s**", markdown.Escape(track.Album.Name)),

		Color: 0xd50000,
	}

	if len(track.Images) > 0 {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: track.Images[len(track.Images)-1].URL,
		}
	}

	return embed
}
