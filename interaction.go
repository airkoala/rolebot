package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func buildComponents(s *discordgo.Session, guildId string, guildConfig *GuildConfig) ([]discordgo.MessageComponent, error) {
	components := make([]discordgo.MessageComponent, 0, len(guildConfig.RoleGroups))
	for _, rg := range guildConfig.RoleGroups {
		c, err := rg.toComponents(s, guildId)
		if err != nil {
			return components, err
		}

		components = append(components, c...)
	}

	return components, nil
}

func (self *RoleGroup) toComponents(s *discordgo.Session, guildId string) ([]discordgo.MessageComponent, error) {
	minValues := 0
	maxValues := 1
	// if self.Multiple {
	// 	// Discord only supports 25 max for dropdown interacts
	// 	maxValues = min(len(self.Roles), 25)
	// }

	roles, err := s.GuildRoles(guildId)
	if err != nil {
		log.Printf("Failed to fetch roles for guild %v.\n", guildId)
		return []discordgo.MessageComponent{}, err
	}

	options := make([]discordgo.SelectMenuOption, len(self.Roles))
	for i, rId := range self.Roles {
		rName := "<unknown role>"
		for _, r := range roles {
			if r.ID == rId {
				rName = r.Name
				break
			}
		}

		options[i] = discordgo.SelectMenuOption{
			Label: rName,
			Value: rId,
		}
	}

	content := fmt.Sprintf("# %s\n%s", self.Heading, self.Description)
	if self.Multiple {
		content += "\n-# (you can have more than one but you have to toggle one at a time. (i was too lazy to code it any better))"
	} else {
		content += "\n-# (only one allowed)"
	}

	return []discordgo.MessageComponent{
		discordgo.TextDisplay{Content: content},
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{discordgo.SelectMenu{
			MenuType:  discordgo.StringSelectMenu,
			CustomID:  "setRoles " + strconv.Itoa(self.Id),
			MinValues: &minValues,
			MaxValues: maxValues,
			Options:   options,
		}}},
	}, nil
}

type RoleSet map[string]bool

func setRoles(s *discordgo.Session, member *discordgo.Member, guildId string, rg *RoleGroup, roleIds []string) ([]string, []string, error) {
	if len(roleIds) == 0 {
		// no action
		return []string{}, []string{}, nil
	}

	rolesAdded := make([]string, 0)
	rolesRemoved := make([]string, 0)
	roleSet := make(RoleSet)
	for _, r := range member.Roles {
		roleSet[r] = true
	}
	log.Printf("roleSet: %v, rg: %v\n", roleSet, rg)

	if rg.Multiple {
		for _, r := range roleIds {
			// Toggle each passed in role

			if roleSet[r] {
				rolesAdded = append(rolesAdded, r)
			} else {
				rolesRemoved = append(rolesRemoved, r)
			}

			roleSet[r] = !roleSet[r]
		}
	} else {
		for _, r := range rg.Roles {
			// Only toggle the selected role. Disable everything else in the group.
			if r == roleIds[0] {
				if roleSet[r] {
					rolesRemoved = append(rolesRemoved, r)
				} else {
					rolesAdded = append(rolesAdded, r)
				}
				roleSet[r] = !roleSet[r]
			} else {
				if roleSet[r] {
					rolesRemoved = append(rolesRemoved, r)
					roleSet[r] = false
				}
			}
		}
	}

	newRoles := make([]string, 0)
	for r, enabled := range roleSet {
		if enabled {
			newRoles = append(newRoles, r)
		}
	}

	log.Printf("member: %+v, newRoles: %v\n", member, newRoles)
	data := discordgo.GuildMemberParams{
		Roles: &newRoles,
	}
	_, err := s.GuildMemberEdit(guildId, member.User.ID, &data)
	log.Printf("member.GuildID: %v, member.User.ID: %v\n", member.GuildID, member.User.ID)

	return rolesAdded, rolesRemoved, err
}
