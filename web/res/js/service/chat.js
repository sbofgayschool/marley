let chatType = "chat";
let chatNumQueryInterval = 5000;
let chatLive;

let chat = new Vue({
    el: "#divChat",
    data: {
        autoScroll: true,
        messages: [],
        newMessage: {
            Message: "",
            MsgType: "text"
        }
    },
    methods: {
        SendMessage: function(e) {
            if (!this.newMessage.Message) {
                return;
            }
            ChatSendMessage(this.newMessage);
            this.newMessage.Message = "";
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

$(function() {
    setInterval(SockSendMessage, chatNumQueryInterval, {"Type": chatType, "Operation": "numQuery"});
});