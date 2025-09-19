package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	configDir := os.Getenv("CONFIG_DIR")
	if len(configDir) == 0 {
		configDir = "/etc/rolebot"
	}

	config, err := readConfig(configDir)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Config: %+v\n", config)

	dg, err := discordgo.New("Bot " + config.token)
	if err != nil {
		panic(err)
	}

	// No perms except for admin
	perm := int64(0)

	for gId := range config.guilds {
		_, err = dg.ApplicationCommandCreate(config.appId, gId, &discordgo.ApplicationCommand{
			Name:                     "roles",
			Description:              "(Re)create the role selection menu.",
			DefaultMemberPermissions: &perm,
		})
		if err != nil {
			fmt.Printf("Cannot create slash command: %v", err)
			panic(err)
		}
	}

	dg.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		fmt.Println("Connected.")

		fmt.Println("-----------------------")
		fmt.Println("Configured guilds: ")
		for gId, gCfg := range config.guilds {
			guild, err := s.Guild(gId)
			if err != nil {
				fmt.Printf("Failed to get guild %v: %v\n", gId, err)
				continue
			}

			managedRoles := make(map[string]bool)
			for _, rg := range gCfg.RoleGroups {
				for _, rId := range rg.Roles {
					managedRoles[rId] = true
				}
			}

			fmt.Printf("%v (%v)\n", guild.Name, gId)
			fmt.Println("Available roles:")
			for _, r := range guild.Roles {
				fmt.Printf("\"%v\",  # @%v", r.ID, r.Name)
				if managedRoles[r.ID] {
					fmt.Print(" (managed)")
				}
				fmt.Print("\n")
			}
			fmt.Println()
		}
		fmt.Println("-----------------------")
	})

	// On receiving interaction
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "roles" {
				fmt.Printf("/roles executed in guild %v by %v.\n", i.GuildID, i.Member.User.ID)

				gCfg, found := config.guilds[i.GuildID]
				if !found {
					err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Flags:   discordgo.MessageFlagsEphemeral,
							Content: "No config found for this server.",
						},
					})
					if err != nil {
						fmt.Println("Interaction response failed:", err)
						return
					}
				}

				res := buildRolesList(&gCfg)
				err = s.InteractionRespond(i.Interaction, &res)

				if err != nil {
					fmt.Println("Interaction response failed:", err)
				}
			}
		case discordgo.InteractionMessageComponent:
			switch strings.Split(i.MessageComponentData().CustomID, " ")[0] {
			case "promptWizard":
				fmt.Printf("promptWizard called by %v in %v.\n", i.Member.User.Username, i.GuildID)
				gcfg := config.guilds[i.GuildID]
				components, err := getWizardComponents(s, i.GuildID, &gcfg, i.Member)
				if err != nil {
					fmt.Printf("Failed to build wizard components: %v\n", err)
					return
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Flags:      discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsIsComponentsV2,
						Components: components,
					},
				})

				if err != nil {
					fmt.Printf("Interaction response failed: %v", err)
				}
			case "setRoles":
				rgId, err := strconv.Atoi(strings.TrimPrefix(i.MessageComponentData().CustomID, "setRoles "))
				if err != nil {
					fmt.Printf("Invalid roleGroup ID: %v\n", rgId)
					return
				}

				roleIds := i.MessageComponentData().Values

				fmt.Printf("Setting roles for %v in %v\n", i.Member.User.ID, i.GuildID)
				fmt.Printf("roleIds: %v, rgId: %v\n", roleIds, rgId)

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
					fmt.Printf("Error setting roles: %v\n", err)
					responseMsg = "Error setting roles."
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: responseMsg,
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})

				if err != nil {
					fmt.Println(err)
				}
			default:
				return // should never be reached though
			}
		}
	})

	dg.Identify.Intents = discordgo.IntentsAll

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println("Started.")

	// Set channel to send signals to
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Await signal
	<-sc

	dg.Close()
}
