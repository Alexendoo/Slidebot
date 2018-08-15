package github

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Alexendoo/Slidebot/config"
	"github.com/Alexendoo/Slidebot/github/templates"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
)

type Handler struct {
	Discord *discordgo.Session
}

type GenericEvent interface {
	GetAction() string
	GetRepo() *github.Repository
	GetSender() *github.User
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	webhookType := github.WebHookType(r)
	event, err := github.ParseWebHook(webhookType, payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event := event.(type) {
	case GenericEvent:
		h.handleGeneric(event, webhookType)
	default:
		log.Printf("unhandled event: %#+v\n", event)
	}
}

func (h *Handler) handleGeneric(event GenericEvent, name string) {
	fullname := *event.GetRepo().FullName
	target, ok := config.Repos[fullname]
	if !ok {
		log.Printf("unhandled repo: %s\n", fullname)
		return
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    *event.GetSender().Login,
			URL:     *event.GetSender().HTMLURL,
			IconURL: *event.GetSender().AvatarURL,
		},

		Title: templates.Exec(event, name, event.GetAction(), "title"),
		URL:   templates.Exec(event, name, event.GetAction(), "URL"),
	}

	log.Printf("embed.Title: %#+v\n", embed.Title)
	log.Printf("embed.URL: %#+v\n", embed.URL)

	_, err := h.Discord.ChannelMessageSendEmbed(target, embed)

	if err != nil {
		log.Printf("err: %s\n", err.Error())
	}
}
