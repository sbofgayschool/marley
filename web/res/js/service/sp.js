let spLiveStartTime = 0;
let spLiveScheduled = false;
let spLiveOpts = [];

function SpLiveStart(pdf, opts) {
    spLiveOpts = opts;
    spLiveStartTime = new Date().getTime();
    let curStartTime = spLiveStartTime;
    spLiveScheduled = true;
}

function SpLiveStop() {
    spLiveStartTime = 0;
    spLiveScheduled = false;
    spLiveOpts = [];
}
