package handlers

import (
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/rsforh/cmd/bot/command"
)

var (
	roleOrganizerName = "Rally Organizer"
	defaultPermission = int64(discordgo.PermissionAdministrator)
)

var (
	ERR_INTERNAL_ERROR = errors.New("Something went wrong while executing this instruction")
)

var Handlers = []command.Command{
	&Role{},
	&Start{},
	&Remove{},
}

func hasOrganizerRole(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	user := i.Member

	// skips check if user has admin access
	if user.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		return nil
	}

	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		return err
	}

	var r *discordgo.Role
	for _, role := range roles {
		if role.Name == roleOrganizerName {
			r = role
		}
	}

	if r == nil {
		return errors.New("Can't find tournament role")
	}

	if slices.Contains(user.Roles, r.ID) {
		return nil
	}

	return errors.New("Insufficent permission to use this command")
}

func sendInteraction(r string, s *discordgo.Session, i *discordgo.InteractionCreate, ephemeral bool) {
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

func editInteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Printf("could not edit interaction response: %v", err)
	}
}

// Handler to create Rally Organizer role upon joining server
func InitOrganizerRole(s *discordgo.Session, guildId string) error {
	st, err := s.GuildRoles(guildId)

	if err != nil {
		return err
	}

	for _, role := range st {
		if role.Name == roleOrganizerName {
			return nil
		}
	}

	_, err = s.GuildRoleCreate(
		guildId,
		&discordgo.RoleParams{
			Name: roleOrganizerName,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}
