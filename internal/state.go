package internal

import (
	"sync"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	sandwich_structs "github.com/WelcomerTeam/Sandwich-Daemon/structs"
)

//
// Guild Operations
//

// GuildFromState converts the structs.StateGuild into a discord.Guild, for use within the application.
// Channels, Roles, Members and Emoji lists will not be populated.
func (ss *SandwichState) GuildFromState(guildState *sandwich_structs.StateGuild) (guild *discord.Guild) {
	guild = &discord.Guild{
		ID:              guildState.ID,
		Name:            guildState.Name,
		Icon:            guildState.Icon,
		IconHash:        guildState.IconHash,
		Splash:          guildState.Splash,
		DiscoverySplash: guildState.DiscoverySplash,

		Owner:       guildState.Owner,
		OwnerID:     guildState.OwnerID,
		Permissions: guildState.Permissions,
		Region:      guildState.Region,

		AFKChannelID: guildState.AFKChannelID,
		AFKTimeout:   guildState.AFKTimeout,

		WidgetEnabled:   guildState.WidgetEnabled,
		WidgetChannelID: guildState.WidgetChannelID,

		VerificationLevel:           guildState.VerificationLevel,
		DefaultMessageNotifications: guildState.DefaultMessageNotifications,
		ExplicitContentFilter:       guildState.ExplicitContentFilter,

		MFALevel:           guildState.MFALevel,
		ApplicationID:      guildState.ApplicationID,
		SystemChannelID:    guildState.SystemChannelID,
		SystemChannelFlags: guildState.SystemChannelFlags,
		RulesChannelID:     guildState.RulesChannelID,

		JoinedAt:    guildState.JoinedAt,
		Large:       guildState.Large,
		Unavailable: guildState.Unavailable,
		MemberCount: guildState.MemberCount,

		MaxPresences:  guildState.MaxPresences,
		MaxMembers:    guildState.MaxMembers,
		VanityURLCode: guildState.VanityURLCode,
		Description:   guildState.Description,
		Banner:        guildState.Banner,

		PremiumTier:               guildState.PremiumTier,
		PremiumSubscriptionCount:  guildState.PremiumSubscriptionCount,
		PreferredLocale:           guildState.PreferredLocale,
		PublicUpdatesChannelID:    guildState.PublicUpdatesChannelID,
		MaxVideoChannelUsers:      guildState.MaxVideoChannelUsers,
		ApproximateMemberCount:    guildState.ApproximateMemberCount,
		ApproximatePresenceCount:  guildState.ApproximatePresenceCount,
		NSFWLevel:                 guildState.NSFWLevel,
		PremiumProgressBarEnabled: guildState.PremiumProgressBarEnabled,

		Features:             guildState.Features,
		StageInstances:       make([]*discord.StageInstance, 0, len(guildState.StageInstances)),
		Stickers:             make([]*discord.Sticker, 0, len(guildState.Stickers)),
		GuildScheduledEvents: make([]*discord.ScheduledEvent, 0, len(guildState.GuildScheduledEvents)),
	}

	for _, stageInstance := range guildState.StageInstances {
		stageInstance := stageInstance
		guild.StageInstances = append(guild.StageInstances, &stageInstance)
	}

	for _, sticker := range guildState.Stickers {
		sticker := sticker
		guild.Stickers = append(guild.Stickers, &sticker)
	}

	for _, scheduledEvent := range guildState.GuildScheduledEvents {
		scheduledEvent := scheduledEvent
		guild.GuildScheduledEvents = append(guild.GuildScheduledEvents, &scheduledEvent)
	}

	return guild
}

