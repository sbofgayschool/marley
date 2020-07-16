let rtcPC;
let rtcStatus = 0;

function RTCStart(broadcast, rawTracks, quality, sp, loadCallback, failedCallback, onTrackCallback) {
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
            if (sp) {
                tracks.push([rawTracks[i][0].id]);
                curPC.addTrack(rawTracks[i][0]);
            } else {
                tracks.push([rawTracks[i][0].id, rawTracks[i][1].id]);
                curPC.addTrack(rawTracks[i][0]);
                curPC.addTrack(rawTracks[i][1]);
            }
        }
    } else {
        curPC.addTransceiver("audio");
        if (!sp) {
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
