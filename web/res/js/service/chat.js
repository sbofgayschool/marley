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
                    // Currently nothing will be done when new message is inserted in Vod mode.
                }
            });
        }
    }
});

function ChatConvertTime(time) {
    if (chatLive) {
        return new Date(time).toUTCString();
    } else {
        let minutes = Math.floor(time / 60000);
        let seconds = Math.floor(((time % 60000) / 1000));
        return minutes + ":" + (seconds < 10 ? '0' : '') + seconds;
    }
}

function ChatLoadMessage(messages) {
    if (chatLive && (!messages || messages.length === 0 || (chat.$data.messages.length > 0 && chat.$data.messages[0].ElapsedTime < messages[0].ElapsedTime))) {
        return;
    }
    chat.$data.messages = [];
    for (let i in messages) {
        let message = {
            Username: messages[i]["Username"],
            MsgType: messages[i]["MsgType"],
            Message: messages[i]["Message"],
            Source: messages[i]["Source"],
            ElapsedTime: messages[i]["ElapsedTime"],
            DisplayTime: ChatConvertTime(messages[i]["ElapsedTime"])
        }
        if (!chatLive) {
            message.Id = "msg_" + chat.$data.messages.length;
        }
        chat.$data.messages.push(message);
    }
    if (!chatLive) {
        setInterval(ChatVodProgress, chatVodProgressInterval);
    }
}

function ChatSearchPosition(elapsedTime) {
    let l = 0, r = chat.$data.messages.length;
    while (l < r) {
        let mid = l + Math.floor((r - l) / 2);
        if (chat.$data.messages[mid].ElapsedTime < elapsedTime) {
            l = mid + 1;
        } else {
            r = mid;
        }
    }
    return l;
}

function ChatOnMessageHandler(msg) {
    if (msg["Operation"] === "message") {
        let message = {
            Username: msg["Username"],
            MsgType: msg["MsgType"],
            Message: msg["Message"],
            Source: msg["Source"],
            ElapsedTime: msg["ElapsedTime"],
            DisplayTime: ChatConvertTime(msg["ElapsedTime"])
        };
        if (!chatLive) {
            message.Id = "msg_" + chat.$data.messages.length;
        }
        if (chatLive) {
            chat.$data.messages.push(message);
        } else {
            chat.$data.messages.splice(ChatSearchPosition(message.ElapsedTime), 0, message);
        }
    } else if (msg["Operation"] === "numQuery") {
        $("#spanAudience").text("Audience: " + msg["Num"]);
    }
}

sockMessageHandler[chatType] = ChatOnMessageHandler;

function ChatSendMessage(msg) {
    let newMessage = msg;
    newMessage["Type"] = "chat";
    newMessage["Operation"] = "message";
    if (!chatLive) {
        newMessage["ElapsedTime"] = GetSourceTime();
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

let chatVodProgressInterval = 500;
let chatPrvId = "";

function ChatVodProgress() {
    if (!chat.$data.autoScroll || vodSource[0].paused) {
        return;
    }
    let i = ChatSearchPosition(GetSourceTime());
    if (i > 0) {
        i--;
    }
    if (i < chat.$data.messages.length) {
        if (chatPrvId !== chat.$data.messages[i].Id) {
            $("#divChatMessage").scrollTop($("#" + chat.$data.messages[i].Id)[0].offsetTop);
        }
        chatPrvId = chat.$data.messages[i].Id;
    }
}

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