package gateway

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheRockettek/Sandwich-Daemon/pkg/snowflake"
	"github.com/TheRockettek/Sandwich-Daemon/structs"
	"github.com/TheRockettek/czlib"
	"github.com/rs/zerolog"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/xerrors"
	"nhooyr.io/websocket"
)

// Amount of snowflake ids per guild members request
const chunkSize = 50
const identifyRatelimit = (5 * time.Second) + (500 * time.Millisecond)

// Shard represents the shard object
type Shard struct {
	Status   structs.ShardStatus
	StatusMu sync.RWMutex

	Logger zerolog.Logger

	ShardID    int
	ShardGroup *ShardGroup
	Manager    *Manager

	User *structs.User
	// TODO: Add deque that can allow for an event queue (maybe)

	ctx    context.Context
	cancel func()

	LastHeartbeatMu   sync.RWMutex
	LastHeartbeatAck  time.Time
	LastHeartbeatSent time.Time

	Heartbeater          *time.Ticker
	HeartbeatInterval    time.Duration
	MaxHeartbeatFailures time.Duration

	wsConn  *websocket.Conn
	wsMutex sync.Mutex

	rp sync.Pool
	pp sync.Pool

	msg structs.ReceivedPayload
	buf []byte

	events        *int64
	executionTime *int64

	seq       *int64
	sessionID string

	// Channel that dictates if the shard has been made ready
	ready chan void
	// Channel to pipe errors
	errs chan error
}

// NewShard creates a new shard object
func (sg *ShardGroup) NewShard(shardID int) *Shard {
	logger := sg.Logger.With().Int("shard", shardID).Logger()
	sh := &Shard{
		Status:   structs.ShardIdle,
		StatusMu: sync.RWMutex{},

		Logger: logger,

		ShardID:    shardID,
		ShardGroup: sg,
		Manager:    sg.Manager,

		ctx: context.Background(),

		LastHeartbeatMu:   sync.RWMutex{},
		LastHeartbeatAck:  time.Now().UTC(),
		LastHeartbeatSent: time.Now().UTC(),

		rp: sync.Pool{
			New: func() interface{} { return new(structs.SentPayload) },
		},
		pp: sync.Pool{
			New: func() interface{} { return new(structs.PublishEvent) },
		},
		msg: structs.ReceivedPayload{},
		buf: make([]byte, 0),

		events:        new(int64),
		executionTime: new(int64),

		seq:       new(int64),
		sessionID: "",

		ready: make(chan void),
		errs:  make(chan error),
	}
	return sh
}

// Open starts up the shard connection
func (sh *Shard) Open() {
	for {
		err := sh.Listen()
		if xerrors.Is(err, context.Canceled) {
			return
		}

		// Check if context is done
		select {
		case <-sh.ctx.Done():
			return
		default:
		}
	}
}

// Connect connects to the gateway such as identifying however does not listen to new messages
func (sh *Shard) Connect() (err error) {
	sh.Logger.Debug().Msg("Starting shard")

	sh.ctx, sh.cancel = context.WithCancel(context.Background())
	gatewayURL := sh.Manager.Gateway.URL

	sh.SetStatus(structs.ShardWaiting)
	err = sh.Manager.Sandwich.Buckets.CreateWaitForBucket(fmt.Sprintf("gw:%s:%d", sh.Manager.Configuration.Token, sh.ShardID%sh.Manager.Gateway.SessionStartLimit.MaxConcurrency), 1, identifyRatelimit)
	if err != nil {
		return
	}

	sh.SetStatus(structs.ShardConnecting)
	sh.Logger.Debug().Str("gurl", gatewayURL).Msg("Connecting to gateway")

	// TODO: Add Concurrent Client Support
	// This will limit the ammount of shards that can be connecting simultaneously
	// May be abandoned as this boy is fast af :pepega:
	// Could help with a shit ton running at once whilst scaling

	defer func() {
		if err != nil {
			sh.CloseWS(websocket.StatusNormalClosure)
		}
	}()

	if sh.wsConn == nil {
		var conn *websocket.Conn
		conn, _, err = websocket.Dial(sh.ctx, gatewayURL, nil)
		if err != nil {
			return
		}
		conn.SetReadLimit(512 << 20)
		sh.wsConn = conn
	} else {
		sh.Logger.Info().Msg("Reusing websocket connection")
	}

	err = sh.readMessage(sh.ctx, sh.wsConn)
	if err != nil {
		return
	}

	hello := structs.Hello{}
	err = sh.decodeContent(&hello)

	sh.LastHeartbeatMu.Lock()
	sh.LastHeartbeatAck = time.Now().UTC()
	sh.LastHeartbeatSent = time.Now().UTC()
	sh.LastHeartbeatMu.Unlock()

	sh.HeartbeatInterval = hello.HeartbeatInterval * time.Millisecond
	sh.MaxHeartbeatFailures = sh.HeartbeatInterval * time.Duration(sh.Manager.Configuration.Bot.MaxHeartbeatFailures)

	sh.Logger.Debug().Dur("interval", sh.HeartbeatInterval).Int("maxfails", sh.Manager.Configuration.Bot.MaxHeartbeatFailures).Msg("Retrieved HELLO event from discord")
	sh.Heartbeater = time.NewTicker(sh.HeartbeatInterval)

	seq := atomic.LoadInt64(sh.seq)
	if sh.sessionID == "" || seq == 0 {
		err = sh.Identify()
		if err != nil {
			return
		}
	} else {
		err = sh.Resume()
		if err != nil {
			return
		}
	}

	err = sh.readMessage(sh.ctx, sh.wsConn)
	if err != nil {
		return
	}

	err = sh.OnEvent(sh.msg)
	if err != nil {
		sh.Logger.Error().Err(err).Msg("Error whilst handling event")
	}

	go sh.Heartbeat()
	sh.SetStatus(structs.ShardConnected)
	return
}

