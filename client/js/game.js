var app = new Vue({
    el: "#game",
    data: {
        socket: null,
    },
    methods: {
        receive: function (data) {
            console.log("data: ", data);
        },
        send: function (data) {
            this.socket.send(data);
        },
    },
    mounted: async function () {
        this.socket = GameSocket(this.receive);
        await this.socket.connect();
    },
})

function Game(code) {

}

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