let allQualities = [
    {val: "0", tag: "Audio Only"},
    {val: "1", tag: "144p"},
    {val: "2", tag: "360p"},
    {val: "3", tag: "576p"}
];

let pageSize = 1;

function Api(name) {
    return "../../api/" + name;
}

function FileUrl(file) {
    return "../res/file/" + file;
}

function MediaUrl(file) {
    return "../res/media/" + file;
}

function GetSourceTime() {
    return Math.floor(vodSource[0].currentTime * 1000);
}

function HandleError(e) {
    console.log(e);
    let str = "Status: " + e.status;
    if (e.responseJSON) {
        str += "\n" + "Message: " + e.responseJSON["Message"];
    }
    alert(str);
}

/**
 * Wrapped ajax call function
 * @param url Request url
 * @param param Url parameters
 * @param data Http body
 * @param method Call method
 * @param success Callback function when succeeded
 * @param err Callback function when failed
 * @param json If the data should be in json format
 */
function Ajax(url, param, data, method, success, err, json) {
    // Encoded all url params
    if (param)
        url += "?" + $.param(param);
    // Generate args object
    let args = {
        url: url,
        type: method,
        data: data,
        success: success ? success: function(data){},
        error: err ? err : HandleError
    };
    // Delete data field, if there is no http body
    // Else if the call should be in json format, encode it
    // Else if the call should be in post form format, encode it
    if (data === null)
        delete args["data"];
    else if (json === true) {
        args["contentType"] = "application/json;charset=utf-8;";
        args["dataType"] = "json";
        args["data"] = JSON.stringify(data);
    }
    else if (json === false) {
        args["contentType"] = false;
        args["processData"] = false;
    }
    // Execute ajax request
    $.ajax(args);
}