// OnEvent processes an event
func (sh *Shard) OnEvent(msg structs.ReceivedPayload) (err error) {

	switch msg.Op {

	case structs.GatewayOpHeartbeat:
		sh.Logger.Debug().Msg("Received heartbeat request")
		err = sh.SendEvent(structs.GatewayOpHeartbeat, atomic.LoadInt64(sh.seq))
		if err != nil {
			sh.Logger.Error().Err(err).Msg("Failed to send heartbeat in response to gateway, reconnecting...")
			sh.Reconnect(websocket.StatusNormalClosure)
			return
		}

	case structs.GatewayOpInvalidSession:
		resumable := json.Get(msg.Data, "d").ToBool()
		sh.Logger.Warn().Bool("resumable", resumable).Msg("Received invalid session from gateway")

		if !resumable || (sh.sessionID == "" || atomic.LoadInt64(sh.seq) == 0) {
			err = sh.Identify()
			if err != nil {
				sh.Logger.Error().Err(err).Msg("Failed to send identify in response to gateway, reconnecting...")
				sh.Reconnect(websocket.StatusNormalClosure)
				return
			}
		} else {
			err = sh.Resume()
			if err != nil {
				sh.Logger.Error().Err(err).Msg("Failed to send identify in response to gateway, reconnecting...")
				sh.Reconnect(websocket.StatusNormalClosure)
				return
			}
		}

	case structs.GatewayOpHello:
		sh.Logger.Warn().Msg("Received HELLO whilst listening. This should not occur.")
		return

	case structs.GatewayOpReconnect:
		sh.Logger.Info().Msg("Reconnecting in response to gateway")
		sh.Reconnect(4000)
		return

	case structs.GatewayOpDispatch:
		err = sh.OnDispatch(msg)
		if err != nil {
			sh.Logger.Error().Err(err).Msg("Error whilst dispatch event")
			return
		}

	case structs.GatewayOpHeartbeatACK:
		sh.LastHeartbeatMu.Lock()
		sh.LastHeartbeatAck = time.Now().UTC()
		sh.Logger.Debug().Int64("RTT", sh.LastHeartbeatAck.Sub(sh.LastHeartbeatSent).Milliseconds()).Msg("Received heartbeack ACK")
		sh.LastHeartbeatMu.Unlock()
		return

	default:
		sh.Logger.Warn().Int("op", int(msg.Op)).Str("type", msg.Type).Msg("Gateway sent unknown packet")
		return
	}

	atomic.StoreInt64(sh.seq, msg.Sequence)
	return
}

