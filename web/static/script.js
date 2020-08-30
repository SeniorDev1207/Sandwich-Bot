// Sandwich-Daemon script.js by TehRockettek
// https://github.com/TheRockettek/Sandwich-Daemon

// Install any plugins
Vue.use(VueChartJs);

Vue.component("line-chart", {
    extends: VueChartJs.Line,
    mixins: [VueChartJs.mixins.reactiveProp],
    props: ['chartData', 'options'],
    mounted() {
        this.renderChart(this.chartData, this.options)
    },
})

Vue.component("card-display", {
    props: ['title', 'value', 'bg'],
    template: `
    <div class="col justify-content-center d-flex">
        <div :class="bg+' card text-white m-1'" style="width: 18rem;">
            <div class="card-header">{{ title }} </div>
            <div class="card-body">
                <h5 class="card-title">{{ value }}</h5>
            </div>
        </div>
    </div>
    `,
})

Vue.component("form-submit", {
    props: {
        label: {
            default: "Save Changes",
        }
    },
    template: `<button type="submit" class="btn btn-dark">{{ label }}</button>`,
})

Vue.component("form-input", {
    props: ['type', 'id', 'label', 'values', 'value'],
    template: `
    <div class="form-check" v-if="type == 'checkbox'">
        <input class="form-check-input" type="checkbox" :id="id" :checked="value" v-on:input="updateValue($event.target.checked)">
        <label class="form-check-label" :for="id">{{ label }}</label>
    </div>
    <div class="mb-3" v-else-if="type == 'text'">
        <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
        <input type="text" class="form-control" :id="id" :value="value" v-on:input="updateValue($event.target.value)">
    </div>
    <div class="mb-3" v-else-if="type == 'number'">
        <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
        <input type="number" class="form-control" :id="id" :value="value" v-on:input="updateValue($event.target.value)">
    </div>
    <div class="mb-3" v-else-if="type == 'password'">
        <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
        <div class="input-group">
            <input type="password" class="form-control" :id="id" autocomplete :value="value" v-on:input="updateValue($event.target.value)">
            <button class="btn btn-outline-dark" type="button">Copy Token</button>
        </div>
    </div>
    <div class="mb-3" v-else-if="type == 'select'">
        <label :for="id" class="col-sm-12 form-label">{{ label }}</label>
        <select class="form-select" :id="id" v-on:input="updateValue($event.target.value)">
            <option v-for="item in values" selected="item == value">{{ item }}</option>
        </select>
    </div>
    <div class="mb-3 row pb-4" v-else-if="type == 'intent'">
        <label for="managerBotIntents" class="col-sm-3 form-label">{{ label }}</label>
        <div class="col-sm-9">
            <input type="number" class="form-control" min=0 :value="value" @input="(v) => {updateValue(v.target.value); fromIntents(v.target.value)}">
            <div class="form-row py-2">
                <div class="form-check form-check-inline col-sm-8 col-md-5" v-for="(intent, index) in this.intents">
                    <input class="form-check-input" type="checkbox" v-bind:value="index" v-bind:id="'managerBotIntentBox'+index" v-model="selectedIntent" @change="calculateIntent()">
                    <label class="form-check-label" v-bind:for="'managerBotIntentBox'+index">{{intent}}</label>
                </div>
            </div>
        </div>
    </div>
    <div class="mb-3 row pb-4" v-else-if="type == 'presence'">
        <label class="col-sm-3 col-form-label">{{ label }}</label>
        <div class="col-sm-9">
            <div class="mb-3">
                <label :for="id + 'status'" class="col-sm-12 form-label">Status</label>
                <select class="form-select" :id="id + 'status'" :value="value.status" @input="(v) => {value.status = v.target.value}">
                    <option v-for="item in ['','online','dnd','idle','invisible','offline']" :key="item" :disabled="!item" :selected="item == value">{{ item }}</option>
                </select>
            </div>
            <div class="mb-3">
                <label :for="id + 'name'" class="col-sm-12 form-label">Name</label>
                <input type="text" class="form-control" :id="id + 'name'" :value="value.name" @input="(v) => {value.name = v.target.value}">
            </div>
            <div class="form-check">
                <input class="form-check-input" type="checkbox" :id="id + 'afk'" :checked="value.afk" @input="(v) => {value.afk = v.target.checked}"">
                <label class="form-check-label" :for="id + 'afk'">AFK</label>
            </div>
        </div>
    </div>
    <span class="badge bg-warning text-dark" v-else>Invalid type "{{ type }}" for "{{ id }}"</span>
    `,
    data: function () {
        return {
            "intents": [
                "GUILDS",
                "GUILD_MEMBERS",
                "GUILD_BANS",
                "GUILD_INTEGRATIONS",
                "GUILD_EMOJIS",
                "GUILD_WEBHOOKS",
                "GUILD_INVITES",
                "GUILD_VOICE_STATES",
                "GUILD_PRESENCES",
                "GUILD_MESSAGES",
                "GUILD_MESSAGE_REACTRIONS",
                "GUILD_MESSAGE_TYPING",
                "DIRECT_MESSAGES",
                "DIRECT_MESSAGE_REACTIONS",
                "DIRECT_MESSAGE_TYPING",
            ],
            "selectedIntent": []
        }
    },
    mounted: function () {
        if (this.type == "intent") {
            this.fromIntents(this.value);
        }
    },
    methods: {
        updateValue: function (value) {
            this.$emit('input', value)
        },
        calculateIntent() {
            this.intentValue = 0
            this.selectedIntent.forEach(a => { this.intentValue += (1 << a); })
            this.updateValue(this.intentValue)
        },
        fromIntents(val) {
            var _binary = Number(val).toString(2).split("").reverse()
            var _newIntent = [];
            _binary.forEach((value, index) => {
                if (value === "1" && this.selectedIntent.indexOf(value) === -1) {
                    _newIntent.push(index)
                };
            });
            this.selectedIntent = _newIntent;
        },
        updateValue: function (value) {
            this.$emit('input', value)
        }
    },
})

