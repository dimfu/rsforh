package handlers

import (
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Remove struct{}

func (h *Remove) Name() string {
	return "remove"
}

func (h *Remove) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:                     h.Name(),
		Description:              "Remove online rally channel category",
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

func (h *Remove) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := hasOrganizerRole(s, i); err != nil {
		sendInteraction(err.Error(), s, i, true)
		return
	}

	data := i.ApplicationCommandData()
	if len(data.Options) <= 0 {
		return
	}

	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		sendError(err, s, i)
		return
	}

	var parentID string
	toBeRemoved := []string{}
	rallyID := data.Options[0].StringValue()
	for _, channel := range channels {
		if channel.Topic == fmt.Sprintf("RALLYID_%s", rallyID) {
			parentID = channel.ParentID
			toBeRemoved = append(toBeRemoved, channel.ID)
		}
	}

	parentChannel, err := s.Channel(parentID)
	if err != nil {
		sendError(err, s, i)
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Removing channel category %s", parentChannel.Name),
			Flags:   1 << 6, // make it ephemeral (only visible to the user)
		},
	})
	if err != nil {
		log.Printf("could not send initial response: %v", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(toBeRemoved))

	for _, id := range toBeRemoved {
		go func(id string) {
			defer wg.Done()
			_, err := s.ChannelDelete(id)
			if err != nil {
				return
			}
		}(id)
	}
	wg.Wait()
	_, err = s.ChannelDelete(parentID)
	if err != nil {
		return
	}
	editInteractionResponse(s, i, fmt.Sprintf("Channel category %q removed", parentChannel.Name))
}
