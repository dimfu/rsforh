package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dimfu/rsforh/cmd/bot/handlers"
	"github.com/dimfu/rsforh/internal/config"
)

func main() {
	start := time.Now()
	cfg, err := config.Setup()
	if err != nil {
		fmt.Printf("error initializing config: %v\n", err)
		os.Exit(1)
	}

	var token string
	args := os.Args
	if len(cfg.RBRInstallationPath) > 0 {
		token = cfg.DiscordToken
	} else {
		if len(args) < 2 {
			fmt.Println("please provide discord bot token as the second argument")
			os.Exit(1)
		}
		token = args[1]
	}

	if len(cfg.DiscordToken) == 0 {
		cfg.Set(config.DiscordToken, token)
	}

	_, err = cfg.Get(config.DiscordToken)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	close := make(chan os.Signal, 1)
	signal.Notify(close, syscall.SIGINT, syscall.SIGTERM)

	discord, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection with discord: %v", err)
	}
	defer discord.Close()

	lat := time.Since(start)
	log.Printf("bot is up and running took %s", lat.String())

	appId := discord.State.User.ID

	registeredCommands := make(map[string]struct{})

	// Register all commands
	for _, handler := range handlers.Handlers {
		_, err = discord.ApplicationCommandCreate(appId, "", handler.Command())
		if err != nil {
			log.Fatalf("error while registering command %q: %v", handler.Name(), err)
		}
		log.Printf("command %q registered and updated", handler.Name())
		registeredCommands[handler.Name()] = struct{}{}
	}

	// Remove unused commands from slash command list
	cmds, err := discord.ApplicationCommands(appId, "")
	if err != nil {
		log.Printf("could not get registered commands: %v", err)
	}
	for _, cmd := range cmds {
		_, ok := registeredCommands[cmd.Name]
		if !ok {
			if err := discord.ApplicationCommandDelete(appId, "", cmd.ID); err != nil {
				log.Printf("could not delete unused commands: %v", err)
				continue
			}
			log.Printf("found %q command unused, deleting now", cmd.Name)
		}
	}

	// Initialize rally organizer role
	discord.AddHandler(func(s *discordgo.Session, gc *discordgo.GuildCreate) {
		if err := handlers.InitOrganizerRole(s, gc.ID); err != nil {
			log.Printf("failed to create Rally Organizer role:\n%v", err)
		}
	})

	// Listens to which command is being executed
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		for _, handler := range handlers.Handlers {
			if handler.Name() == i.ApplicationCommandData().Name {
				handler.Handler(discord, i)
				return
			}
		}
	})

	<-close
	err = discord.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
	log.Println("connection closed")
}