// OnDispatch handles a dispatch event
func (sh *Shard) OnDispatch(msg structs.ReceivedPayload) (err error) {
	start := time.Now().UTC()
	defer func() {
		change := time.Now().UTC().Sub(start)
		if change > time.Second {
			sh.Logger.Warn().Msgf("%s took %d ms", msg.Type, change.Milliseconds())
		}
	}()

	switch msg.Type {

	case "READY":
		readyPayload := structs.Ready{}
		if err = sh.decodeContent(&readyPayload); err != nil {
			return
		}
		sh.User = readyPayload.User
		sh.sessionID = readyPayload.SessionID
		sh.Logger.Info().Msg("Received READY payload")

		unavailables := make(map[snowflake.ID]bool)
		events := make([]structs.ReceivedPayload, 0)

		guildIDs := make([]int64, 0)

		for _, guild := range readyPayload.Guilds {
			unavailables[guild.ID] = guild.Unavailable
		}

		// I really wanted to just use a context here but it kept cancelling the main
		// context even if i made a completely new one. Hmu if you know a solution.
		wait := time.Now().UTC().Add(2 * time.Second)
		for {
			timedout := time.Now().UTC().Sub(wait) > (2 * time.Second)

			if !timedout {
				err = sh.readMessage(sh.ctx, sh.wsConn)
			}

			if err != nil || timedout {
				if xerrors.Is(err, context.Canceled) || timedout {
					sh.Logger.Debug().Msg("Shard has finished lazy loading")
					err = nil
				} else {
					sh.Logger.Error().Err(err).Msg("Errored whilst waiting lazy loading")
				}

				if sh.Manager.Configuration.Caching.RequestMembers {
					var chunk []int64
					for len(guildIDs) >= chunkSize {
						chunk, guildIDs = guildIDs[:chunkSize], guildIDs[chunkSize:]

						sh.Logger.Trace().Msgf("Requesting guild members for %d guild(s)", len(chunk))
						if err := sh.SendEvent(structs.GatewayOpRequestGuildMembers, structs.RequestGuildMembers{
							GuildID: chunk,
							Query:   "",
							Limit:   0,
						}); err != nil {
							sh.Logger.Error().Err(err).Msgf("Failed to request guild members")
						}
					}
					if len(guildIDs) > 0 {
						sh.Logger.Trace().Msgf("Requesting guild members for %d guild(s)", len(chunk))
						if err := sh.SendEvent(structs.GatewayOpRequestGuildMembers, structs.RequestGuildMembers{
							GuildID: guildIDs,
							Query:   "",
							Limit:   0,
						}); err != nil {
							sh.Logger.Error().Err(err).Msgf("Failed to request guild members")
						}
					}
				}

				break
			}

			if sh.msg.Type == "GUILD_CREATE" {
				guildPayload := structs.GuildCreate{}
				if err = sh.decodeContent(&guildPayload); err != nil {
					sh.Logger.Error().Err(err).Msg("Failed to unmarshal GUILD_CREATE whilst lazy loading")
				} else {
					if ok, unavailable := unavailables[guildPayload.ID]; ok && unavailable {
						// Guild has been lazy loaded
						sh.Logger.Trace().Msgf("Lazy loaded guild ID %d", guildPayload.ID)
						guildPayload.Lazy = true
					}
					guildIDs = append(guildIDs, guildPayload.ID.Int64())

					// StateGuildCreate(guildPayload)
				}

				wait = time.Now().UTC().Add(2 * time.Second)
			} else {
				events = append(events, sh.msg)
			}
		}

		close(sh.ready)
		sh.SetStatus(structs.ShardReady)

		sh.Logger.Debug().Int("events", len(events)).Msg("Dispatching preemtive events")
		for _, event := range events {
			sh.Logger.Debug().Str("type", event.Type).Send()
			if err = sh.OnDispatch(event); err != nil {
				sh.Logger.Error().Err(err).Msg("Failed whilst dispatching preemtive events")
			}
		}

		sh.Logger.Debug().Msg("Finished dispatching events")
		return

	case "GUILD_MEMBERS_CHUNK":
		guildMembersPayload := structs.GuildMembersChunk{}
		if err = sh.decodeContent(&guildMembersPayload); err != nil {
			sh.Logger.Error().Err(err).Msg("Failed to unmarshal GUILD_MEMBERS_CHUNK whilst lazy loading")
		} else {
			if err = sh.Manager.StateGuildMembersChunk(guildMembersPayload); err != nil {
				sh.Logger.Error().Err(err).Msg("Failed to process state")
			}
		}

	// case "GUILD_CREATE":
	// 	guildCreatePayload := structs.GuildCreate{}
	// 	if err = sh.decodeContent(&guildCreatePayload); err != nil {
	// 		return
	// 	}

	default:
		// sh.Logger.Warn().Str("type", msg.Type).Msg("No handler for dispatch message")
	}
	return
}