// GuildFromState converts from discord.Guild to structs.StateGuild, for storing in cache.
// Does not add Channels, Roles, Members and Emojis to state.
func (ss *SandwichState) GuildToState(guild *discord.Guild) (guildState *sandwich_structs.StateGuild) {
	guildState = &sandwich_structs.StateGuild{
		ID:              guild.ID,
		Name:            guild.Name,
		Icon:            guild.Icon,
		IconHash:        guild.IconHash,
		Splash:          guild.Splash,
		DiscoverySplash: guild.DiscoverySplash,

		OwnerID:     guild.OwnerID,
		Permissions: guild.Permissions,
		Region:      guild.Region,

		AFKChannelID: guild.AFKChannelID,
		AFKTimeout:   guild.AFKTimeout,

		WidgetEnabled:   guild.WidgetEnabled,
		WidgetChannelID: guild.WidgetChannelID,

		VerificationLevel:           guild.VerificationLevel,
		DefaultMessageNotifications: guild.DefaultMessageNotifications,
		ExplicitContentFilter:       guild.ExplicitContentFilter,

		Features: guild.Features,

		MFALevel:           guild.MFALevel,
		ApplicationID:      guild.ApplicationID,
		SystemChannelID:    guild.SystemChannelID,
		SystemChannelFlags: guild.SystemChannelFlags,
		RulesChannelID:     guild.RulesChannelID,

		JoinedAt:    guild.JoinedAt,
		Large:       guild.Large,
		Unavailable: guild.Unavailable,
		MemberCount: guild.MemberCount,

		MaxPresences:  guild.MaxPresences,
		MaxMembers:    guild.MaxMembers,
		VanityURLCode: guild.VanityURLCode,
		Description:   guild.Description,
		Banner:        guild.Banner,
		PremiumTier:   guild.PremiumTier,

		PremiumSubscriptionCount: guild.PremiumSubscriptionCount,
		PreferredLocale:          guild.PreferredLocale,
		PublicUpdatesChannelID:   guild.PublicUpdatesChannelID,
		MaxVideoChannelUsers:     guild.MaxVideoChannelUsers,
		ApproximateMemberCount:   guild.ApproximateMemberCount,
		ApproximatePresenceCount: guild.ApproximatePresenceCount,

		NSFWLevel:      guild.NSFWLevel,
		StageInstances: make([]discord.StageInstance, 0),
		Stickers:       make([]discord.Sticker, 0),
	}

	for _, stageInstance := range guild.StageInstances {
		guildState.StageInstances = append(guildState.StageInstances, *stageInstance)
	}

	for _, sticker := range guild.Stickers {
		guildState.Stickers = append(guildState.Stickers, *sticker)
	}

	return guildState
}

// GetGuild returns the guild with the same ID from the cache.
// Returns a boolean to signify a match or not.
func (ss *SandwichState) GetGuild(guildID discord.Snowflake) (guild *discord.Guild, ok bool) {
	ss.guildsMu.RLock()
	defer ss.guildsMu.RUnlock()

	stateGuild, ok := ss.Guilds[guildID]
	if !ok {
		return
	}

	guild = ss.GuildFromState(stateGuild)

	return
}

// SetGuild creates or updates a guild entry in the cache.
func (ss *SandwichState) SetGuild(ctx *StateCtx, guild *discord.Guild) {
	ss.guildsMu.Lock()
	defer ss.guildsMu.Unlock()

	ctx.ShardGroup.guildsMu.Lock()
	ctx.ShardGroup.Guilds[guild.ID] = true
	ctx.ShardGroup.guildsMu.Unlock()

	for _, role := range guild.Roles {
		ss.SetGuildRole(ctx, guild.ID, role)
	}

	for _, channel := range guild.Channels {
		ss.SetGuildChannel(ctx, &guild.ID, channel)
	}

	for _, emoji := range guild.Emojis {
		ss.SetGuildEmoji(ctx, guild.ID, emoji)
	}

	if ctx.CacheMembers {
		for _, member := range guild.Members {
			ss.SetGuildMember(ctx, guild.ID, member)
		}
	}

	ss.Guilds[guild.ID] = ss.GuildToState(guild)
}

// RemoveGuild removes a guild from the cache.
func (ss *SandwichState) RemoveGuild(ctx *StateCtx, guildID discord.Snowflake) {
	ss.guildsMu.Lock()
	defer ss.guildsMu.Unlock()

	if !ctx.Stateless {
		ctx.ShardGroup.guildsMu.Lock()
		delete(ctx.ShardGroup.Guilds, guildID)
		ctx.ShardGroup.guildsMu.Unlock()
	}

	ss.RemoveAllGuildRoles(guildID)
	ss.RemoveAllGuildChannels(guildID)
	ss.RemoveAllGuildEmojis(guildID)
	ss.RemoveAllGuildMembers(guildID)

	delete(ss.Guilds, guildID)
}

//
// GuildMember Operations
//

// GuildMemberFromState converts the structs.StateGuildMembers into a discord.GuildMember,
// for use within the application.
// This will not populate the user object from cache, it will be an empty object with only an ID.
func (ss *SandwichState) GuildMemberFromState(guildState *sandwich_structs.StateGuildMember) (guild *discord.GuildMember) {
	return &discord.GuildMember{
		User: &discord.User{
			ID: guildState.UserID,
		},
		Nick: guildState.Nick,

		Roles:        guildState.Roles,
		JoinedAt:     guildState.JoinedAt,
		PremiumSince: guildState.PremiumSince,
		Deaf:         guildState.Deaf,
		Mute:         guildState.Mute,
		Pending:      guildState.Pending,
		Permissions:  guildState.Permissions,
	}
}

