let spBroadcast = null;

let spCanvasPdf;
let spContextPdf;
let spCanvasOpt;
let spContextOpt;

let spStartTime = null;
let spOpts = [];
let spPageOpts = [];
let spCurPage = -1;

let spFileUploadDir = "res/file/sp/";

function SpStart(pdf, opts, broadcast) {
    if (spStartTime) {
        return;
    }
    spBroadcast = broadcast;
    spOpts = opts;
    spStartTime = new Date();
    let curStartTime = spStartTime;
    spCurPage = 0;
}

function SpStop() {
    if (!spStartTime) {
        return;
    }
    SpReset();
    spBroadcast = null;
    spStartTime = null;
    spOpts = [];
    spCurPage = -1;
}

$(function () {
    spCanvasPdf = $("#canvasPdf")[0];
    spCanvasPdf.width = $("#divCanvasBody").width();
    spCanvasPdf.height = 400;
    spContextPdf = spCanvasPdf.getContext("2d");
    spCanvasOpt = $("#canvasOpt")[0];
    spCanvasOpt.width = $("#divCanvasBody").width();
    spCanvasOpt.height = 400;
    spContextOpt = spCanvasOpt.getContext("2d");
    spContextOpt.strokeStyle = "#FF0000";
    spContextOpt.lineWidth = 2;
});

function SpReset() {
    if (!spStartTime) {
        return;
    }
    spPageOpts = [];
    spCurPage = 0;
    spContextPdf.clearRect (0, 0, spCanvasPdf.width, spCanvasPdf.height);
    spContextOpt.clearRect (0, 0, spCanvasOpt.width, spCanvasOpt.height);
}

function SpChangePage(page, omitDraw) {
    if (!spStartTime || page === spCurPage) {
        return;
    }
    spCurPage = page;
    while (spPageOpts.length <= page) {
        spPageOpts.push([]);
    }
    if (spBroadcast) {
        spOpts.push([page]);
    }
    if (!omitDraw) {
        // TODO: Redraw the content of this page.
    }

}

function SpAct(opt, page, omitDraw) {
    if (!spStartTime) {
        return;
    }
    if (page !== spCurPage) {
        SpChangePage(page);
    }
    spPageOpts[page].push(opt[0], opt[1]);
    if (spBroadcast) {
        if (spOpts.length === 0) {
            spOpts.push([page]);
        }
        spOpts[spOpts.length - 1].push(opt[0], opt[1]);
    }
    if (!omitDraw) {
        // TODO: Draw the line.
    }
}

function SpRedraw(clear, elapsedTime) {
    if (!spStartTime) {
        return;
    }
    if (clear) {
        SpReset();
    }
    let opts = spOpts;
    if (elapsedTime >= 0) {

    }
    for (let i = 0; i < opts.length; i++) {
        SpChangePage(opts[i][0], i === opts.length - 1);
        for (let j = 1; j < opts[i].length; j += 2) {
            SpAct([opts[i][j], opts[i][j + 1]], opts[i][0], i === opts.length - 1);
        }
    }
}