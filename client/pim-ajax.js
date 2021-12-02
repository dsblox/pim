/*
=========================================================================
 PIM Ajax
-------------------------------------------------------------------------
 This file holds a compeltely generic Ajax implementation, providing a
 very weak wrapper over a standard XMLHttpRequest().

 Essentially, the ajaxObj() function wrappers a XMLHttpRequest() object
 so we can conveniently wait on onreadystatechange events to collect
 the results as needed.  Then we provide some simple wrappers to open()
 and send() to abstract things for our clients.
 =======================================================================*/
function ajaxObj() {
 var xmlhttp;
  if (window.XMLHttpRequest) {
    xmlhttp = new XMLHttpRequest();
  } else {
    // code for older browsers
    xmlhttp = new ActiveXObject("Microsoft.XMLHTTP");
  }
  return xmlhttp;  
}

function ajaxSimple(xmlhttp, url, directive) {
  xmlhttp.open(directive, url, true);
  xmlhttp.send();
}

// the replacer is needed only to extract the URL from the
// hyperlink because our UI has coded more functionality
// than we can squeeze through the API so far.  So we're
// extracting just the string from the task before adding
// it to the payload for Hyperlinks.  All other values
// in the task are stringified into JSON directly.
function replacer(key, value) {
  let retValue = value
  if (key == 'links' && value && value.length) {
    retValue = value.map(x => x.url);    
  }
  // console.log("replacer: " + key + " " + retValue)
  return retValue
}

function ajaxPayload(xmlhttp, url, payload, directive) {
  // console.log("payload")
  json = JSON.stringify(payload, replacer); // replacer just for hyperlinks
  xmlhttp.open(directive, url, true);
  xmlhttp.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
  xmlhttp.send(json);
}

function ajaxGet(xmlhttp, url) {
  ajaxSimple(xmlhttp, url, "GET");
}

function ajaxDelete(xmlhttp, url) {
  ajaxSimple(xmlhttp, url, "DELETE");
}

function ajaxPost(xmlhttp, url, payload) {
  ajaxPayload(xmlhttp, url, payload, "POST");
}

function ajaxPut(xmlhttp, url, payload) {
  ajaxPayload(xmlhttp, url, payload, "PUT");
}

