package main

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func getWizardComponents(s *discordgo.Session, guildId string, guildConfig *GuildConfig, member *discordgo.Member) ([]discordgo.MessageComponent, error) {
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

func setRoles(s *discordgo.Session, member *discordgo.Member, guildId string, rg *RoleGroup, selectedRoleIds []string) ([]string, []string, error) {
	rolesAdded := make([]string, 0)
	rolesRemoved := make([]string, 0)
	newRolesSet := make(RoleSet)
	for _, r := range member.Roles {
		newRolesSet[r] = true
	}

	for _, rid := range rg.Roles {
		isSelected := slices.Contains(selectedRoleIds, rid)

		if newRolesSet[rid] && !isSelected {
			rolesRemoved = append(rolesRemoved, rid)
			newRolesSet[rid] = isSelected
		} else if !newRolesSet[rid] && isSelected {
			rolesAdded = append(rolesAdded, rid)
			newRolesSet[rid] = isSelected
		}
	}

	newRoles := make([]string, 0)
	for r, enabled := range newRolesSet {
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