// GuildMemberFromState converts from discord.GuildMember to structs.StateGuildMembers, for storing in cache.
// This does not add the user to the cache.
func (ss *SandwichState) GuildMemberToState(guild *discord.GuildMember) (guildState *sandwich_structs.StateGuildMember) {
	return &sandwich_structs.StateGuildMember{
		UserID: guild.User.ID,
		Nick:   guild.Nick,

		Roles:        guild.Roles,
		JoinedAt:     guild.JoinedAt,
		PremiumSince: guild.PremiumSince,
		Deaf:         guild.Deaf,
		Mute:         guild.Mute,
		Pending:      guild.Pending,
		Permissions:  guild.Permissions,
	}
}

// GetGuildMember returns the guildMember with the same ID from the cache. Populated user field from cache.
// Returns a boolean to signify a match or not.
func (ss *SandwichState) GetGuildMember(guildID discord.Snowflake, guildMemberID discord.Snowflake) (guildMember *discord.GuildMember, ok bool) {
	ss.guildMembersMu.RLock()
	defer ss.guildMembersMu.RUnlock()

	guildMembers, ok := ss.GuildMembers[guildID]
	if !ok {
		return
	}

	guildMembers.MembersMu.RLock()
	defer guildMembers.MembersMu.RUnlock()

	stateGuildMember, ok := guildMembers.Members[guildMemberID]
	if !ok {
		return
	}

	guildMember = ss.GuildMemberFromState(stateGuildMember)

	user, ok := ss.GetUser(guildMember.User.ID)
	if ok {
		guildMember.User = user
	}

	return
}

// SetGuildMember creates or updates a guildMember entry in the cache. Adds user in guildMember object to cache.
func (ss *SandwichState) SetGuildMember(ctx *StateCtx, guildID discord.Snowflake, guildMember *discord.GuildMember) {
	if !ctx.CacheMembers {
		return
	}

	ss.guildMembersMu.Lock()
	defer ss.guildMembersMu.Unlock()

	guildMembers, ok := ss.GuildMembers[guildID]
	if !ok {
		guildMembers = &sandwich_structs.StateGuildMembers{
			MembersMu: sync.RWMutex{},
			Members:   make(map[discord.Snowflake]*sandwich_structs.StateGuildMember),
		}

		ss.GuildMembers[guildID] = guildMembers
	}

	guildMembers.MembersMu.Lock()
	defer guildMembers.MembersMu.Unlock()

	guildMembers.Members[guildMember.User.ID] = ss.GuildMemberToState(guildMember)

	if ctx.CacheUsers {
		ss.SetUser(ctx, guildMember.User)
	}
}

// RemoveGuildMember removes a guildMember from the cache.
func (ss *SandwichState) RemoveGuildMember(guildID discord.Snowflake, guildMemberID discord.Snowflake) {
	ss.guildMembersMu.RLock()
	defer ss.guildMembersMu.RUnlock()

	guildMembers, ok := ss.GuildMembers[guildID]
	if !ok {
		return
	}

	guildMembers.MembersMu.Lock()
	defer guildMembers.MembersMu.Unlock()

	delete(guildMembers.Members, guildMemberID)
}

// GetAllGuildMembers returns all guildMembers of a specific guild from the cache.
func (ss *SandwichState) GetAllGuildMembers(guildID discord.Snowflake) (guildMembersList []*discord.GuildMember, ok bool) {
	ss.guildMembersMu.RLock()
	defer ss.guildMembersMu.RUnlock()

	guildMembers, ok := ss.GuildMembers[guildID]
	if !ok {
		return
	}

	guildMembers.MembersMu.RLock()
	defer guildMembers.MembersMu.RUnlock()

	for _, guildMember := range guildMembers.Members {
		guildMembersList = append(guildMembersList, ss.GuildMemberFromState(guildMember))
	}

	return
}

// RemoveAllGuildMembers removes all guildMembers of a specific guild from the cache.
func (ss *SandwichState) RemoveAllGuildMembers(guildID discord.Snowflake) {
	ss.guildMembersMu.Lock()
	defer ss.guildMembersMu.Unlock()

	delete(ss.GuildMembers, guildID)
}

//
// Role Operations
//

// RoleFromState converts the structs.StateRole into a discord.Role, for use within the application.
func (ss *SandwichState) RoleFromState(guildState *sandwich_structs.StateRole) (guild *discord.Role) {
	return &discord.Role{
		ID:           guildState.ID,
		Name:         guildState.Name,
		Color:        guildState.Color,
		Hoist:        guildState.Hoist,
		Icon:         guildState.Icon,
		UnicodeEmoji: guildState.UnicodeEmoji,
		Position:     guildState.Position,
		Permissions:  guildState.Permissions,
		Managed:      guildState.Managed,
		Mentionable:  guildState.Mentionable,
		Tags:         guildState.Tags,
	}
}

