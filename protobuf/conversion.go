package sandwich

import (
	discord "github.com/WelcomerTeam/Sandwich-Daemon/discord/structs"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/xerrors"
	"google.golang.org/protobuf/encoding/protojson"
)

// discord.* -> gRPC Converters.

// Converts discord.User to gRPC counterpart.
func UserToGRPC(user *discord.User) (sandwichUser *User, err error) {
	userJSON, err := jsoniter.Marshal(user)
	if err != nil {
		return
	}

	sandwichUser = &User{}

	err = protojson.Unmarshal(userJSON, sandwichUser)
	if err != nil {
		return sandwichUser, xerrors.Errorf("Failed to unmarshal user: %v", err)
	}

	return
}

// Converts discord.Guild to gRPC counterpart.
func GuildToGRPC(guild *discord.Guild) (sandwichGuild *Guild, err error) {
	guildJSON, err := jsoniter.Marshal(guild)
	if err != nil {
		return
	}

	sandwichGuild = &Guild{}

	err = protojson.Unmarshal(guildJSON, sandwichGuild)
	if err != nil {
		return sandwichGuild, xerrors.Errorf("Failed to unmarshal guild: %v", err)
	}

	return
}

// Converts discord.Channel to gRPC counterpart.
func ChannelToGRPC(channel *discord.Channel) (sandwichChannel *Channel, err error) {
	channelJSON, err := jsoniter.Marshal(channel)
	if err != nil {
		return
	}

	sandwichChannel = &Channel{}

	err = protojson.Unmarshal(channelJSON, sandwichChannel)
	if err != nil {
		return sandwichChannel, xerrors.Errorf("Failed to unmarshal channel: %v", err)
	}

	return
}

// Converts discord.Emoji to gRPC counterpart.
func EmojiToGRPC(emoji *discord.Emoji) (sandwichEmoji *Emoji, err error) {
	emojiJSON, err := jsoniter.Marshal(emoji)
	if err != nil {
		return
	}

	sandwichEmoji = &Emoji{}

	err = protojson.Unmarshal(emojiJSON, sandwichEmoji)
	if err != nil {
		return sandwichEmoji, xerrors.Errorf("Failed to unmarshal emoji: %v", err)
	}

	return
}

// Converts discord.GuildMember to gRPC counterpart.
func GuildMemberToGRPC(guildMember *discord.GuildMember) (sandwichGuildMember *GuildMember, err error) {
	guildMemberJSON, err := jsoniter.Marshal(guildMember)
	if err != nil {
		return
	}

	sandwichGuildMember = &GuildMember{}

	err = protojson.Unmarshal(guildMemberJSON, sandwichGuildMember)
	if err != nil {
		return sandwichGuildMember, xerrors.Errorf("Failed to unmarshal guild member: %v", err)
	}

	return
}

// Converts discord.Role to gRPC counterpart.
func RoleToGRPC(role *discord.Role) (sandwichRole *Role, err error) {
	guildRoleJSON, err := jsoniter.Marshal(role)
	if err != nil {
		return
	}

	sandwichRole = &Role{}

	err = protojson.Unmarshal(guildRoleJSON, sandwichRole)
	if err != nil {
		return sandwichRole, xerrors.Errorf("Failed to unmarshal guild role: %v", err)
	}

	return
}

// gRPC -> discord.* Converters.

// Converts gRPC to discord.User counterpart.
func GRPCToUser(sandwichUser *User) (user *discord.User, err error) {
	userJSON, err := protojson.Marshal(sandwichUser)
	if err != nil {
		return
	}

	user = &discord.User{}

	err = jsoniter.Unmarshal(userJSON, user)
	if err != nil {
		return user, xerrors.Errorf("Failed to unmarshal user: %v", err)
	}

	return
}

// Converts gRPC to discord.Guild counterpart.
func GRPCToGuild(sandwichGuild *Guild) (guild *discord.Guild, err error) {
	guildJSON, err := protojson.Marshal(sandwichGuild)
	if err != nil {
		return
	}

	guild = &discord.Guild{}

	err = jsoniter.Unmarshal(guildJSON, guild)
	if err != nil {
		return guild, xerrors.Errorf("Failed to unmarshal guild: %v", err)
	}

	return
}

// Converts gRPC to discord.Channel counterpart.
func GRPCToChannel(sandwichChannel *Channel) (channel *discord.Channel, err error) {
	channelJSON, err := protojson.Marshal(sandwichChannel)
	if err != nil {
		return
	}

	channel = &discord.Channel{}

	err = jsoniter.Unmarshal(channelJSON, channel)
	if err != nil {
		return channel, xerrors.Errorf("Failed to unmarshal channel: %v", err)
	}

	return
}

// Converts gRPC to discord.Emoji counterpart.
func GRPCToEmoji(sandwichEmoji *Emoji) (emoji *discord.Emoji, err error) {
	emojiJSON, err := protojson.Marshal(sandwichEmoji)
	if err != nil {
		return
	}

	emoji = &discord.Emoji{}

	err = jsoniter.Unmarshal(emojiJSON, emoji)
	if err != nil {
		return emoji, xerrors.Errorf("Failed to unmarshal emoji: %v", err)
	}

	return
}

// Converts gRPC to discord.GuildMember counterpart.
func GRPCToGuildMember(sandwichGuildMember *GuildMember) (guildMember *discord.GuildMember, err error) {
	guildMemberJSON, err := protojson.Marshal(sandwichGuildMember)
	if err != nil {
		return
	}

	guildMember = &discord.GuildMember{}

	err = jsoniter.Unmarshal(guildMemberJSON, guildMember)
	if err != nil {
		return guildMember, xerrors.Errorf("Failed to unmarshal guild member: %v", err)
	}

	return
}

// Converts gRPC to discord.Role counterpart.
func GRPCToRole(sandwichRole *Role) (role *discord.Role, err error) {
	guildRoleJSON, err := protojson.Marshal(sandwichRole)
	if err != nil {
		return
	}

	role = &discord.Role{}

	err = jsoniter.Unmarshal(guildRoleJSON, role)
	if err != nil {
		return role, xerrors.Errorf("Failed to unmarshal guild role: %v", err)
	}

	return
}
