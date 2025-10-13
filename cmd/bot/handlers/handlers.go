package handlers

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/rsforh/cmd/bot/command"
)

var defaultPermission = int64(discordgo.PermissionAdministrator)

var (
	ERR_INTERNAL_ERROR = errors.New("Something went wrong while executing this instruction")
)

var Handlers = []command.Command{
	&Role{},
}

func sendInteraction(r string, s *discordgo.Session, i *discordgo.InteractionCreate, ephemeral bool) {
	log.Printf("[sendInteraction] %s", r)
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: r,
		},
	}
	if ephemeral {
		response.Data.Flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, response)
}

func sendError(err error, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println(err)
	sendInteraction(ERR_INTERNAL_ERROR.Error(), s, i, true)
}

// Handler to create Rally Organizer role upon joining server
func InitOrganizerRole(s *discordgo.Session, guildId string) error {
	st, err := s.GuildRoles(guildId)

	if err != nil {
		return err
	}

	for _, role := range st {
		if role.Name == "Rally Organizer" {
			return nil
		}
	}

	_, err = s.GuildRoleCreate(
		guildId,
		&discordgo.RoleParams{
			Name: "Rally Organizer",
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}
