package gateway

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/TheRockettek/Sandwich-Daemon/structs"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"golang.org/x/xerrors"
)

var colours = [][]string{
	{"rgba(149, 165, 165, 0.5)", "#7E8C8D"},
	{"rgba(236, 240, 241, 0.5)", "#BEC3C7"},
	{"rgba(232, 76, 61, 0.5)", "#C1392B"},
	{"rgba(231, 126, 35, 0.5)", "#D25400"},
	{"rgba(241, 196, 15, 0.5)", "#F39C11"},
	{"rgba(52, 73, 94, 0.5)", "#2D3E50"},
	{"rgba(155, 88, 181, 0.5)", "#8F44AD"},
	{"rgba(53, 152, 219, 0.5)", "#2A80B9"},
	{"rgba(45, 204, 112, 0.5)", "#27AE61"},
	{"rgba(27, 188, 155, 0.5)", "#16A086"},
}

// RestResponse is the response when returning rest requests
type RestResponse struct {
	Success  bool        `json:"success"`
	Response interface{} `json:"response,omitempty"`
	Error    error       `json:"error,omitempty"`
}

// HandleRequest handles any incomming HTTP requests
func (sg *Sandwich) HandleRequest(ctx *fasthttp.RequestCtx) {
	var res []byte
	var err error

	defer func() {
		sg.Logger.Info().Msgf("%s %s %s %d",
			ctx.RemoteAddr(),
			ctx.Request.Header.Method(),
			ctx.Request.URI().Path(),
			ctx.Response.StatusCode())
	}()

	path := string(ctx.Request.URI().Path())
	if strings.HasPrefix(path, "/static") {
		_, filename := filepath.Split(path)
		root, _ := os.Getwd()
		filepath := filepath.Join(root, "web/static", filename)

		if _, err := os.Stat(filepath); !os.IsNotExist(err) {
			ctx.SendFile(filepath)
		} else {
			ctx.SetStatusCode(404)
		}
	} else {
		switch path {
		case "/":
			b, _ := ioutil.ReadFile("web/spa.html")
			ctx.Response.Header.Set("content-type", "text/html;charset=UTF-8")
			ctx.Write(b)

			// ctx.SendFile("web/spa.html")
			ctx.SetStatusCode(200)

		case "/api/configuration":
			if sg.Configuration.HTTP.Enabled {
				res, err = json.Marshal(RestResponse{true, sg.Managers, nil})
			} else {
				res, err = json.Marshal(RestResponse{false, "HTTP Interface is not enabled", nil})
			}

			if err == nil {
				ctx.Write(res)
				ctx.Response.Header.Set("content-type", "application/javascript;charset=UTF-8")
			}
		case "/api/cluster":
			if sg.Configuration.HTTP.Enabled {
				clusterData := make(map[string]map[int32]*ShardGroup)
				for key, mg := range sg.Managers {
					clusterData[key] = mg.ShardGroups
				}
				res, err = json.Marshal(RestResponse{true, clusterData, nil})
			} else {
				res, err = json.Marshal(RestResponse{false, "HTTP Interface is not enabled", nil})
			}

			if err == nil {
				ctx.Write(res)
				ctx.Response.Header.Set("content-type", "application/javascript;charset=UTF-8")
			}
		case "/api/analytics":
			if sg.Configuration.HTTP.Enabled {

				clusters := make([]ClusterInformation, 0, len(sg.Managers))
				guilds := make(map[string]int64)
				for _, mg := range sg.Managers {
					statuses := make(map[int32]structs.ShardGroupStatus)

					mg.ShardGroupMu.Lock()
					for i, sg := range mg.ShardGroups {
						statuses[i] = sg.Status
					}
					mg.ShardGroupMu.Unlock()

					guildCount, err := sg.RedisClient.HLen(context.Background(), mg.CreateKey("guilds")).Result()
					guilds[mg.Configuration.Caching.RedisPrefix] = guildCount
					if err != nil {
						mg.Logger.Error().Err(err).Msg("Failed to retrieve Hashset Length")
					}

					clusters = append(clusters, ClusterInformation{
						Name:      mg.Configuration.DisplayName,
						Guilds:    guildCount,
						Status:    statuses,
						AutoStart: mg.Configuration.AutoStart,
					})
				}

				now := time.Now()
				guildCount := int64(0)
				for _, count := range guilds {
					guildCount += count
				}

				response := AnalyticResponse{
					Graph:    sg.ConstructAnalytics(),
					Guilds:   guildCount,
					Uptime:   now.Sub(sg.Start).Round(time.Second).String(),
					Events:   atomic.LoadInt64(sg.TotalEvents),
					Clusters: clusters,
				}

				res, err = json.Marshal(RestResponse{true, response, nil})
			} else {
				res, err = json.Marshal(RestResponse{false, "HTTP Interface is not enabled", nil})
			}

			if err == nil {
				ctx.Write(res)
				ctx.Response.Header.Set("content-type", "application/javascript;charset=UTF-8")
			}
		case "/api/rpc":
			rpcMessage := RPCRequest{}
			err = json.Unmarshal(ctx.PostBody(), &rpcMessage)

			if err == nil {
				sg.Logger.Debug().Str("method", rpcMessage.Method).Str("params", string(rpcMessage.Params)).Str("id", rpcMessage.ID).Msg("Received RPC request")
				if sg.Configuration.HTTP.Enabled {

					switch rpcMessage.Method {
					case "shardgroup:create":
						shardGroupCreateEvent := struct {
							AutoIDs          bool   `json:"autoIDs"`
							AutoShard        bool   `json:"autoShard"`
							Cluster          string `json:"cluster"`
							ShardCount       int    `json:"shardCount"`
							RawShardIDs      string `json:"shardIDs"`
							ShardIDs         []int  `json:"FinalShardIDs"`
							StartImmediately bool   `json:"startImmediately"`
						}{}

						json.Unmarshal(rpcMessage.Params, &shardGroupCreateEvent)

						// Check if cluster exists
						if mg, ok := sg.Managers[shardGroupCreateEvent.Cluster]; ok {
							// Auto Shards
							if shardGroupCreateEvent.AutoShard {
								mg.GatewayMu.Lock()
								gw, err := mg.GetGateway()
								if err != nil {
									mg.Logger.Warn().Err(err).Msg("Received error retrieving gateway object. Using old response.")
								} else {
									// We will only overwrite the gateway if it does not error as we
									// will just recycle the old response.
									mg.Gateway = gw
								}
								// shardGroupCreateEvent.ShardCount = mg.GatherShardCount()
								shardGroupCreateEvent.ShardCount = mg.Gateway.Shards
								mg.GatewayMu.Unlock()
							}
							if shardGroupCreateEvent.ShardCount < 1 {
								sg.Logger.Debug().Msg("Set ShardCount to 1 as it was less than 1")
								shardGroupCreateEvent.ShardCount = 1
							}

							if shardGroupCreateEvent.AutoIDs {
								shardGroupCreateEvent.ShardIDs = mg.GenerateShardIDs(shardGroupCreateEvent.ShardCount)
							} else {
								shardGroupCreateEvent.ShardIDs = returnRange(shardGroupCreateEvent.RawShardIDs, shardGroupCreateEvent.ShardCount)
							}

							sg.Logger.Debug().Msgf("Created ShardIDs: %v", shardGroupCreateEvent.ShardIDs)

							if len(shardGroupCreateEvent.ShardIDs) == 0 {
								sg.Logger.Debug().Msg("Set ShardIDs to [0] as it was empty")
								shardGroupCreateEvent.ShardIDs = []int{0}
							}

							if len(shardGroupCreateEvent.ShardIDs) > shardGroupCreateEvent.ShardCount {
								sg.Logger.Warn().Msgf("Length of ShardIDs is larger than the ShardCount %d > %d", len(shardGroupCreateEvent.ShardIDs), shardGroupCreateEvent.ShardCount)
								// TODO: We should handle this properly but it will error out when it starts up anyway
							}

							mg.GatewayMu.RLock()
							if len(shardGroupCreateEvent.ShardIDs) < mg.Gateway.SessionStartLimit.Remaining {
								mg.Scale(shardGroupCreateEvent.ShardIDs, shardGroupCreateEvent.ShardCount, true)
								res, err = json.Marshal(RPCResponse{true, "", rpcMessage.ID})
							} else {
								res, err = json.Marshal(RPCResponse{nil, xerrors.Errorf("Not enough sessions to start %d shards. %d remain", len(shardGroupCreateEvent.ShardIDs), mg.Gateway.SessionStartLimit.Remaining).Error(), rpcMessage.ID})
							}
							mg.GatewayMu.RUnlock()
						} else {
							res, err = json.Marshal(RPCResponse{nil, xerrors.New("Invalid Cluster provided").Error(), rpcMessage.ID})
						}
					case "shardgroup:stop":
						shardGroupStopEvent := struct {
							Cluster    string `json:"cluster"`
							ShardGroup int32  `json:"shardgroup"`
						}{}

						json.Unmarshal(rpcMessage.Params, &shardGroupStopEvent)

						// Check if cluster exists
						if mg, ok := sg.Managers[shardGroupStopEvent.Cluster]; ok {
							mg.ShardGroupMu.Lock()
							if sg, ok := mg.ShardGroups[shardGroupStopEvent.ShardGroup]; ok {
								sg.Close()
								res, err = json.Marshal(RPCResponse{true, "", rpcMessage.ID})
							} else {
								res, err = json.Marshal(RPCResponse{nil, xerrors.New("Invalid ShardGroup provided").Error(), rpcMessage.ID})
							}
							mg.ShardGroupMu.Unlock()
						} else {
							res, err = json.Marshal(RPCResponse{nil, xerrors.New("Invalid Cluster provided").Error(), rpcMessage.ID})
						}
					case "manager:refresh_gateway":
						manageRefreshGatewayEvent := struct {
							Cluster string `json:"cluster"`
						}{}

						json.Unmarshal(rpcMessage.Params, &manageRefreshGatewayEvent)

						// Check if cluster exists
						if mg, ok := sg.Managers[manageRefreshGatewayEvent.Cluster]; ok {
							gw, err := mg.GetGateway()
							if err == nil {
								mg.GatewayMu.Lock()
								mg.Gateway = gw
								mg.GatewayMu.Unlock()
								res, err = json.Marshal(RPCResponse{gw, "", rpcMessage.ID})
							} else {
								res, err = json.Marshal(RPCResponse{nil, err.Error(), rpcMessage.ID})
							}
						} else {
							res, err = json.Marshal(RPCResponse{nil, xerrors.New("Invalid Cluster provided").Error(), rpcMessage.ID})
						}
					case "shardgroup:delete":
						shardGroupStopEvent := struct {
							Cluster    string `json:"cluster"`
							ShardGroup int32  `json:"shardgroup"`
						}{}

						json.Unmarshal(rpcMessage.Params, &shardGroupStopEvent)

						// Check if cluster exists
						if mg, ok := sg.Managers[shardGroupStopEvent.Cluster]; ok {
							mg.ShardGroupMu.Lock()
							if sg, ok := mg.ShardGroups[shardGroupStopEvent.ShardGroup]; ok {
								if sg.Status == structs.ShardGroupClosed {
									delete(mg.ShardGroups, shardGroupStopEvent.ShardGroup)
									res, err = json.Marshal(RPCResponse{true, "", rpcMessage.ID})
								} else {
									res, err = json.Marshal(RPCResponse{nil, "ShardGroup has not closed", rpcMessage.ID})
								}
							} else {
								res, err = json.Marshal(RPCResponse{nil, xerrors.New("Invalid ShardGroup provided").Error(), rpcMessage.ID})
							}
							mg.ShardGroupMu.Unlock()
						} else {
							res, err = json.Marshal(RPCResponse{nil, xerrors.New("Invalid Cluster provided").Error(), rpcMessage.ID})
						}
					default:
						res, err = json.Marshal(RPCResponse{nil, xerrors.Errorf("Unknown event: %s", rpcMessage.Method).Error(), rpcMessage.ID})
					}
				} else {
					res, err = json.Marshal(RPCResponse{nil, xerrors.New("HTTP Interface is not enabled").Error(), rpcMessage.ID})
				}
			} else {
				sg.Logger.Error().Err(err).Msg("Failed to unmarshal RPC request")
				res, err = json.Marshal(RPCResponse{nil, xerrors.New("Invalid RPC Payload").Error(), ""})
			}

			ctx.Write(res)
			ctx.Response.Header.Set("content-type", "application/javascript;charset=UTF-8")
		default:
			ctx.SetStatusCode(404)
		}
	}

	if err != nil {
		sg.Logger.Warn().Err(err).Msg("Failed to process request")

		if res, err = json.Marshal(RestResponse{false, nil, err}); err == nil {
			ctx.Write(res)
			ctx.Response.Header.Set("content-type", "application/javascript;charset=UTF-8")
		}
		ctx.SetStatusCode(500)
	}

	// GET /managers - lists all managers
	// GET /manager/<> - gets details on manager such as shardgroup, shards and status
	// GET /manager/<>/shards - gets more detailed info on each shard and shardgroups

	// PUT /managers - creates a manager
	// PUT /manager/<>/shardgroup - creates a new shard group

	// POST /manager/<> - update config
	// POST /manager/<>/signal - change status such as turn off and on

}

