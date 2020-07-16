let sock;
let sockConnected = false;
let sockUrl = "ws://" + window.location.host + "/api/sock/";
let sockRetryInterval = 2000;
let sockMessageHandler = {};
let sockTypeField = "Type";

function SockOpen(id, onCloseCallback) {
    sock = new WebSocket(sockUrl + id);
    sock.onopen = function(e) {
        sockConnected = true;
    };
    sock.onmessage = function(e) {
        e = JSON.parse(e.data);
        // console.log(e);
        if (sockMessageHandler[e[sockTypeField]]) {
            sockMessageHandler[e[sockTypeField]](e);
        }
    };
    sock.onclose = function(e) {
        sockConnected = false;
        if (onCloseCallback) {
            onCloseCallback();
        }
        setTimeout(SockOpen, sockRetryInterval, id, onCloseCallback);
    };
}

function SockSendMessage(message) {
    if (!sockConnected) {
        return false;
    }
    if (typeof(message) !== "string") {
        message = JSON.stringify(message);
    }
    sock.send(message);
    return true;
}
