var app = new Vue({
    el: "#game",
    directives: {
        focus: {
            inserted: function (el) {
                el.focus();
            },
        },
    },
    data: {
        socket: null,
        game: null,
        settings: false,
        playerName: "",
        playerNameErr: "",
        code: "",
        joinLoading: false,
        error: null,
    },
    computed: {
        leader: function () {
            if (this.game) {
                return this.game.leader.name == this.playerName;
            }
            return false;
        },
        canStart: function () {
            if (!this.game) { return false; }
            return this.game.team1.players.length > 1 && this.game.team2.players.length > 1;
        },
    },
    methods: {
        receive: function (msg) {
            switch (msg.type) {
                case "state":
                    this.joinLoading = false;
                    this.game = msg.data;
                    break;
                case "error":
                    this.error = msg.data;
                case "ping":
                    this.socket.send({ type: "pong" });
                    break;
            }
        },
        join: function () {
            this.playerNameErr = "";
            if (!this.playerName) {
                this.playerNameErr = "You must provide a name before joining";
                return;
            }
            this.joinLoading = true;
            this.socket.send({
                type: "join",
                data: {
                    code: this.code,
                    name: this.playerName,
                }
            });
            localStorage.setItem("playerName", this.playerName);
        },
        send: function (type, data) {
            this.socket.send({ type: type, data: data });
        },
        setNamesPerPlayer: function (increment) {
            this.game.namesPerPlayer += increment;
            if (this.game.namesPerPlayer <= 0) {
                this.game.namesPerPlayer = 1;
            } else if (this.game.namesPerPlayer >= 20) {
                this.game.namesPerPlayer = 20;
            }
            this.socket.send({ type: "namesperplayer", data: this.game.namesPerPlayer });
        },
    },
    mounted: async function () {
        this.code = document.getElementById("game").getAttribute("data-code");
        this.socket = GameSocket(this.receive);
        await this.socket.connect();
        this.playerName = localStorage.getItem("playerName");
    },
})

function GameSocket(onmessage) {
    const retryPoll = 3000;
    let url = window.location.origin.toString().replace("http://", "ws://").replace("https://", "wss://") + "/game";
    return {
        connect() {
            return new Promise((resolve, reject) => {
                this.connection = new WebSocket(url);
                this.connection.onopen = () => {
                    this.manualClose = false;
                    this.connection.onmessage = (event) => {
                        onmessage(JSON.parse(event.data));
                    };
                    this.connection.onerror = (event) => {
                        console.log("Web Socket error, retrying: ", event);
                        this.retry();
                    };
                    // will always retry closed connections until a message is sent from the server to
                    // for the client to close the connection themselves.
                    this.connection.onclose = () => {
                        if (this.manualClose) {
                            return;
                        }
                        this.retry();
                    };
                    resolve();
                };

                this.connection.onerror = (event) => {
                    reject(event);
                };
            });
        },
        async send(data) {
            if (!this.connection || this.connection.readyState !== WebSocket.OPEN) {
                await this.connect();
            }

            let msg;
            if (typeof data === "string" || data instanceof ArrayBuffer || data instanceof Blob) {
                msg = data;
            } else {
                msg = JSON.stringify(data);
            }

            this.connection.send(msg);
        },
        close(code, reason) {
            if (this.connection) {
                this.manualClose = true;
                this.connection.close(code, reason);
            }
        },
        retry() {
            setTimeout(async () => {
                try {
                    await this.connect();
                } catch (err) {
                    console.log("Web Socket Errored, retrying: ", err);
                    this.retry();
                }
            }, this.retryPoll);
        },
    };
}