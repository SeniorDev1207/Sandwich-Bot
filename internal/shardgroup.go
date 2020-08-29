package gateway

import (
	"sync"

	"github.com/TheRockettek/Sandwich-Daemon/structs"
	"github.com/rs/zerolog"
	"github.com/tevino/abool"
	"golang.org/x/net/context"
	"golang.org/x/xerrors"
)

// ShardGroup groups a selection of shards
type ShardGroup struct {
	StatusMu sync.RWMutex             `json:"-"`
	Status   structs.ShardGroupStatus `json:"status"`
	Error    error                    `json:"error"`

	Manager *Manager       `json:"-"`
	Logger  zerolog.Logger `json:"-"`

	ShardCount int            `json:"shard_count"`
	ShardIDs   []int          `json:"shard_ids"`
	Shards     map[int]*Shard `json:"shards"`

	// WaitGroup for detecting when all shards are ready
	Wait *sync.WaitGroup `json:"-"`

	// Used for detecting errors during shard startup
	err chan error
	// Used to close active goroutines
	close chan void

	floodgate *abool.AtomicBool
}

// NewShardGroup creates a new shardgroup
func (mg *Manager) NewShardGroup() *ShardGroup {
	return &ShardGroup{
		Status:   structs.ShardGroupIdle,
		StatusMu: sync.RWMutex{},
		Error:    nil,

		Manager: mg,
		Logger:  mg.Logger,

		Shards: make(map[int]*Shard),
		Wait:   &sync.WaitGroup{},
		err:    make(chan error),
		close:  make(chan void),

		floodgate: abool.New(),
	}
}

// Open starts up the shardgroup
func (sg *ShardGroup) Open(ShardIDs []int, ShardCount int) (ready chan bool, err error) {
	sg.SetStatus(structs.ShardGroupStarting)

	sg.ShardCount = ShardCount
	sg.ShardIDs = ShardIDs

	ready = make(chan bool, 1)

	sg.Logger.Info().Msgf("Starting ShardGroup with %d shards", len(sg.ShardIDs))

	for _, shardID := range sg.ShardIDs {
		shard := sg.NewShard(shardID)
		sg.Shards[shardID] = shard

		err = shard.Connect()
		if err != nil && !xerrors.Is(err, context.Canceled) {
			sg.Logger.Error().Err(err).Msg("Failed to connect shard. Cannot continue")
			sg.Close()
			sg.SetStatus(structs.ShardGroupError)
			err = xerrors.Errorf("ShardGroup open: %w", err)
			sg.Error = err
			return
		}

		go shard.Open()
	}

	sg.Logger.Debug().Msgf("All shards are now listening")
	sg.SetStatus(structs.ShardGroupConnecting)

	go func(sg *ShardGroup) {
		for _, shard := range sg.Shards {
			sg.Logger.Debug().Msgf("Waiting for shard %d to be ready", shard.ShardID)
			shard.WaitForReady()
		}
		sg.Logger.Debug().Msg("All shards in ShardGroup are ready")
		sg.SetStatus(structs.ShardGroupReady)
		close(ready)
	}(sg)

	return
}

// SetStatus changes the ShardGroup status
func (sg *ShardGroup) SetStatus(status structs.ShardGroupStatus) {
	sg.StatusMu.Lock()
	sg.Status = status
	sg.StatusMu.Unlock()
	sg.Manager.PublishEvent("SHARD_STATUS", structs.MessagingStatusUpdate{Status: int32(status)})
}

// Close closes the shard group and finishes any shards
func (sg *ShardGroup) Close() {
	sg.Logger.Info().Msg("Closing ShardGroup")
	sg.SetStatus(structs.ShardGroupClosing)
	for _, shard := range sg.Shards {
		shard.Close(1000)
	}
	sg.SetStatus(structs.ShardGroupClosed)
}
