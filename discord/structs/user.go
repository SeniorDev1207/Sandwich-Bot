package discord

import "github.com/WelcomerTeam/RealRock/snowflake"

// user.go represents all structures for a discord user.

// UserFlags represents the flags on a user's account.
type UserFlags int

// User flags.
const (
	UserFlagsNone UserFlags = 1 << iota
	UserFlagsDiscordEmployee
	UserFlagsPartneredServerOwner
	UserFlagsHypeSquadEvents
	UserFlagsBugHunterLevel1
	UserFlagsHouseBravery
	UserFlagsHouseBrilliance
	UserFlagsHouseBalance
	UserFlagsEarlySupporter
	UserFlagsTeamUser
	UserFlagsSystem
	UserFlagsBugHunterLevel2
	UserFlagsVerifiedBot
	UserFlagsEarlyVerifiedBotDeveloper
)

// UserPremiumType represents the type of Nitro on a user's account.
type UserPremiumType int

// User premium type.
const (
	UserPremiumTypeNone UserPremiumType = iota
	UserPremiumTypeNitroClassic
	UserPremiumTypeNitro
)

// User represents a user on Discord.
type User struct {
	ID            snowflake.ID     `json:"id"`
	Username      string           `json:"username"`
	Discriminator string           `json:"discriminator"`
	Avatar        *string          `json:"avatar,omitempty"`
	Bot           *bool            `json:"bot,omitempty"`
	System        *bool            `json:"system,omitempty"`
	MFAEnabled    *bool            `json:"mfa_enabled,omitempty"`
	Banner        *string          `json:"banner,omitempty"`
	Locale        *string          `json:"locale,omitempty"`
	Verified      *bool            `json:"verified,omitempty"`
	Email         *string          `json:"email,omitempty"`
	Flags         *UserFlags       `json:"flags,omitempty"`
	PremiumType   *UserPremiumType `json:"premium_type,omitempty"`
	PublicFlags   *UserFlags       `json:"public_flags,omitempty"`
}
