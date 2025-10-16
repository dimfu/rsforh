package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
)

type rallyInfo struct {
	id   string
	name string
	legs [][]string
}

type Start struct{}

func (h *Start) Name() string {
	return "start"
}

func (h *Start) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:                     h.Name(),
		Description:              "Start online rally",
		DefaultMemberPermissions: &defaultPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "rallyid",
				Description: "Online rally id",
				Required:    true,
			},
		},
	}
}

func (h *Start) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := hasOrganizerRole(s, i); err != nil {
		sendInteraction(err.Error(), s, i, true)
		return
	}

	data := i.ApplicationCommandData()
	if len(data.Options) <= 0 {
		return
	}

	rallyInfo := rallyInfo{
		id: data.Options[0].StringValue(),
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Creating channels for rally... please wait",
			Flags:   1 << 6, // make it ephemeral (only visible to the user)
		},
	})
	if err != nil {
		log.Printf("could not send initial response: %v", err)
		return
	}

	if err := h.collectRallyInfo(&rallyInfo); err != nil {
		log.Printf("could not fetch online rally page with id %s: \n%v", rallyInfo.id, err)
		sendError(ERR_INTERNAL_ERROR, s, i)
		return
	}

	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		log.Printf("could not fetch online rally page with id %s: \n%v", rallyInfo.id, err)
		editInteractionResponse(s, i, "Failed to fetch channels.")
		return
	}
	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildCategory && channel.Name == rallyInfo.name {
			editInteractionResponse(s, i, "Channel category for this rally already exists.")
			return
		}
	}

	cat, err := s.GuildChannelCreate(i.GuildID, rallyInfo.name, discordgo.ChannelTypeGuildCategory)
	if err != nil {
		log.Printf("could not create online rally category %q: \n%v", rallyInfo.name, err)
		editInteractionResponse(s, i, "Failed to create rally category.")
		return
	}

	for leg, stages := range rallyInfo.legs {
		for _, stage := range stages {
			// +1 because stages is zero indexed
			name := fmt.Sprintf("leg-%d-%s", leg+1, stage)
			_, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
				Name:     name,
				Type:     discordgo.ChannelTypeGuildText,
				ParentID: cat.ID,
				Topic:    fmt.Sprintf("RALLYID_%s", rallyInfo.id),
			})
			if err != nil {
				log.Printf("could not create channel %q: \n%v", name, err)
				editInteractionResponse(s, i, fmt.Sprintf("Failed to create channel %q.", name))
				return
			}
		}
	}

	editInteractionResponse(s, i, fmt.Sprintf("Created category **%s** for rally ID `%s`", rallyInfo.name, rallyInfo.id))
}

func (h *Start) collectRallyInfo(r *rallyInfo) error {
	var leg int
	c := colly.NewCollector()

	c.OnHTML("tr.fejlec", func(e *colly.HTMLElement) {
		r.name = strings.TrimSpace(e.Text)
	})

	c.OnHTML("td.szdb table:nth-of-type(2) tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			text := strings.TrimSpace(tr.Text)
			if strings.Contains(strings.ToLower(text), "leg") {
				leg++
				r.legs = append(r.legs, []string{})
				return
			}
			className := tr.Attr("class")
			if className == "paratlan" || className == "paros" {
				stage := strings.TrimSpace(tr.ChildText("td:nth-of-type(2)"))
				r.legs[leg-1] = append(r.legs[leg-1], stage)
			}
		})
	})

	err := c.Visit(h.onlineRallyUrl(r.id))
	if err != nil {
		return err
	}
	return nil
}

func (h *Start) onlineRallyUrl(id string) string {
	return fmt.Sprintf("https://rallysimfans.hu/rbr/rally_online.php?centerbox=rally_list_details.php&rally_id=%s", id)
}
