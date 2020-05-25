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

function ajaxPayload(xmlhttp, url, payload, directive) {
  json = JSON.stringify(payload);
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