// RoleFromState converts from discord.Role to structs.StateRole, for storing in cache.
func (ss *SandwichState) RoleToState(guild *discord.Role) (guildState *sandwich_structs.StateRole) {
	return &sandwich_structs.StateRole{
		ID:           guild.ID,
		Name:         guild.Name,
		Color:        guild.Color,
		Hoist:        guild.Hoist,
		Icon:         guild.Icon,
		UnicodeEmoji: guild.UnicodeEmoji,
		Position:     guild.Position,
		Permissions:  guild.Permissions,
		Managed:      guild.Managed,
		Mentionable:  guild.Mentionable,
		Tags:         guild.Tags,
	}
}

// GetGuildRole returns the role with the same ID from the cache.
// Returns a boolean to signify a match or not.
func (ss *SandwichState) GetGuildRole(guildID discord.Snowflake, roleID discord.Snowflake) (role *discord.Role, ok bool) {
	ss.guildRolesMu.RLock()
	defer ss.guildRolesMu.RUnlock()

	stateGuildRoles, ok := ss.GuildRoles[roleID]
	if !ok {
		return
	}

	stateGuildRoles.RolesMu.RLock()
	defer stateGuildRoles.RolesMu.RUnlock()

	stateGuildRole, ok := stateGuildRoles.Roles[roleID]
	if !ok {
		return
	}

	role = ss.RoleFromState(stateGuildRole)

	return
}

// SetGuildRole creates or updates a role entry in the cache.
func (ss *SandwichState) SetGuildRole(ctx *StateCtx, guildID discord.Snowflake, role *discord.Role) {
	ss.guildRolesMu.Lock()
	defer ss.guildRolesMu.Unlock()

	guildRoles, ok := ss.GuildRoles[guildID]
	if !ok {
		guildRoles = &sandwich_structs.StateGuildRoles{
			RolesMu: sync.RWMutex{},
			Roles:   make(map[discord.Snowflake]*sandwich_structs.StateRole),
		}

		ss.GuildRoles[guildID] = guildRoles
	}

	guildRoles.RolesMu.Lock()
	defer guildRoles.RolesMu.Unlock()

	guildRoles.Roles[role.ID] = ss.RoleToState(role)
}

// RemoveGuildRole removes a role from the cache.
func (ss *SandwichState) RemoveGuildRole(guildID discord.Snowflake, roleID discord.Snowflake) {
	ss.guildRolesMu.RLock()
	defer ss.guildRolesMu.RUnlock()

	guildRoles, ok := ss.GuildRoles[guildID]
	if !ok {
		return
	}

	guildRoles.RolesMu.Lock()
	defer guildRoles.RolesMu.Unlock()

	delete(guildRoles.Roles, roleID)
}

// GetAllGuildRoles returns all guildRoles of a specific guild from the cache.
func (ss *SandwichState) GetAllGuildRoles(guildID discord.Snowflake) (guildRolesList []*discord.Role, ok bool) {
	ss.guildRolesMu.RLock()
	defer ss.guildRolesMu.RUnlock()

	guildRoles, ok := ss.GuildRoles[guildID]
	if !ok {
		return
	}

	guildRoles.RolesMu.RLock()
	defer guildRoles.RolesMu.RUnlock()

	for _, guildRole := range guildRoles.Roles {
		guildRolesList = append(guildRolesList, ss.RoleFromState(guildRole))
	}

	return
}

// RemoveGuildRoles removes all guild roles of a specifi guild from the cache.
func (ss *SandwichState) RemoveAllGuildRoles(guildID discord.Snowflake) {
	ss.guildRolesMu.Lock()
	defer ss.guildRolesMu.Unlock()

	delete(ss.GuildRoles, guildID)
}

//
// Emoji Operations
//

// EmojiFromState converts the structs.StateEmoji into a discord.Emoji, for use within the application.
func (ss *SandwichState) EmojiFromState(guildState *sandwich_structs.StateEmoji) (guild *discord.Emoji) {
	return &discord.Emoji{
		ID:    guildState.ID,
		Name:  guildState.Name,
		Roles: guildState.Roles,
		User: &discord.User{
			ID: guildState.UserID,
		},
		RequireColons: guildState.RequireColons,
		Managed:       guildState.Managed,
		Animated:      guildState.Animated,
		Available:     guildState.Available,
	}
}

