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
        code: "",
        codeErr: "",
    },
    methods: {
        join: function () {
            this.codeErr = "";
            if (this.code.length !== 4) {
                this.codeErr = "Game codes must be 4 letters long";
                return;
            }
            window.location = "/game/" + this.code;
        },
    }
});
