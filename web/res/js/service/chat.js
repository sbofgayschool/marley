let chatType = "chat";
let chatNumQueryInterval = 5000;

let chat = new Vue({
    el: "#divChat",
    data: {
        scrollToEnd: true,
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
                if (!this.scrollToEnd) {
                    return;
                }
                let divChatMessage = $("#divChatMessage");
                divChatMessage.scrollTop(divChatMessage[0].scrollHeight);
            });
        }
    }
});

function ChatOnMessageHandler(msg) {
    if (msg["Operation"] === "message") {
        chat.$data.messages.push({
            Username: msg["Username"],
            MsgType: msg["MsgType"],
            Message: msg["Message"],
            ElapsedTime: new Date(msg["ElapsedTime"] * 1000).toUTCString()
        })
    } else if (msg["Operation"] === "numQuery") {
        $("#spanAudience").text("Audience: " + msg["Num"])
    }
}

sockMessageHandler[chatType] = ChatOnMessageHandler;

function ChatSendMessage(msg) {
    let newMessage = msg;
    newMessage["Type"] = "chat";
    newMessage["Operation"] = "message";
    SockSendMessage(newMessage);
}

function ChatSendNumQuery() {
    SockSendMessage({"Type": chatType, "Operation": "numQuery"});
}

$(function() {
    setInterval(ChatSendNumQuery, chatNumQueryInterval);
})