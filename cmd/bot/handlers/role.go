package handlers

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Role struct{}

func (h *Role) Name() string {
	return "role"
}

func (h *Role) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:                     h.Name(),
		Description:              "Manage organizer role",
		DefaultMemberPermissions: &defaultPermission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "Add organizer role to user",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "target",
						Description: "Target User",
						Required:    true,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove organizer role from user",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "target",
						Description: "Target User",
						Required:    true,
					},
				},
			},
		},
	}
}

func (h *Role) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	if len(data.Options) <= 0 {
		log.Println("empty slash command options given")
		return
	}

	cmd := i.ApplicationCommandData().Options
	subcmd := data.Options[0]
	var st *discordgo.User

	if len(subcmd.Options) > 0 {
		payload := subcmd.Options[0].StringValue()
		usrid := payload[2 : len(payload)-1]
		u, err := s.User(usrid)
		if err != nil {
			sendInteraction("Cannot add invalid user, use @user to properly target user", s, i, true)
		}
		st = u
	}

	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		log.Println(err.Error())
		return
	}

	var tm *discordgo.Role
	for _, role := range roles {
		if role.Name == "Rally Organizer" {
			tm = role
		}
	}

	if tm == nil {
		sendInteraction("Can't find Rally Organizer role", s, i, false)
		return
	}

	ret := ""

	switch cmd[0].Name {
	case "add":
		if err = s.GuildMemberRoleAdd(i.GuildID, st.ID, tm.ID); err != nil {
			break
		}
		ret = fmt.Sprintf("<@%s> is now rally organizer", st.ID)
	case "remove":
		if err = s.GuildMemberRoleRemove(i.GuildID, st.ID, tm.ID); err != nil {
			break
		}
		ret = fmt.Sprintf("<@%s> is no longer rally organizer", st.ID)
	default:
	}

	if err != nil {
		log.Println(err)
		sendError(ERR_INTERNAL_ERROR, s, i)
		return
	}

	sendInteraction(ret, s, i, false)
}