// EmojiFromState converts from discord.Emoji to structs.StateEmoji, for storing in cache.
// This does not add the user to the cache.
// This will not populate the user object from cache, it will be an empty object with only an ID.
func (ss *SandwichState) EmojiToState(emoji *discord.Emoji) (guildState *sandwich_structs.StateEmoji) {
	guildState = &sandwich_structs.StateEmoji{
		ID:            emoji.ID,
		Name:          emoji.Name,
		Roles:         emoji.Roles,
		RequireColons: emoji.RequireColons,
		Managed:       emoji.Managed,
		Animated:      emoji.Animated,
		Available:     emoji.Available,
	}

	if emoji.User != nil {
		guildState.UserID = emoji.User.ID
	}

	return guildState
}

// GetGuildEmoji returns the emoji with the same ID from the cache. Populated user field from cache.
// Returns a boolean to signify a match or not.
func (ss *SandwichState) GetGuildEmoji(guildID discord.Snowflake, emojiID discord.Snowflake) (guildEmoji *discord.Emoji, ok bool) {
	ss.guildEmojisMu.RLock()
	defer ss.guildEmojisMu.RUnlock()

	guildEmojis, ok := ss.GuildEmojis[guildID]
	if !ok {
		return
	}

	guildEmojis.EmojisMu.RLock()
	defer guildEmojis.EmojisMu.RUnlock()

	stateGuildEmoji, ok := guildEmojis.Emojis[emojiID]
	if !ok {
		return
	}

	guildEmoji = ss.EmojiFromState(stateGuildEmoji)

	user, ok := ss.GetUser(guildEmoji.User.ID)
	if ok {
		guildEmoji.User = user
	}

	return
}

// SetGuildEmoji creates or updates a emoji entry in the cache. Adds user in user object to cache.
func (ss *SandwichState) SetGuildEmoji(ctx *StateCtx, guildID discord.Snowflake, emoji *discord.Emoji) {
	ss.guildEmojisMu.Lock()
	defer ss.guildEmojisMu.Unlock()

	guildEmojis, ok := ss.GuildEmojis[guildID]
	if !ok {
		guildEmojis = &sandwich_structs.StateGuildEmojis{
			EmojisMu: sync.RWMutex{},
			Emojis:   make(map[discord.Snowflake]*sandwich_structs.StateEmoji),
		}

		ss.GuildEmojis[guildID] = guildEmojis
	}

	guildEmojis.EmojisMu.Lock()
	defer guildEmojis.EmojisMu.Unlock()

	guildEmojis.Emojis[emoji.ID] = ss.EmojiToState(emoji)

	if emoji.User != nil && ctx.CacheUsers {
		ss.SetUser(ctx, emoji.User)
	}
}

// RemoveGuildEmoji removes a emoji from the cache.
func (ss *SandwichState) RemoveGuildEmoji(guildID discord.Snowflake, emojiID discord.Snowflake) {
	ss.guildEmojisMu.RLock()
	defer ss.guildEmojisMu.RUnlock()

	guildEmojis, ok := ss.GuildEmojis[guildID]
	if !ok {
		return
	}

	guildEmojis.EmojisMu.Lock()
	defer guildEmojis.EmojisMu.Unlock()

	delete(guildEmojis.Emojis, emojiID)
}

// GetAllGuildEmojis returns all guildEmojis on a specific guild from the cache.
func (ss *SandwichState) GetAllGuildEmojis(guildID discord.Snowflake) (guildEmojisList []*discord.Emoji, ok bool) {
	ss.guildEmojisMu.RLock()
	defer ss.guildEmojisMu.RUnlock()

	guildEmojis, ok := ss.GuildEmojis[guildID]
	if !ok {
		return
	}

	guildEmojis.EmojisMu.RLock()
	defer guildEmojis.EmojisMu.RUnlock()

	for _, guildEmoji := range guildEmojis.Emojis {
		guildEmojisList = append(guildEmojisList, ss.EmojiFromState(guildEmoji))
	}

	return
}

// RemoveGuildEmojis removes all guildEmojis of a specific guild from the cache.
func (ss *SandwichState) RemoveAllGuildEmojis(guildID discord.Snowflake) {
	ss.guildEmojisMu.Lock()
	defer ss.guildEmojisMu.Unlock()

	delete(ss.GuildEmojis, guildID)
}

//
// User Operations
//

// UserFromState converts the structs.StateUser into a discord.User, for use within the application.
func (ss *SandwichState) UserFromState(userState *sandwich_structs.StateUser) (user *discord.User) {
	return &discord.User{
		ID:            userState.ID,
		Username:      userState.Username,
		Discriminator: userState.Discriminator,
		Avatar:        userState.Avatar,
		Bot:           userState.Bot,
		System:        userState.System,
		MFAEnabled:    userState.MFAEnabled,
		Banner:        userState.Banner,
		AccentColour:  userState.AccentColour,
		Locale:        userState.Locale,
		Verified:      userState.Verified,
		Email:         userState.Email,
		Flags:         userState.Flags,
		PremiumType:   userState.PremiumType,
		PublicFlags:   userState.PublicFlags,
		DMChannelID:   userState.DMChannelID,
	}
}

