package discord

// emoji.go contains all structures for emojis.

// Emoji represents an Emoji on discord.
type Emoji struct {
	ID            Snowflake   `json:"id"`
	GuildID       *Snowflake  `json:"guild_id,omitempty"`
	Name          string      `json:"name"`
	Roles         []Snowflake `json:"roles,omitempty"`
	User          *User       `json:"user,omitempty"`
	RequireColons *bool       `json:"require_colons,omitempty"`
	Managed       *bool       `json:"managed,omitempty"`
	Animated      *bool       `json:"animated,omitempty"`
	Available     *bool       `json:"available,omitempty"`
}