// ConstructAnalytics returns a LineChart struct based off of manager analytics
func (sg *Sandwich) ConstructAnalytics() LineChart {
	datasets := make([]Dataset, 0, len(sg.Managers))

	mankeys := make([]string, 0, len(sg.Managers))
	for key := range sg.Managers {
		mankeys = append(mankeys, key)
	}
	sort.Strings(mankeys)

	for i, ident := range mankeys {
		mg := sg.Managers[ident]
		if mg.Analytics == nil {
			continue
		}

		data := make([]interface{}, 0, len(mg.Analytics.Samples))

		for _, sample := range mg.Analytics.Samples {
			data = append(data, DataStamp{sample.StoredAt, sample.Value})
		}

		colour := colours[i%len(colours)]
		datasets = append(datasets, Dataset{
			Label:            mg.Configuration.DisplayName,
			BackgroundColour: colour[0],
			BorderColour:     colour[1],
			Data:             data,
		})
	}

	return LineChart{
		Datasets: datasets,
	}
}

func (sg *Sandwich) handleRequests() {
	// if !sg.Configuration.HTTP.Enabled {
	// 	return
	// }

	for {
		sg.Logger.Info().Msgf("Running HTTP server at %s", sg.Configuration.HTTP.Host)
		err := fasthttp.ListenAndServe(sg.Configuration.HTTP.Host, sg.HandleRequest)
		sg.Logger.Error().Err(err).Msg("Error occured whilst running fasthttp")
	}
}