// UserFromState converts from discord.User to structs.StateUser, for storing in cache.
func (ss *SandwichState) UserToState(user *discord.User) (userState *sandwich_structs.StateUser) {
	return &sandwich_structs.StateUser{
		ID:            user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		Avatar:        user.Avatar,
		Bot:           user.Bot,
		System:        user.System,
		MFAEnabled:    user.MFAEnabled,
		Banner:        user.Banner,
		AccentColour:  user.AccentColour,
		Locale:        user.Locale,
		Verified:      user.Verified,
		Email:         user.Email,
		Flags:         user.Flags,
		PremiumType:   user.PremiumType,
		PublicFlags:   user.PublicFlags,
		DMChannelID:   user.DMChannelID,
	}
}

// GetUser returns the user with the same ID from the cache.
// Returns a boolean to signify a match or not.
func (ss *SandwichState) GetUser(userID discord.Snowflake) (user *discord.User, ok bool) {
	ss.usersMu.RLock()
	defer ss.usersMu.RUnlock()

	stateUser, ok := ss.Users[userID]
	if !ok {
		return
	}

	user = ss.UserFromState(stateUser)

	return
}

// SetUser creates or updates a user entry in the cache.
func (ss *SandwichState) SetUser(ctx *StateCtx, user *discord.User) {
	if !ctx.CacheUsers {
		return
	}

	ss.usersMu.Lock()
	defer ss.usersMu.Unlock()

	ss.Users[user.ID] = ss.UserToState(user)
}

// RemoveUser removes a user from the cache.
func (ss *SandwichState) RemoveUser(userID discord.Snowflake) {
	ss.usersMu.Lock()
	defer ss.usersMu.Unlock()

	delete(ss.Users, userID)
}

//
// Channel Operations
//

// ChannelFromState converts the structs.StateChannel into a discord.Channel, for use within the application.
// This will not populate the recipient user object from cache.
func (ss *SandwichState) ChannelFromState(guildState *sandwich_structs.StateChannel) (guild *discord.Channel) {
	guild = &discord.Channel{
		ID:                         guildState.ID,
		Type:                       guildState.Type,
		GuildID:                    guildState.GuildID,
		Position:                   guildState.Position,
		PermissionOverwrites:       make([]*discord.ChannelOverwrite, 0, len(guildState.PermissionOverwrites)),
		Name:                       guildState.Name,
		Topic:                      guildState.Topic,
		NSFW:                       guildState.NSFW,
		LastMessageID:              guildState.LastMessageID,
		Bitrate:                    guildState.Bitrate,
		UserLimit:                  guildState.UserLimit,
		RateLimitPerUser:           guildState.RateLimitPerUser,
		Recipients:                 make([]*discord.User, 0, len(guildState.Recipients)),
		Icon:                       guildState.Icon,
		OwnerID:                    guildState.OwnerID,
		ApplicationID:              guildState.ApplicationID,
		ParentID:                   guildState.ParentID,
		LastPinTimestamp:           guildState.LastPinTimestamp,
		RTCRegion:                  guildState.RTCRegion,
		VideoQualityMode:           guildState.VideoQualityMode,
		MessageCount:               guildState.MessageCount,
		MemberCount:                guildState.MemberCount,
		ThreadMetadata:             guildState.ThreadMetadata,
		ThreadMember:               guildState.ThreadMember,
		DefaultAutoArchiveDuration: guildState.DefaultAutoArchiveDuration,
		Permissions:                guildState.Permissions,
	}

	for _, permissionOverride := range guildState.PermissionOverwrites {
		permissionOverride := permissionOverride
		guild.PermissionOverwrites = append(guild.PermissionOverwrites, &permissionOverride)
	}

	for _, recepientID := range guildState.Recipients {
		guild.Recipients = append(guild.Recipients, &discord.User{
			ID: recepientID,
		})
	}

	return guild
}

