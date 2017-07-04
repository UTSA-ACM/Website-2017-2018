function makeRequest(method, uri, queryParams, formParams) {

    var httpRequest = new XMLHttpRequest();

    if (!httpRequest) {
        console.error("Fatal: Could not create XHR instance :(");
        return false;
    }

    httpRequest.onreadystatechange = callback;
    httpRequest.open(method, uri);

    httpRequest.setRequestHeader('X-Requested-With', 'XMLHttpRequest');

    if (method === "POST") {
        httpRequest.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
        httpRequest.send(encodeParams(formParams));
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