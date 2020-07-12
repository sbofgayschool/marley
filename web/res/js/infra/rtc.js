let rtcPC;
let rtcStatus = 0;

function RTCStartBroadcast(rawTracks, quality, sp, loadCallback, failedCallback) {
    if (rtcStatus !== 0) {
        return false;
    }
    rtcStatus = 1;
    rtcPC = new RTCPeerConnection(null);
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
        console.log(e);
        if (rtcPC.iceConnectionState === "connected") {
            rtcStatus = 2;
        } else if (rtcPC.iceConnectionState === "failed" || rtcPC.iceConnectionState === "closed") {
            if (rtcStatus === 0) {
                return;
            }
            RTCStopConnection();
            failedCallback();
        }
    };
    rtcPC.onicecandidate = function(event) {
        console.log(event);
        if (event.candidate === null) {
            loadCallback(rtcPC.localDescription, tracks);
        } else {
            console.log(event.candidate);
        }
    };
    rtcPC.createOffer().then(function(sdp) {
        rtcPC.setLocalDescription(sdp);
        console.log(sdp);
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
