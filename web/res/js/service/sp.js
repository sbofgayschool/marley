let spBroadcast = undefined;

let spCanvasPdf;
let spContextPdf;
let spCanvasOpt;
let spContextOpt;

let spStartTime = null;
let spOpts = [];
let spPageOpts = [];
let spCurPage = -1;
let spPdfObject = null;
let spPdfDrawingPage = -1;

let spBroadcastPrevCoord = null;
let spBroadcastPrevTime = null;

let spBroadcastSendInterval = 125;
let spBroadcastOptInterval = 50;

function SpStart(pdf, opts, timestamp) {
    if (spStartTime) {
        return;
    }
    if (opts) {
        spOpts = opts;
    }
    spStartTime = new Date();
    let curStartTime = spStartTime;
    if (spBroadcast === true || spBroadcast === false) {
        SpRedraw(null);
    } else {
        SpRedraw(1);
    }
    pdfjsLib.getDocument(pdf).promise.then(function(pdfObj) {
        if (curStartTime !== spStartTime) {
            return;
        }
        spPdfObject = pdfObj;
        if (spBroadcast) {
            $("#spanSpNote").hide();
            $("#btnSpPagePrev").show();
            $("#btnSpClean").show();
            $("#btnSpPageNext").show();
        }
        SpDrawPdf(spCurPage);
    });
    if (spBroadcast) {
        spCanvasOpt.onmousedown = SpBroadcastMouseDown;
        spCanvasOpt.onmouseup = SpBroadcastMouseUp;
        spCanvasOpt.onmousemove = SpBroadcastMouseMove;
        setTimeout(function SendOpt() {
            if (curStartTime !== spStartTime) {
                return;
            }
            if (spOpts.length !== 0) {
                if (SockSendMessage({Type: liveType, Operation: "opt", Opt: JSON.stringify(spOpts), Timestamp: timestamp})) {
                    spOpts = [];
                }
            }
            setTimeout(SendOpt, spBroadcastSendInterval);
        }, spBroadcastSendInterval);
    }
}

function SpStop() {
    if (!spStartTime) {
        return;
    }
    SpReset();
    spCanvasOpt.onmousedown = spCanvasOpt.onmouseup = spCanvasOpt.onmousemove = null;
    spStartTime = null;
    spOpts = [];
    spCurPage = -1;
    spPdfObject = null;
    if (spBroadcast) {
        $("#spanSpNote").show();
        $("#btnSpPagePrev").hide();
        $("#btnSpClean").hide();
        $("#btnSpPageNext").hide();
    }
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
    if (spPdfDrawingPage === -1) {
        spPdfDrawingPage = page;
        console.log("Page Started:", page, spPdfDrawingPage);
        function draw() {
            console.log("Loading Page:", page, spPdfDrawingPage);
            spPdfObject.getPage(page + 1).then(function(p) {
                console.log("Page Loaded:", page, spPdfDrawingPage);
                if (page !== spPdfDrawingPage) {
                    page = spPdfDrawingPage;
                    draw();
                } else {
                    let scale = Math.min((spCanvasPdf.width / p.view[2]), (spCanvasPdf.height / p.view[3]));
                    let viewport = p.getViewport({scale: scale});
                    console.log("Drawing Page:", page, spPdfDrawingPage);
                    p.render({canvasContext: spContextPdf, viewport: viewport}).promise.then(function () {
                        console.log("Page Drawn:", page, spPdfDrawingPage);
                        if (page !== spPdfDrawingPage) {
                            page = spPdfDrawingPage;
                            draw();
                        } else {
                            spPdfDrawingPage = -1;
                            console.log("Page Stopped:", page, spPdfDrawingPage);
                        }
                    });
                }
            });
        }
        draw();
    } else {
        spPdfDrawingPage = page;
        console.log("Page Scheduled:", page, spPdfDrawingPage);
    }
}

function SpDrawLine(start, end) {
    if (!start && !end) {
        spContextOpt.clearRect (0, 0, spCanvasOpt.width, spCanvasOpt.height);
    } else {
        spContextOpt.beginPath();
        spContextOpt.moveTo(start[0] * spCanvasOpt.width, start[1] * spCanvasOpt.height);
        spContextOpt.lineTo(end[0] * spCanvasOpt.width, end[1] * spCanvasOpt.height);
        spContextOpt.stroke();
    }
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
        SpDrawPdf(page);
        spContextOpt.clearRect (0, 0, spCanvasOpt.width, spCanvasOpt.height);
        for (let i = 0; i < spPageOpts[page].length; i += 2) {
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
    spPageOpts[page].push(opt[0]);
    spPageOpts[page].push(opt[1]);
    if (spBroadcast) {
        if (spOpts.length === 0) {
            spOpts.push([page]);
        }
        spOpts[spOpts.length - 1].push(opt[0]);
        spOpts[spOpts.length - 1].push(opt[1]);
    }
    if (!omitDraw) {
        SpDrawLine(opt[0], opt[1]);
    }
}

function SpDrawSeries(opts) {
    if (opts.length === 0) {
        return;
    }
    for (let i = 0; i < opts.length; i++) {
        SpChangePage(opts[i][0], opts[i][0] !== opts[opts.length - 1][0]);
        for (let j = 1; j < opts[i].length; j += 2) {
            SpAct([opts[i][j], opts[i][j + 1]], opts[i][0], opts[i][0] !== opts[opts.length - 1][0]);
        }
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
    SpDrawSeries(opts);
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
    if (!spStartTime || !spBroadcast || !spPdfObject || !spBroadcastPrevTime || Date.now() - spBroadcastPrevTime < spBroadcastOptInterval) {
        return;
    }
    let curCoord = SpGenerateCoord(e);
    SpAct([spBroadcastPrevCoord, curCoord], spCurPage, false);
    spBroadcastPrevTime = Date.now();
    spBroadcastPrevCoord = curCoord;
}

function SpBroadcastSetPage(next) {
    if (!spStartTime || !spBroadcast || !spPdfObject) {
        return;
    }
    if (next) {
        if (spCurPage === spPdfObject.numPages - 1) {
            return;
        }
        SpChangePage(spCurPage + 1, false);
    } else {
        if (spCurPage === 0) {
            return;
        }
        SpChangePage(spCurPage - 1, false);
    }
}

function SpGenerateCoord(e) {
    let bounds = spCanvasOpt.getBoundingClientRect();
    return [(e.clientX - bounds.left) / bounds.width, (e.clientY - bounds.top) / bounds.height];
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