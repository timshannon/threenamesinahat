// Copyright 2020 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

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
        loading: false,
        error: null,
        addName: "",
        currentName: "",
        stealCheck: false,
        notification: "",
        startTurnReady: false,
        nameHints: [
            "Someone you're playing with",
            "A family member",
            "A singer",
            "A president",
            "An author",
            "A movie star",
            "The name of a tv character",
            "A news anchor",
            "A famous scientist",
            "A historical figure",
            "A character from a book",
            "A children's tv show character",
            "A daytime tv talk-show host",
            "A celebrity",
            "A pet's name",
            "A famous artist",
            "A director",
            "A famous chef",
        ],
    },
    computed: {
        leader: function () {
            if (this.game) {
                return this.game.leader.name === this.playerName;
            }
            return false;
        },
        canStart: function () {
            if (!this.game) { return false; }
            return this.game.team1.players.length > 1 && this.game.team2.players.length > 1;
        },
        player: function () {
            if (!this.game) { return null; }
            let res = this.game.team1.players.find(player => player.name === this.playerName);
            if (res) {
                return res;
            }

            res = this.game.team2.players.find(player => player.name === this.playerName);
            return res;
        },
        namesLeft: function () {
            if (this.game && this.player) {
                if (this.player.names) {
                    return this.game.namesPerPlayer - this.player.names.length;
                }
                return this.game.namesPerPlayer;
            }
            return 0;
        },
        timerPercent: function () {
            if (this.game && this.game.timer) {
                return (this.game.timer.left / this.game.timer.seconds) * 100;
            }
            return 0;
        },
        timerStyle: function () {
            if (this.game && this.game.timer) {
                let ratio = this.game.timer.left / this.game.timer.seconds;
                if (ratio < .25) {
                    return "danger";
                }
                if (ratio < .5) {
                    return "warning";
                }
            }
            return "success";
        },
        nameHint: function () {
            if (this.game && this.game.timer && this.game.stage === "setup") {
                // wait 5 seconds, then display name hints evenly across the remaining timer time
                const firstHintDelay = 5;
                let passed = Math.round(this.game.timer.seconds - this.game.timer.left);
                if (passed < firstHintDelay) {
                    return "";
                }

                let partition = (this.game.timer.seconds - firstHintDelay) / this.nameHints.length;
                let index = Math.round(this.game.timer.left / partition);

                return this.nameHints[index];
            }
            return "";
        },
        team: function () {
            if (!this.game) { return null; }
            let res = this.game.team1.players.find(player => player.name === this.playerName);
            if (res) {
                return 1;
            }
            return 2;
        },
        guessingTeam: function () {
            if (!this.game || !this.game.clueGiver) { return null; }
            let res = this.game.team1.players.find(player => player.name === this.game.clueGiver.name);
            if (res) {
                return 1;
            }
            return 2;
        },
        waitingTeam: function () {
            if (this.guessingTeam === 1) {
                return 2;
            }
            return 1;
        },
        isGuessing: function () {
            return this.team === this.guessingTeam;
        },
        isWaiting: function () {
            return this.team === this.waitingTeam;
        },
        isClueGiver: function () {
            if (!this.game || !this.game.clueGiver) { return null; }
            return this.game.clueGiver.name === this.playerName;
        },
        gameStarted: function () {
            if (!this.game) { return false; }
            return this.game.stage !== "pregame";
        },
    },
    methods: {
        receive: function (msg) {
            switch (msg.type) {
                case "state":
                    this.loading = false;
                    this.stateChange(msg.data, this.game);
                    this.game = msg.data;
                    break;
                case "error":
                    this.error = msg.data;
                    break;
                case "name":
                    this.currentName = msg.data;
                    break;
                case "stealcheck":
                    this.stealCheck = true;
                    break;
                case "ping":
                    this.socket.send({ type: "pong" });
                    break;
                case "playsound":
                    let sound = new Audio("/audio/" + msg.data + ".mp3");
                    sound.play();
                    break;
                case "startcheck":
                    this.startTurnReady = true;
                case "notification":
                    this.notification = msg.data;
                    break;
            }
        },
        join: function () {
            this.playerNameErr = "";
            if (!this.playerName) {
                this.playerNameErr = "You must provide a name before joining";
                return;
            }
            this.loading = true;
            this.socket.send({
                type: "join",
                data: {
                    code: this.code,
                    name: this.playerName,
                }
            });
            localStorage.setItem("playerName", this.playerName);
        },
        startGame: function () {
            this.loading = true;
            this.send("start");
        },
        submitName: function () {
            if (!this.addName) {
                return;
            }
            this.send("addname", this.addName);
            this.addName = "";
        },
        removeName: function (name) {
            this.send("removename", name);
        },
        send: function (type, data) {
            this.socket.send({ type: type, data: data });
        },
        startTurn: function () {
            this.send("startturn");
            this.startTurnReady = false;
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
        stealCheckConfirm: function (correct) {
            this.stealConfirm = false;
            this.currentName = "";
            if (correct) {
                this.send("stealyes");
            } else {
                this.send("stealno");
            }
        },
        reset: function () {
            this.send("reset");
        },
        stateChange: function (newState, oldState) {
            if (!newState || !oldState) { return; }

            if (oldState.stage !== newState.stage) {
                this.currentName = "";
                if (newState.stage === "setup") {
                    shuffle(this.nameHints);
                }
            }
        },
        reconnect: function () {
            if (this.playerName && this.gameStarted) {
                console.log("reconnect");
                this.join();
            }
        },
    },
    mounted: function () {
        this.playerName = localStorage.getItem("playerName");
        this.code = document.getElementById("game").getAttribute("data-code");
        this.socket = GameSocket(this.receive, this.reconnect);
        this.socket.connect()
            .catch((err) => {
                this.error = "Error connecting to the game server: " + err;
            });
    },
})

function GameSocket(onmessage, onreconnect) {
    const retryPoll = 1500;
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
        send(data) {
            if (!this.connection || this.connection.readyState !== WebSocket.OPEN) {
                setTimeout(() => {
                    this.send(data);
                }, retryPoll);
                return;
            }
            this.connection.send(JSON.stringify(data));
        },
        close(code, reason) {
            if (this.connection) {
                this.manualClose = true;
                this.connection.close(code, reason);
            }
        },
        retry() {
            setTimeout(() => {
                this.connect()
                    .then(() => {
                        onreconnect()
                    })
                    .catch((err) => {
                        console.log("Web Socket Errored, retrying: ", err);
                        this.retry();
                    });
            }, this.retryPoll);
        },
    };
}

function shuffle(a) {
    var j, x, i;
    for (i = a.length - 1; i > 0; i--) {
        j = Math.floor(Math.random() * (i + 1));
        x = a[i];
        a[i] = a[j];
        a[j] = x;
    }
    return a;
}