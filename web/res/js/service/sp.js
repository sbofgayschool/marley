let spBroadcast = null;

let spCanvasPdf;
let spContextPdf;
let spCanvasOpt;
let spContextOpt;

let spStartTime = null;
let spOpts = [];
let spPageOpts = [];
let spCurPage = -1;
let spPdfObject = null;

let spFileUploadDir = "res/file/sp/";

let spBroadcastPrevCoord = null;
let spBroadcastPrevTime = null;

function SpStart(pdf, opts, broadcast) {
    if (spStartTime) {
        return;
    }
    spBroadcast = broadcast;
    spOpts = opts;
    spStartTime = new Date();
    let curStartTime = spStartTime;
    pdfjsLib.getDocument(pdf).promise.then(function(pdfObj) {
        if (curStartTime !== spStartTime) {
            return;
        }
        spPdfObject = pdfObj;
        SpRedraw(null);
    });
    if (broadcast) {
        spCanvasOpt.onmousedown = SpBroadcastMouseDown;
        spCanvasOpt.onmouseup = SpBroadcastMouseUp;
        spCanvasOpt.onmousemove = SpBroadcastMouseMove;
        // TODO: Start opt send timer
    }
}

function SpStop() {
    if (!spStartTime) {
        return;
    }
    SpReset();
    spBroadcast = null;
    spCanvasOpt.onmousedown = spCanvasOpt.onmouseup = spCanvasOpt.onmousemove = null;
    spStartTime = null;
    spOpts = [];
    spCurPage = -1;
    spPdfObject = null;
}

function SpReset() {
    if (!spStartTime) {
        return;
    }
    spPageOpts = [];
    spCurPage = 0;
    spContextPdf.clearRect (0, 0, spCanvasPdf.width, spCanvasPdf.height);
    spContextOpt.clearRect (0, 0, spCanvasOpt.width, spCanvasOpt.height);
}

function SpDrawPdf(page) {
    if (!spStartTime || !spPdfObject) {
        return;
    }
    spPdfObject.getPage(page + 1).then(function(p) {
        let scale = Math.min((spCanvasPdf.width / p.view[2]), (spCanvasPdf.height / p.view[3]));
        let viewport = p.getViewport({scale: scale});
        p.render({canvasContext: spContextPdf, viewport: viewport});
    });
}

function SpDrawLine(start, end) {
    spContextOpt.beginPath();
    spContextOpt.moveTo(start[0] * spCanvasOpt.width, start[1] * spCanvasOpt.height);
    spContextOpt.lineTo(end[0] * spCanvasOpt.width, end[1] * spCanvasOpt.height);
    spContextOpt.stroke();
}

function SpChangePage(page, omitDraw) {
    if (!spStartTime || page === spCurPage) {
        return;
    }
    if (spBroadcast && (!spPdfObject || spPdfObject.numPages <= page)) {
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
        SpDrawPdf(page);
        for (let i = 0; i < spPageOpts[page]; i += 2) {
            SpDrawLine(spPageOpts[page][i], spPageOpts[page][i + 1]);
        }
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
        SpDrawLine(opt[0], opt[1]);
    }
}

function SpRedraw(elapsedTime) {
    if (!spStartTime) {
        return;
    }
    SpReset();
    SpDrawPdf(0);
    spPageOpts.push([]);
    let opts = spOpts;
    if (elapsedTime >= 0) {
        // TODO: Select options according to elapsed time.
    }
    for (let i = 0; i < opts.length; i++) {
        SpChangePage(opts[i][0], i !== opts.length - 1);
        for (let j = 1; j < opts[i].length; j += 2) {
            SpAct([opts[i][j], opts[i][j + 1]], opts[i][0], i !== opts.length - 1);
        }
    }
}

function SpBroadcastMouseDown(e) {
    if (!spStartTime || !spBroadcast || !spPdfObject) {
        return;
    }
    spBroadcastPrevTime = Date.now();
    spBroadcastPrevCoord = SpGenerateCoord(e);
}

function SpBroadcastMouseUp(e) {
    if (!spStartTime || !spBroadcast || !spPdfObject) {
        return;
    }
    spBroadcastPrevTime = null;
}

function SpBroadcastMouseMove(e) {
    if (!spStartTime || !spBroadcast || !spPdfObject || !spBroadcastPrevTime || Date.now() - spBroadcastPrevTime <= 70) {
        return;
    }
    let curCoord = SpGenerateCoord(e);
    SpAct([spBroadcastPrevCoord, curCoord], spCurPage, false);
    spBroadcastPrevTime = Date.now();
    spBroadcastPrevCoord = curCoord;
}

function SpGenerateCoord(e) {
    console.log(e.clientX, e.clientY);
    console.log(
        (e.clientX + document.body.scrollLeft + document.documentElement.scrollLeft - spCanvasOpt.offsetLeft) / spCanvasOpt.width,
        (e.clientY + document.body.scrollTop + document.documentElement.scrollTop - spCanvasOpt.offsetTop) / spCanvasOpt.height
    )
    return [
        (e.clientX + document.body.scrollLeft + document.documentElement.scrollLeft - spCanvasOpt.offsetLeft) / spCanvasOpt.width,
        (e.clientY + document.body.scrollTop + document.documentElement.scrollTop - spCanvasOpt.offsetTop) / spCanvasOpt.height
    ];
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