// Listen to gateway and process accordingly
func (sh *Shard) Listen() (err error) {
	wsConn := sh.wsConn
	evnts := int64(0)
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-sh.ctx.Done():
			return
		case <-t.C:
			atomic.AddInt64(sh.ShardGroup.Events, evnts)
			evnts = 0
		default:
		}

		err = sh.readMessage(sh.ctx, wsConn)
		if err != nil {
			if xerrors.Is(err, context.Canceled) || xerrors.Is(err, context.DeadlineExceeded) {
				break
			}

			sh.Logger.Error().Err(err).Msg("Error reading from gateway")
			if wsConn == sh.wsConn {
				// We have likely closed so we should attempt to reconnect
				sh.Logger.Warn().Msg("We have encountered an error whilst in the same connection, reconnecting...")
				sh.Reconnect(websocket.StatusNormalClosure)
			}
			wsConn = sh.wsConn
		}

		start := time.Now().UTC()

		sh.OnEvent(sh.msg)

		// In the event we have reconnected, the wsConn could have changed,
		// we will use the new wsConn if this is the case
		if sh.wsConn != wsConn {
			sh.Logger.Debug().Msg("New wsConn was assigned to shard")
			wsConn = sh.wsConn
		}

		atomic.AddInt64(sh.executionTime, time.Now().UTC().Sub(start).Nanoseconds())
	}
	return
}

// Heartbeat maintains a heartbeat with discord
// TODO: Make a shardgroup specific heartbeat function to heartbeat on behalf of all running shards
func (sh *Shard) Heartbeat() {
	for {
		select {
		case <-sh.ctx.Done():
			return
		case <-sh.Heartbeater.C:
			sh.Logger.Debug().Msg("Heartbeating")
			seq := atomic.LoadInt64(sh.seq)

			err := sh.SendEvent(structs.GatewayOpHeartbeat, seq)

			_time := time.Now().UTC()
			sh.LastHeartbeatMu.Lock()
			sh.LastHeartbeatSent = _time
			lastAck := sh.LastHeartbeatAck
			sh.LastHeartbeatMu.Unlock()

			if err != nil || _time.Sub(lastAck) > sh.MaxHeartbeatFailures {
				if err != nil {
					sh.Logger.Error().Err(err).Msg("Failed to heartbeat. Reconnecting")
				} else {
					sh.Logger.Warn().Err(err).Msgf("Gateway failed to ACK and has passed MaxHeartbeatFailures of %d. Reconnecting", sh.Manager.Configuration.Bot.MaxHeartbeatFailures)
				}

				sh.Reconnect(websocket.StatusNormalClosure)
				return
			}
		}
	}
}

// decodeContent converts the stored msg into the passed interface
func (sh *Shard) decodeContent(out interface{}) (err error) {
	err = json.Unmarshal(sh.msg.Data, &out)
	return
}

// readMessage fills the shard msg buffer from a websocket message
func (sh *Shard) readMessage(ctx context.Context, wsConn *websocket.Conn) (err error) {
	var mt websocket.MessageType
	mt, sh.buf, err = wsConn.Read(ctx)
	select {
	case <-ctx.Done():
		return
	default:
	}

	if err != nil {
		return xerrors.Errorf("readMessage read: %w", err)
	}

	if mt == websocket.MessageBinary {
		sh.buf, err = czlib.Decompress(sh.buf)
		if err != nil {
			return xerrors.Errorf("readMessage decompress: %w", err)
		}
	}

	err = json.Unmarshal(sh.buf, &sh.msg)
	atomic.AddInt64(sh.events, 1)
	return
}

// CloseWS closes the websocket
func (sh *Shard) CloseWS(statusCode websocket.StatusCode) (err error) {
	sh.Logger.Debug().Str("code", statusCode.String()).Msg("Closing websocket connection")

	if sh.wsConn != nil {
		err = sh.wsConn.Close(statusCode, "")
		sh.wsConn = nil
	}
	return
}

// Resume sends the resume packet to gateway
func (sh *Shard) Resume() (err error) {
	sh.Logger.Debug().Msg("Sending resume")

	return sh.SendEvent(structs.GatewayOpResume, structs.Resume{
		Token:     sh.Manager.Configuration.Token,
		SessionID: sh.sessionID,
		Sequence:  atomic.LoadInt64(sh.seq),
	})
}

