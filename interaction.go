package main

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func buildComponents(s *discordgo.Session, guildId string, guildConfig *GuildConfig, member *discordgo.Member) ([]discordgo.MessageComponent, error) {
	components := make([]discordgo.MessageComponent, 0, len(guildConfig.RoleGroups))
	components = append(components, discordgo.TextDisplay{Content: "# Pick your roles"})
	for _, rg := range guildConfig.RoleGroups {
		c, err := rg.toComponents(s, guildId, member)
		if err != nil {
			return components, err
		}

		components = append(components, c...)
	}

	return components, nil
}

type RoleSet map[string]bool

func (self *RoleGroup) toComponents(s *discordgo.Session, guildId string, member *discordgo.Member) ([]discordgo.MessageComponent, error) {

	roleset := make(RoleSet)
	for _, r := range member.Roles {
		roleset[r] = true
	}

	minValues := 0
	maxValues := 1
	if self.Multiple {
		// Discord only supports 25 max for dropdown interacts
		maxValues = min(len(self.Roles), 25)
	}

	roles, err := s.GuildRoles(guildId)
	if err != nil {
		fmt.Printf("Failed to fetch roles for guild %v.\n", guildId)
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

		// fill in picked roles
		set := false
		if roleset[rId] {
			set = true
		}

		options[i] = discordgo.SelectMenuOption{
			Label:   rName,
			Value:   rId,
			Default: set,
		}
	}

	content := fmt.Sprintf("## %s\n%s", self.Heading, self.Description)
	if self.Multiple {
		content += "\n-# (you can have more than one)"
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

func setRoles(s *discordgo.Session, member *discordgo.Member, guildId string, rg *RoleGroup, roleIds []string) ([]string, []string, error) {
	rolesAdded := make([]string, 0)
	rolesRemoved := make([]string, 0)
	roleSet := make(RoleSet)
	for _, r := range member.Roles {
		roleSet[r] = true
	}

	if rg.Multiple {
		for _, rid := range rg.Roles {
			selected := slices.Contains(roleIds, rid)

			if roleSet[rid] && !selected {
				rolesRemoved = append(rolesRemoved, rid)
				roleSet[rid] = false // TODO: perhaps roleSet needs to be treated differently, repeating this is a bit ugly
			} else if !roleSet[rid] && selected {
				rolesAdded = append(rolesAdded, rid)
				roleSet[rid] = true
			}
		}
	} else {
		if len(roleIds) < 1 {
			for _, rid := range rg.Roles {
				if roleSet[rid] {
					rolesRemoved = append(rolesRemoved, rid)
					roleSet[rid] = false
				}
			}

		} else {
			selected := roleIds[0]
			for _, rid := range rg.Roles {
				if roleSet[rid] && !(selected == rid) {
					rolesRemoved = append(rolesRemoved, rid)
					roleSet[rid] = false
				} else if !roleSet[rid] && selected == rid {
					rolesAdded = append(rolesAdded, rid)
					roleSet[rid] = true
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

	data := discordgo.GuildMemberParams{
		Roles: &newRoles,
	}
	_, err := s.GuildMemberEdit(guildId, member.User.ID, &data)

	fmt.Printf("Roles changed for @%v (%v):", member.User.Username, member.User.ID)
	fmt.Printf("rolesAdded: %v, rolesRemoved: %v\n", rolesAdded, rolesRemoved)

	return rolesAdded, rolesRemoved, err
}
