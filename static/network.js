/*
Assorted network functions for UI ajax requests.
 */

function makeRequest(options) {

    // HTTP method must be defined
    if (typeof options.method === "undefined" || options.method.length < 1) {
        console.error("Fatal: method must be defined in options object");
        return false;
    }

    // URI to request must be defined
    if (typeof options.uri === "undefined" || options.uri.length < 1) {
        console.error("Fatal: uri must be defined in options object");
        return false;
    }

    var httpRequest = new XMLHttpRequest();

    if (!httpRequest) {
        console.error("Fatal: Could not create XHR instance :(");
        return false;
    }

    // Build any query parameters
    if (typeof options.queryParams !== "undefined") {
        var queryParams = encodeParams(options.queryParams);
        options.uri = options.uri + "?" + queryParams;
    }

    httpRequest.open(options.method, options.uri);
    httpRequest.setRequestHeader('X-Requested-With', 'XMLHttpRequest');

    // Wrap load handler around onload event. A handler for the load event
    // should take the following params: (xhr object, http code, response json)
    if (typeof options.load !== "undefined") {
        httpRequest.onload = function(event) {
            options.load(event.target, event.target.status, JSON.parse(event.target.responseText));
        };
    }

    // Wrap error handler around onerror event. A handler for the onerror event
    // should take the following params: (xhr object, http code, response text)
    if (typeof options.failure !== "undefined") {
        httpRequest.onerror = function(event) {
            options.failure(event.target, event.target.status, event.target.responseText);
        };
    }

    if (options.method === "POST") {
        httpRequest.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
        httpRequest.send(encodeParams(options.formParams));
    } else {
        httpRequest.send();
    }

}

function encodeParams(params) {
    var out = [];
    for (var key in params) {
        out.push(key + '=' + encodeURIComponent(params[key]));
    }
    return out.join('&');
}