// Identify sends the identify packet to gateway
func (sh *Shard) Identify() (err error) {
	sh.Logger.Debug().Msg("Sending identify")

	return sh.SendEvent(structs.GatewayOpIdentify, structs.Identify{
		Token: sh.Manager.Configuration.Token,
		Properties: &structs.IdentifyProperties{
			OS:      runtime.GOOS,
			Browser: "Sandwich " + VERSION,
			Device:  "Sandwich " + VERSION,
		},
		Compress:           sh.Manager.Configuration.Bot.Compression,
		LargeThreshold:     sh.Manager.Configuration.Bot.LargeThreshold,
		Shard:              [2]int{sh.ShardID, sh.ShardGroup.ShardCount},
		Presence:           sh.Manager.Configuration.Bot.DefaultPresence,
		GuildSubscriptions: sh.Manager.Configuration.Bot.GuildSubscriptions,
		Intents:            sh.Manager.Configuration.Bot.Intents,
	})
}

// PublishEvent sends an event to consaumers
func (sh *Shard) PublishEvent(Type string, Data interface{}) (err error) {
	packet := sh.pp.Get().(*structs.PublishEvent)
	defer sh.pp.Put(packet)

	packet.Data = Data
	packet.From = sh.Manager.Configuration.Identifier
	packet.From = Type

	data, err := msgpack.Marshal(packet)
	if err != nil {
		return xerrors.Errorf("publishEvent marshal: %w", err)
	}

	err = sh.Manager.StanClient.Publish(
		sh.Manager.Configuration.Messaging.ChannelName,
		data,
	)
	if err != nil {
		return xerrors.Errorf("publishEvent publish: %w", err)
	}

	return
}

// SendEvent sends an event to discord
func (sh *Shard) SendEvent(op structs.GatewayOp, data interface{}) (err error) {
	packet := sh.rp.Get().(*structs.SentPayload)
	defer sh.rp.Put(packet)

	packet.Op = int(op)
	packet.Data = data

	err = sh.WriteJSON(packet)
	if err != nil {
		return xerrors.Errorf("sendEvent writeJson: %w", err)
	}

	return
}

// WriteJSON writes json data to the websocket
func (sh *Shard) WriteJSON(i interface{}) (err error) {
	res, err := json.Marshal(i)
	if err != nil {
		return xerrors.Errorf("writeJSON marshal: %w", err)
	}

	sh.Manager.Buckets.CreateWaitForBucket(
		fmt.Sprintf("ws:%d:%d", sh.ShardID, sh.ShardGroup.ShardCount),
		120,
		time.Minute,
	)

	sh.Logger.Trace().Msg(string(res))
	err = sh.wsConn.Write(sh.ctx, websocket.MessageText, res)
	if err != nil {
		return xerrors.Errorf("writeJSON write: %w", err)
	}

	return
}

// WaitForReady waits until the shard is ready
func (sh *Shard) WaitForReady() {
	select {
	case <-sh.ready:
	case <-sh.ctx.Done():
	}
	return
}

// Reconnect attempts to reconnect to the gateway
func (sh *Shard) Reconnect(code websocket.StatusCode) {
	wait := time.Second

	sh.Close(code)
	sh.SetStatus(structs.ShardReconnecting)

	for {
		sh.Logger.Info().Msg("Trying to reconnect to gateway")

		err := sh.Connect()
		if err == nil {
			sh.Logger.Info().Msg("Successfuly reconnected to gateway")
			return
		}

		sh.Logger.Warn().Err(err).Dur("retry", wait).Msg("Failed to reconnect to gateway")
		<-time.After(wait)

		wait *= 2
		if wait > 600 {
			wait = 600
		}
	}
}

// SetStatus changes the Shard status
func (sh *Shard) SetStatus(status structs.ShardStatus) {
	sh.StatusMu.Lock()
	sh.Status = status
	sh.StatusMu.Unlock()
	sh.PublishEvent("SHARD_STATUS", structs.MessagingStatusUpdate{ShardID: sh.ShardID, Status: int32(status)})
}

// Close closes the shard connection
func (sh *Shard) Close(code websocket.StatusCode) {
	// Ensure that if we close during shardgroup connecting, it will not
	// feedback loop.

	// cancel is only defined when Connect() has been ran on a shard.
	// If the ShardGroup was closed before this happens, it would segfault.
	if sh.ctx != nil && sh.cancel != nil {
		sh.cancel()
	}
	sh.CloseWS(code)
}
