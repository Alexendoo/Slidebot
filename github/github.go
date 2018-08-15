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
	GetRepo() *github.Repository
	GetSender() *github.User
}

type ActionEvent interface {
	GetAction() string
	GenericEvent
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
	// TODO: add PushEvent and alter behaviour based on tag/commits/etc
	case ActionEvent:
		log.Println("Action Event", webhookType, event.GetAction())
		h.handleGeneric(event, webhookType+"_"+event.GetAction())
	case GenericEvent:
		log.Println("Generic Event", webhookType)
		h.handleGeneric(event, webhookType)
	default:
		log.Println("Unhandled Event", webhookType)
	}
}

func (h *Handler) handleGeneric(event GenericEvent, tag string) {
	fullname := *event.GetRepo().FullName
	target, ok := config.Repos[fullname]
	if !ok {
		log.Printf("Unhandled repo: %s\n", fullname)
		return
	}

	sender := event.GetSender()

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    *sender.Login,
			URL:     *sender.HTMLURL,
			IconURL: *sender.AvatarURL,
		},

		Title:       templates.Exec(event, tag, "title"),
		URL:         templates.Exec(event, tag, "URL"),
		Description: templates.Exec(event, tag, "description"),
	}

	if footer := templates.Exec(event, tag, "footer"); footer != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: footer,
		}
	}

	_, err := h.Discord.ChannelMessageSendEmbed(target, embed)

	if err != nil {
		log.Printf("Error sending generic message event: %s\n", err.Error())
	}
}
