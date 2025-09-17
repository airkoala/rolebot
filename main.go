package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	config, err := readConfig("config")
	if err != nil {
		panic(err)
	}

	log.Printf("Config: %+v\n", config)

	dg, err := discordgo.New("Bot " + config.token)
	if err != nil {
		log.Panic(err)
	}

	// No perms except for admin
	perm := int64(0)

	_, err = dg.ApplicationCommandCreate(config.appId, "593359698800017418", &discordgo.ApplicationCommand{
		Name:                     "roles",
		Description:              "(Re)create the role selection menu.",
		DefaultMemberPermissions: &perm,
	})
	if err != nil {
		log.Panicf("Cannot create slash command: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		log.Println("Connected.")
	})

	// On receiving interaction
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			log.Println("InteractionApplicationCommand")

			if i.ApplicationCommandData().Name == "roles" {
				gcfg := config.guilds[i.GuildID]
				components, err := buildComponents(s, i.GuildID, &gcfg)
				if err != nil {
					log.Printf("Failed to build components: %v\n", err)
					return
				}

				res := discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags: discordgo.MessageFlagsIsComponentsV2,
						// Content: "asdf",
						Components: components,
						// Components: components,
					},
				}

				err = s.InteractionRespond(i.Interaction, &res)

				if err != nil {
					log.Println(err)
				}
			}
		case discordgo.InteractionMessageComponent:
			if !strings.HasPrefix(i.MessageComponentData().CustomID, "setRoles ") {
				// all selection dropdowns have customId in the form `setRoles <rgId>`
				return
			}

			rgId, err := strconv.Atoi(strings.TrimPrefix(i.MessageComponentData().CustomID, "setRoles "))
			if err != nil {
				log.Printf("Invalid roleGroup ID: %v\n", rgId)
				return
			}

			roleIds := i.MessageComponentData().Values

			log.Printf("Setting roles for %v in %v\n", i.Member.User.ID, i.GuildID)
			log.Printf("roleIds: %v, rgId: %v\n", roleIds, rgId)

			rg := &config.guilds[i.GuildID].RoleGroups[rgId]

			added, removed, err := setRoles(s, i.Member, i.GuildID, rg, roleIds)

			responseMsg := ""
			if len(added) > 0 {
				responseMsg += "Roles added:\n"
				for _, rId := range added {
					responseMsg += "* <@&" + rId + ">\n"
				}
			}

			if len(removed) > 0 {
				responseMsg += "Roles removed:\n"
				for _, rId := range removed {
					responseMsg += "* <@&" + rId + ">\n"
				}
			}

			if len(responseMsg) == 0 {
				responseMsg = "No change."
			}

			if err != nil {
				log.Printf("Error setting roles: %v\n", err)
				responseMsg = "Error setting roles."
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: responseMsg,
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})

			if err != nil {
				log.Println(err)
			}
		}
	})

	dg.Identify.Intents = discordgo.IntentsAll

	err = dg.Open()
	if err != nil {
		log.Panic(err)
	}

	log.Println("Started.")

	// Set channel to send signals to
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Await signal
	<-sc

	dg.Close()
}
