package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func buildRolesList(gCfg *GuildConfig) discordgo.InteractionResponse {
	content := "# Available Roles:\n"

	for _, rolegroup := range gCfg.RoleGroups {
		content += "## " + rolegroup.Heading + "\n"
		content += "> " + rolegroup.Description + "\n"

		for _, rId := range rolegroup.Roles {
			content += fmt.Sprintf("<@&%v> ", rId)
		}

		content += "\n"
	}

	components := []discordgo.MessageComponent{
		discordgo.TextDisplay{
			Content: content,
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Pick roles",
					Style:    discordgo.PrimaryButton,
					CustomID: "promptWizard",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🧙‍♀️",
					},
				},
			},
		},
	}

	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:           discordgo.MessageFlagsIsComponentsV2,
			AllowedMentions: &discordgo.MessageAllowedMentions{}, // dont ping anyone
			Components:      components,
		},
	}
}