// ChannelFromState converts from discord.Channel to structs.StateChannel, for storing in cache.
// This does not add the recipients to the cache.
func (ss *SandwichState) ChannelToState(guild *discord.Channel) (guildState *sandwich_structs.StateChannel) {
	guildState = &sandwich_structs.StateChannel{
		ID:                   guild.ID,
		Type:                 guild.Type,
		GuildID:              guild.GuildID,
		Position:             guild.Position,
		PermissionOverwrites: make([]discord.ChannelOverwrite, 0),
		Name:                 guild.Name,
		Topic:                guild.Topic,
		NSFW:                 guild.NSFW,
		// LastMessageID:        guild.LastMessageID,
		Bitrate:          guild.Bitrate,
		UserLimit:        guild.UserLimit,
		RateLimitPerUser: guild.RateLimitPerUser,
		// RecipientIDs:         make([]*discord.Snowflake, 0),
		Icon:    guild.Icon,
		OwnerID: guild.OwnerID,
		// ApplicationID:        guild.ApplicationID,
		ParentID: guild.ParentID,
		// LastPinTimestamp:     guild.LastPinTimestamp,

		// RTCRegion: guild.RTCRegion,
		// VideoQualityMode: guild.VideoQualityMode,

		// MessageCount:               guild.MessageCount,
		// MemberCount:                guild.MemberCount,
		ThreadMetadata: guild.ThreadMetadata,
		// ThreadMember:               guild.ThreadMember,
		// DefaultAutoArchiveDuration: guild.DefaultAutoArchiveDuration,

		Permissions: guild.Permissions,
	}

	for _, permissionOverride := range guild.PermissionOverwrites {
		permissionOverride := permissionOverride
		guildState.PermissionOverwrites = append(guildState.PermissionOverwrites, *permissionOverride)
	}

	// for _, recipient := range guild.Recipients {
	// 	guildState.RecipientIDs = append(guildState.RecipientIDs, &recipient.ID)
	// }

	return guildState
}

// GetGuildChannel returns the channel with the same ID from the cache.
// Returns a boolean to signify a match or not.
func (ss *SandwichState) GetGuildChannel(guildIDPtr *discord.Snowflake, channelID discord.Snowflake) (guildChannel *discord.Channel, ok bool) {
	ss.guildChannelsMu.RLock()
	defer ss.guildChannelsMu.RUnlock()

	var guildID discord.Snowflake

	if guildIDPtr != nil {
		guildID = *guildIDPtr
	} else {
		guildID = discord.Snowflake(0)
	}

	stateChannels, ok := ss.GuildChannels[guildID]
	if !ok {
		return
	}

	stateChannels.ChannelsMu.RLock()
	defer stateChannels.ChannelsMu.RUnlock()

	stateGuildChannel, ok := stateChannels.Channels[channelID]
	if !ok {
		return
	}

	guildChannel = ss.ChannelFromState(stateGuildChannel)

	newRecepients := make([]*discord.User, 0)

	for _, recipient := range guildChannel.Recipients {
		recipientUser, ok := ss.GetUser(recipient.ID)
		if ok {
			recipient = recipientUser
		}

		newRecepients = append(newRecepients, recipient)
	}

	guildChannel.Recipients = newRecepients

	return guildChannel, ok
}

// SetGuildChannel creates or updates a channel entry in the cache.
func (ss *SandwichState) SetGuildChannel(ctx *StateCtx, guildIDPtr *discord.Snowflake, channel *discord.Channel) {
	ss.guildChannelsMu.Lock()
	defer ss.guildChannelsMu.Unlock()

	var guildID discord.Snowflake

	if guildIDPtr != nil {
		guildID = *guildIDPtr
	} else {
		guildID = discord.Snowflake(0)
	}

	guildChannels, ok := ss.GuildChannels[guildID]
	if !ok {
		guildChannels = &sandwich_structs.StateGuildChannels{
			ChannelsMu: sync.RWMutex{},
			Channels:   make(map[discord.Snowflake]*sandwich_structs.StateChannel),
		}

		ss.GuildChannels[guildID] = guildChannels
	}

	guildChannels.ChannelsMu.Lock()
	defer guildChannels.ChannelsMu.Unlock()

	guildChannels.Channels[channel.ID] = ss.ChannelToState(channel)

	if ctx.CacheUsers {
		for _, recipient := range channel.Recipients {
			recipient := recipient
			ss.SetUser(ctx, recipient)
		}
	}
}

// RemoveGuildChannel removes a channel from the cache.
func (ss *SandwichState) RemoveGuildChannel(guildIDPtr *discord.Snowflake, channelID discord.Snowflake) {
	ss.guildChannelsMu.RLock()
	defer ss.guildChannelsMu.RUnlock()

	var guildID discord.Snowflake

	if guildIDPtr != nil {
		guildID = *guildIDPtr
	} else {
		guildID = discord.Snowflake(0)
	}

	guildChannels, ok := ss.GuildChannels[guildID]
	if !ok {
		return
	}

	guildChannels.ChannelsMu.Lock()
	defer guildChannels.ChannelsMu.Unlock()

	delete(guildChannels.Channels, channelID)
}

