let rtcPC;
let rtcStatus = 0;

function RTCStart(broadcast, rawTracks, quality, loadCallback, failedCallback, onTrackCallback) {
    if (rtcStatus !== 0) {
        return false;
    }
    rtcStatus = 1;
    rtcPC = new RTCPeerConnection(null);
    let curPC = rtcPC;
    let readyCount = 0;
    let tracks = [];
    if (broadcast) {
        for (let i = 0; i <= quality; i++) {
            tracks.push(rawTracks[i].id);
            curPC.addTrack(rawTracks[i]);
        }
    } else {
        curPC.addTransceiver("audio");
        if (quality > 0) {
            curPC.addTransceiver("video");
        }
        curPC.ontrack = onTrackCallback;
    }
    let finishLoad = function () {
        readyCount++;
        if (readyCount === 2) {
            loadCallback(curPC.localDescription, tracks);
        }
    };
    curPC.oniceconnectionstatechange = function(e) {
        console.log("ice state changed to " + curPC.iceConnectionState);
    };
    curPC.onconnectionstatechange = function(e) {
        console.log("connection state changed to " + curPC.connectionState);
        if (curPC !== rtcPC) {
            return;
        }
        if (curPC.connectionState === "connected") {
            rtcStatus = 2;
        } else if (curPC.connectionState === "failed" || curPC.connectionState === "closed") {
            if (rtcStatus === 0) {
                return;
            }
            RTCStopConnection();
            failedCallback();
        }
    };
    curPC.onicecandidate = function(event) {
        if (event.candidate === null) {
            console.log("ice gathered");
            finishLoad();
        } else {
            console.log(event.candidate);
        }
    };
    curPC.createOffer().then(function(sdp) {
        curPC.setLocalDescription(sdp).then(function () {
            console.log("offer set");
            finishLoad();
        });
    });
    return true;
}

function RTCStopConnection() {
    if (rtcStatus === 0) {
        return;
    }
    rtcStatus = 0;
    let curPC = rtcPC;
    rtcPC = null;
    curPC.close();
}
