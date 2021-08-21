package discord

// invites.go contains all structures for invites.

// InviteTargetType represents the type of an invites target
type InviteTargetType int8

const (
	InviteTargetTypeStream InviteTargetType = 1 + iota
	InviteTargetTypeEmbeddedApplication
)
