package github

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/Alexendoo/Slidebot/config"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
)

type Handler struct {
	Discord *discordgo.Session
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch event := event.(type) {
	case *github.PushEvent:
		h.handlePushEvent(event)
	case *github.IssuesEvent:
		h.handleIssuesEvent(event, w)
	default:
		log.Printf("unhandled event: %#+v\n", event)
	}
}

func (h *Handler) handlePushEvent(event *github.PushEvent) {
	log.Printf("event: %#+v\n", event)
}

func createTemplate(name, text string) *template.Template {
	return template.Must(template.New(name).Parse(text))
}
func executeTemplate(tpl *template.Template, data interface{}) string {
	var buf bytes.Buffer
	err := tpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

var issueOpen = createTemplate("issueOpen", "{{.Sender.Login}} opened issue {{.Issue.Title}}")

func (h *Handler) handleIssuesEvent(event *github.IssuesEvent, w http.ResponseWriter) {
	target, ok := config.Repos[*event.Repo.FullName]
	if !ok {
		log.Printf("unknown event.Repo.FullName: %s\n", *event.Repo.FullName)
		return
	}

	var tpl *template.Template
	switch *event.Action {
	case "opened":
		tpl = issueOpen
	}

	message := executeTemplate(tpl, event)

	_, err := h.Discord.ChannelMessageSend(target, message)

	if err != nil {
		log.Printf("err: %#+v\n", err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
	}
}
