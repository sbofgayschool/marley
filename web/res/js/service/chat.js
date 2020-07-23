let chatType = "chat";
let chatNumQueryInterval = 5000;
let chatLive;
let chatRecorder = null;

let chat = new Vue({
    el: "#divChat",
    data: {
        autoScroll: true,
        messages: [],
        newMessage: {
            MsgType: "text",
            Message: "",
            Source: ""
        },
        recorderStatus: 0
    },
    methods: {
        SendMessage: function(e) {
            if (!this.newMessage.Message) {
                return;
            }
            ChatSendMessage(this.newMessage);
            this.newMessage.Message = "";
        },
        Record: function () {
            if (chatRecorder === null) {
                return;
            }
            if (this.recorderStatus === 0) {
                this.recorderStatus = 1;
                chatRecorder.record();
            } else if (this.recorderStatus === 1) {
                this.recorderStatus = 2;
                chatRecorder.stop();
                chatRecorder.exportWAV(function (blob) {
                    chat.$data.recorderStatus = 0;
                    chatRecorder.clear();
                    let form = new FormData();
                    form.append("file", blob, "record.wav");
                    Ajax(Api("file"), null, form, "POST", function(data) {
                        ChatSendMessage({
                            "MsgType": data["Type"],
                            "Message": "",
                            "Source": data["File"]
                        });
                    }, null, false);
                })
            }
        }
    },
    watch: {
        messages: function () {
            this.$nextTick(function() {
                if (chatLive) {
                    if (!this.autoScroll) {
                        return;
                    }
                    $("#divChatMessage").scrollTop($("#divChatMessage")[0].scrollHeight);
                } else {
                    // TODO: Scroll the message div to correct position.
                }
            });
        }
    }
});

function ChatLoadMessage(messages) {
    if (!messages || messages.length === 0 || (chat.$data.messages.length > 0 && chat.$data.messages[0].ElapsedTime < messages[0].ElapsedTime)) {
        return;
    }
    chat.$data.messages = [];
    for (let i in messages) {
        chat.$data.messages.push({
            Username: messages[i]["Username"],
            MsgType: messages[i]["MsgType"],
            Message: messages[i]["Message"],
            Source: messages[i]["Source"],
            ElapsedTime: messages[i]["ElapsedTime"],
            DisplayTime: new Date(messages[i]["ElapsedTime"]).toUTCString()
        });
    }
}

function ChatOnMessageHandler(msg) {
    if (msg["Operation"] === "message") {
        if (chatLive) {
            chat.$data.messages.push({
                Username: msg["Username"],
                MsgType: msg["MsgType"],
                Message: msg["Message"],
                Source: msg["Source"],
                ElapsedTime: msg["ElapsedTime"],
                DisplayTime: new Date(msg["ElapsedTime"]).toUTCString()
            });
        } else {
            // TODO: find the correct place and insert the message.
        }
    } else if (msg["Operation"] === "numQuery") {
        $("#spanAudience").text("Audience: " + msg["Num"])
    }
}

sockMessageHandler[chatType] = ChatOnMessageHandler;

function ChatSendMessage(msg) {
    let newMessage = msg;
    newMessage["Type"] = "chat";
    newMessage["Operation"] = "message";
    if (!chatLive) {

    }
    SockSendMessage(newMessage);
}

function ChatUploadFile(confirm) {
    if (!confirm) {
        $("#inputFile").val("");
        $("#dlgFileUpload").modal("hide");
        return;
    }
    Ajax(Api("file"), null, new FormData($("#formFile")[0]), "POST", function(data) {
        console.log(data);
        if ($("#inputFileName").val() === "") {
            $("#inputFileName").val(data["File"]);
        }
        ChatSendMessage({
            "MsgType": data["Type"],
            "Message": $("#inputFileName").val(),
            "Source": data["File"]
        });
        $("#inputFile").val("");
        $("#inputFileName").val("");
        $("#dlgFileUpload").modal("hide");
    }, null, false);
}

// TODO: Vod Progress check and handler

$(function() {
    navigator.mediaDevices.getUserMedia({video: false, audio: true}).then(
        function(stream) {
            let audioContext = new AudioContext();

            let input = audioContext.createMediaStreamSource(stream);

            chatRecorder = new Recorder(input,{numChannels: 1});
        }
    );
    setInterval(SockSendMessage, chatNumQueryInterval, {"Type": chatType, "Operation": "numQuery"});
});