// Converts a string like 0-4,6-7 to [0,1,2,3,4,6,7]
func returnRange(_range string, max int) (result []int) {
	for _, split := range strings.Split(_range, ",") {
		ranges := strings.Split(split, "-")
		if low, err := strconv.Atoi(ranges[0]); err == nil {
			if hi, err := strconv.Atoi(ranges[len(ranges)-1]); err == nil {
				for i := low; i < hi+1; i++ {
					if 0 < i && i < max {
						result = append(result, i)
					}
				}
			}
		}
	}
	return result
}

// RPCRequest is the structure the client sends when an JSON-RPC call is made
type RPCRequest struct {
	Method string              `json:"method"`
	Params jsoniter.RawMessage `json:"params"`
	ID     string              `json:"id"`
}

// RPCResponse is the structure the server sends to respond to a JSON-RPC request
type RPCResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
	ID     string      `json:"id"`
}

// AnalyticResponse is the analytic response when you request the analytics
type AnalyticResponse struct {
	Graph    LineChart            `json:"chart"`
	Guilds   int64                `json:"guilds"`
	Uptime   string               `json:"uptime"`
	Events   int64                `json:"events"`
	Clusters []ClusterInformation `json:"clusters"`
}

// ClusterInformation is cluster information...
type ClusterInformation struct {
	Name      string                             `json:"name"`
	Guilds    int64                              `json:"guilds"`
	Status    map[int32]structs.ShardGroupStatus `json:"status"`
	AutoStart bool                               `json:"autostart"`
}

// DataStamp is a struct to store a time and a corresponding value
type DataStamp struct {
	Time  interface{} `json:"x"`
	Value interface{} `json:"y"`
}

// LineChart is a struct to store LineChart data easier
type LineChart struct {
	Labels   []string  `json:"labels,omitempty"`
	Datasets []Dataset `json:"datasets"`
}

// Dataset is a struct to store data for a Chart
type Dataset struct {
	Label            string        `json:"label"`
	BackgroundColour string        `json:"backgroundColor,omitempty"`
	BorderColour     string        `json:"borderColor,omitempty"`
	Data             []interface{} `json:"data"`
}
