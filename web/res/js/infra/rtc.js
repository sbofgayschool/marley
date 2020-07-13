let rtcPC;
let rtcStatus = 0;

function RTCStartBroadcast(rawTracks, quality, sp, loadCallback, failedCallback) {
    if (rtcStatus !== 0) {
        return false;
    }
    rtcStatus = 1;
    rtcPC = new RTCPeerConnection(null);
    let readyCount = 0;
    let tracks = [];
    for (let i = 0; i <= quality; i++) {
        if (sp) {
            tracks.push([rawTracks[i][0].id]);
            rtcPC.addTrack(rawTracks[i][0]);
        } else {
            tracks.push([rawTracks[i][0].id, rawTracks[i][1].id]);
            rtcPC.addTrack(rawTracks[i][0]);
            rtcPC.addTrack(rawTracks[i][1]);
        }
    }
    rtcPC.oniceconnectionstatechange = function(e) {
        console.log("iCE state changed")
        console.log(e);
    };
    rtcPC.onconnectionstatechange = function(e) {
        console.log("connection state changed")
        console.log(e);
        if (rtcPC.connectionState === "connected") {
            rtcStatus = 2;
        } else if (rtcPC.connectionState === "failed" || rtcPC.connectionState === "closed") {
            if (rtcStatus === 0) {
                return;
            }
            RTCStopConnection();
            failedCallback();
        }
    };
    rtcPC.onicecandidate = function(event) {
        if (event.candidate === null) {
            console.log("ice gathered");
            readyCount++;
            if (readyCount === 2) {
                loadCallback(rtcPC.localDescription, tracks);
            }
        } else {
            console.log(event.candidate);
        }
    };
    rtcPC.createOffer().then(function(sdp) {
        rtcPC.setLocalDescription(sdp).then(function () {
            console.log("offer set");
            readyCount++;
            if (readyCount === 2) {
                loadCallback(rtcPC.localDescription, tracks);
            }
        });
    });
    return true;
}

function RTCStopConnection() {
    if (rtcStatus === 0) {
        return;
    }
    rtcStatus = 0;
    rtcPC.close();
    rtcPC = null;
}