vue = new Vue({
    el: '#app',
    data() {
        return {
            loading: true,
            error: false,
            data: {},
            analytics: {
                chart: {},
                uptime: "...",
                visible: "...",
                events: "...",
                online: "...",
                colour: "bg-success",
            },
            loadingAnalytics: true,

            clusterConfiguration: {
                cluster: "",
                autoShard: true,
                shardCount: 1,
                autoIDs: true,
                shardIDs: "",
                startImmediately: true,
            },

            statusShard: ["Idle", "Waiting", "Connecting", "Connected", "Ready", "Reconnecting", "Closed", "Error"],
            colourShard: ["dark", "info", "info", "success", "success", "warn", "dark", "danger"],

            statusGroup: ["Idle", "Starting", "Connecting", "Ready", "Replaced", "Closing", "Closed"],

            colourCluster: ["dark", "info", "info", "success", "warn", "warn", "dark", "danger"],
        }
    },
    filters: {
        pretty: function (value) {
            return JSON.stringify(value, null, 2);
        }
    },
    mounted() {
        this.fetchConfiguration();
        this.fetchAnalytics();
        this.$nextTick(function () {
            window.setInterval(() => {
                this.fetchAnalytics();
            }, 15000);
        })
    },
    methods: {
        sendRPC(method, params, id) {
            axios
                .post('/api/rpc', {
                    'method': method,
                    'params': params,
                    'id': id,
                })
                .then(result => {
                    return result
                })
                .catch(err => console.log(error))
        },

        newShardGroup(cluster) {
            this.clusterConfiguration.cluster = cluster

            this.clusterConfiguration.autoShard = true
            this.clusterConfiguration.shardCount = 1
            this.clusterConfiguration.autoIDs = true
            this.clusterConfiguration.shardIDs = ""
            this.clusterConfiguration.startImmediately = true

            const modal = new bootstrap.Modal(document.getElementById("shardGroupModal"), {})
            modal.show()
        },
        createShardGroup() {
            const modal = new bootstrap.Modal(document.getElementById("shardGroupModal"), {})

            config = Object.assign({}, this.clusterConfiguration)
            console.log(this.sendRPC("shardgroup:create", config))

            modal.hide()
        },

        fetchConfiguration() {
            axios
                .get('/api/configuration')
                .then(result => { this.data = result.data.response; this.error = !result.data.success })
                .catch(error => console.log(error))
                .finally(() => this.loading = false)
        },
        fetchAnalytics() {
            axios
                .get('/api/analytics')
                .then(result => {
                    this.analytics = result.data.response;

                    let up = 0
                    let total = 0
                    let guilds = 0
                    this.analytics.colour = "bg-success";

                    clusters = Object.values(this.analytics.clusters)
                    for (mgindex in clusters) {
                        cluster = clusters[mgindex]
                        guilds += cluster.guilds
                        shardgroups = Object.values(cluster.status)
                        for (sgindex in shardgroups) {
                            shardgroupstatus = shardgroups[sgindex]
                            if (1 < shardgroupstatus && shardgroupstatus < 6) {
                                up++
                            }
                            total++
                        }
                    }

                    this.analytics.visible = guilds
                    this.analytics.online = up + "/" + total

                    this.error = this.error | !result.data.success;
                })
                .catch(error => console.log(error))
                .finally(() => this.loadingAnalytics = false)
        },
        fromClusters(clusters) {
            _clusters = {}
            Object.entries(clusters).forEach((item) => {
                key = item[0]
                value = item[1]

                shardgroups = Object.values(value.shard_groups)
                if (shardgroups.length == 0) {
                    status = 0
                } else {
                    status = shardgroups.slice(-1)[0].status
                }

                _clusters[key] = {
                    configuration: value.configuration,
                    shardgroups: value.shard_groups,
                    status: status,
                }
            })
            return _clusters
        },
        calculateAverage(cluster) {
            totalShards = 0;
            totalLatency = 0;

            shardgroups = Object.values(cluster.shardgroups)
            for (sgindex in shardgroups) {
                shardgroup = shardgroups[sgindex]
                if (shardgroup.status < 6) {
                    shards = Object.values(shardgroup.shards)
                    for (shindex in shards) {
                        shard = shards[shindex]
                        totalLatency = totalLatency + (new Date(shard.last_heartbeat_ack) - new Date(shard.last_heartbeat_sent))
                        totalShards = totalShards + 1
                    }
                }
            }
            return (totalLatency / totalShards) || '-'
        }
    }
})
