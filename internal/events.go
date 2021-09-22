package internal

import (
	"context"
	"sync"
	"time"

	discord "github.com/WelcomerTeam/Sandwich-Daemon/next/discord/structs"
	structs "github.com/WelcomerTeam/Sandwich-Daemon/next/structs"
	"github.com/savsgio/gotils/strings"
	"golang.org/x/xerrors"
)

// List of handlers for gateway events.
var gatewayHandlers = make(map[discord.GatewayOp]func(ctx context.Context, sh *Shard, msg discord.GatewayPayload) (err error))

// List of handlers for dispatch events.
var dispatchHandlers = make(map[string]func(ctx *StateCtx, msg discord.GatewayPayload) (result structs.StateResult, ok bool, err error))

type StateCtx struct {
	context context.Context

	*Shard

	Vars map[string]interface{}
}

// SandwichState stores the collective state of all ShardGroups
// across all Managers.
type SandwichState struct {
	guildsMu sync.RWMutex
	Guilds   map[discord.Snowflake]*structs.StateGuild

	guildMembersMu sync.RWMutex
	GuildMembers   map[discord.Snowflake]*structs.StateGuildMembers

	channelsMu sync.RWMutex
	Channels   map[discord.Snowflake]*structs.StateChannel

	rolesMu sync.RWMutex
	Roles   map[discord.Snowflake]*structs.StateRole

	emojisMu sync.RWMutex
	Emojis   map[discord.Snowflake]*structs.StateEmoji

	usersMu sync.RWMutex
	Users   map[discord.Snowflake]*structs.StateUser
}

func NewSandwichState() (st *SandwichState) {
	st = &SandwichState{
		guildsMu: sync.RWMutex{},
		Guilds:   make(map[discord.Snowflake]*structs.StateGuild),

		guildMembersMu: sync.RWMutex{},
		GuildMembers:   make(map[discord.Snowflake]*structs.StateGuildMembers),

		channelsMu: sync.RWMutex{},
		Channels:   make(map[discord.Snowflake]*structs.StateChannel),

		rolesMu: sync.RWMutex{},
		Roles:   make(map[discord.Snowflake]*structs.StateRole),

		emojisMu: sync.RWMutex{},
		Emojis:   make(map[discord.Snowflake]*structs.StateEmoji),

		usersMu: sync.RWMutex{},
		Users:   make(map[discord.Snowflake]*structs.StateUser),
	}

	return st
}

func (sh *Shard) OnEvent(ctx context.Context, msg discord.GatewayPayload) {
	fin := make(chan void, 1)

	go func() {
		since := time.Now()

		t := time.NewTicker(DispatchWarningTimeout)
		defer t.Stop()

		for {
			select {
			case <-fin:
				return
			case <-t.C:
				sh.Logger.Warn().
					Str("type", msg.Type).
					Int("op", int(msg.Op)).
					Dur("since", time.Now().Sub(since)).
					Msg("Event is taking too long")
			}
		}
	}()

	defer close(fin)

	err := GatewayDispatch(ctx, sh, msg)
	if err != nil {
		if xerrors.Is(err, ErrNoGatewayHandler) {
			sh.Logger.Warn().
				Int("op", int(msg.Op)).
				Str("type", msg.Type).
				Msg("Gateway sent unknown packet")
		}
	}

	return
}

// OnDispatch handles routing of discord event.
func (sh *Shard) OnDispatch(ctx context.Context, msg discord.GatewayPayload) (err error) {
	if sh.Manager.ProducerClient == nil {
		return ErrProducerMissing
	}

	sh.Manager.eventBlacklistMu.RLock()
	contains := strings.Include(sh.Manager.eventBlacklist, msg.Type)
	sh.Manager.eventBlacklistMu.RUnlock()

	if contains {
		return
	}

	result, continuable, err := StateDispatch(&StateCtx{
		Shard: sh,
	}, msg)

	if err != nil {
		return err
	}

	if !continuable {
		return
	}

	sh.Manager.produceBlacklistMu.RLock()
	contains = strings.Include(sh.Manager.produceBlacklist, msg.Type)
	sh.Manager.produceBlacklistMu.RUnlock()

	if contains {
		return
	}

	packet := sh.Sandwich.payloadPool.Get().(*structs.SandwichPayload)
	defer sh.Sandwich.payloadPool.Put(packet)

	packet.GatewayPayload = msg
	packet.Data = result.Data
	packet.Extra = result.Extra

	return sh.PublishEvent(ctx, packet)
}

func registerGatewayEvent(op discord.GatewayOp, handler func(ctx context.Context, sh *Shard, msg discord.GatewayPayload) (err error)) {
	gatewayHandlers[op] = handler
}

func registerDispatch(eventType string, handler func(ctx *StateCtx, msg discord.GatewayPayload) (result structs.StateResult, ok bool, err error)) {
	dispatchHandlers[eventType] = handler
}

// GatewayDispatch handles selecting the proper gateway handler and executing it.
func GatewayDispatch(ctx context.Context, sh *Shard,
	event discord.GatewayPayload) (err error) {
	if f, ok := gatewayHandlers[event.Op]; ok {
		return f(ctx, sh, event)
	}

	sh.Logger.Warn().Int("op", int(event.Op)).Msg("No gateway handler found")

	return ErrNoGatewayHandler
}

// StateDispatch handles selecting the proper state handler and executing it.
func StateDispatch(ctx *StateCtx,
	event discord.GatewayPayload) (result structs.StateResult, ok bool, err error) {
	if f, ok := dispatchHandlers[event.Type]; ok {
		ctx.Logger.Trace().Str("type", event.Type).Msg("State Dispatch")

		return f(ctx, event)
	}

	// ctx.Logger.Warn().Str("type", event.Type).Msg("No dispatch handler found")

	return result, false, ErrNoDispatchHandler
}