// GetAllGuildChannels returns all guildChannels of a specific guild from the cache.
func (ss *SandwichState) GetAllGuildChannels(guildID discord.Snowflake) (guildChannelsList []*discord.Channel, ok bool) {
	ss.guildChannelsMu.RLock()
	defer ss.guildChannelsMu.RUnlock()

	guildChannels, ok := ss.GuildChannels[guildID]
	if !ok {
		return
	}

	guildChannels.ChannelsMu.RLock()
	defer guildChannels.ChannelsMu.RUnlock()

	for _, guildRole := range guildChannels.Channels {
		guildChannelsList = append(guildChannelsList, ss.ChannelFromState(guildRole))
	}

	return
}

// RemoveAllGuildChannels removes all guildChannels of a specific guild from the cache.
func (ss *SandwichState) RemoveAllGuildChannels(guildID discord.Snowflake) {
	ss.guildChannelsMu.Lock()
	defer ss.guildChannelsMu.Unlock()

	delete(ss.GuildChannels, guildID)
}

// GetDMChannel returns the DM channel of a user.
func (ss *SandwichState) GetDMChannel(userID discord.Snowflake) (channel *discord.Channel, ok bool) {
	ss.dmChannelsMu.RLock()
	dmChannel, ok := ss.dmChannels[userID]
	ss.dmChannelsMu.RUnlock()

	if !ok || int64(dmChannel.ExpiresAt) < time.Now().Unix() {
		ok = false

		return
	}

	channel = dmChannel.Channel
	dmChannel.ExpiresAt = discord.Int64(time.Now().Add(memberDMExpiration).Unix())

	ss.dmChannelsMu.Lock()
	ss.dmChannels[userID] = dmChannel
	ss.dmChannelsMu.Unlock()

	return
}

// AddDMChannel adds a DM channel to a user.
func (ss *SandwichState) AddDMChannel(userID discord.Snowflake, channel *discord.Channel) {
	ss.dmChannelsMu.Lock()
	defer ss.dmChannelsMu.Unlock()

	dmChannel := &sandwich_structs.StateDMChannel{
		Channel:   channel,
		ExpiresAt: discord.Int64(time.Now().Add(memberDMExpiration).Unix()),
	}

	ss.dmChannels[userID] = dmChannel
}

// RemoveDMChannel removes a DM channel from a user.
func (ss *SandwichState) RemoveDMChannel(userID discord.Snowflake) {
	ss.dmChannelsMu.Lock()
	defer ss.dmChannelsMu.Unlock()

	delete(ss.dmChannels, userID)
}

// GetUserMutualGuilds returns a list of snowflakes of mutual guilds a member is seen on.
func (ss *SandwichState) GetUserMutualGuilds(userID discord.Snowflake) (guildIDs []discord.Snowflake, ok bool) {
	ss.mutualsMu.RLock()
	defer ss.mutualsMu.RUnlock()

	mutualGuilds, ok := ss.Mutuals[userID]
	if !ok {
		return
	}

	mutualGuilds.GuildsMu.RLock()
	defer mutualGuilds.GuildsMu.RUnlock()

	for guildID := range mutualGuilds.Guilds {
		guildIDs = append(guildIDs, guildID)
	}

	return
}

// AddUserMutualGuild adds a mutual guild to a user.
func (ss *SandwichState) AddUserMutualGuild(ctx *StateCtx, userID discord.Snowflake, guildID discord.Snowflake) {
	if !ctx.StoreMutuals {
		return
	}

	ss.mutualsMu.Lock()
	defer ss.mutualsMu.Unlock()

	mutualGuilds, ok := ss.Mutuals[userID]
	if !ok {
		mutualGuilds = &sandwich_structs.StateMutualGuilds{
			GuildsMu: sync.RWMutex{},
			Guilds:   make(map[discord.Snowflake]bool),
		}

		ss.Mutuals[userID] = mutualGuilds
	}

	mutualGuilds.GuildsMu.Lock()
	defer mutualGuilds.GuildsMu.Unlock()

	mutualGuilds.Guilds[guildID] = true
}

// RemoveUserMutualGuild removes a mutual guild from a user.
func (ss *SandwichState) RemoveUserMutualGuild(userID discord.Snowflake, guildID discord.Snowflake) {
	ss.mutualsMu.RLock()
	defer ss.mutualsMu.RUnlock()

	mutualGuilds, ok := ss.Mutuals[userID]
	if !ok {
		return
	}

	mutualGuilds.GuildsMu.Lock()
	defer mutualGuilds.GuildsMu.Unlock()

	delete(mutualGuilds.Guilds, guildID)
}
