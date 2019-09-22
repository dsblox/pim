var baseURL = "https://localhost:4000/";

function makeURL(rest) {
  return baseURL + rest;
}

function tasksURL(id = "") {
  var rest = "tasks";
  if (id) {
    rest += "/";
    rest += id;
  }
  return makeURL(rest)
}

function tasksTodayURL() {
  return makeURL("tasks/today")
}

function tasksThisWeekURL() {
  return makeURL("tasks/thisweek")
}


function tasksFindURL(date) {
  var rest = "tasks";
  if (date) {
    rest += "/date/";
    rest += date;
  }
  return makeURL(rest)
}


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
  console.log("ajaxPayload(): payload = " + json);
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

function loadTask(id) {
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      displayRawResponse(this.status, this.responseText);
      if (this.status == 200) {
        task = JSON.parse(this.responseText);
        document.getElementById("detail.id").value = task.id;
        document.getElementById("detail.name").value = task.name;
        document.getElementById("detail.state").value = task.State;
      }
    }
  };
  ajaxGet(ajax, tasksURL(id));
}


// directives:
// POST = Create
// PATCH = Update
// PUT = Replace
function writeTask(directive) {

  // collect the task from the form elements
  var task = {};
  id = document.getElementById("detail.id").value;
  task.name = document.getElementById("detail.name").value;
  // task.state = document.getElementById("detail.state").value;
  if (directive != "POST") {
    task.id = id;
  }
  else {
    id = "";
  }

  ajax = ajaxObj();

  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      displayRawResponse(this.status, this.responseText);
      if (this.status == 200) {
        // successful - should we do something?
      }
    }
  };

  // a little confusing but it saves a lot of code
  // wrapper task passes in the right directive and based
  // on the directive (POST, PUT, PATCH) we've set up
  // the id to be valid or not to bulid the right URL
  ajaxPayload(ajax, tasksURL(id), task, directive);
}

function createTask() {
  writeTask("POST"); // POST means create
}
function replaceTask() {
  writeTask("PUT"); // PUT means replace
}
function updateTask() {
  // this isn't working yet - if a field has not been changed it should
  // not even be included in the JSON that gets sent to the server.  We
  // need to add "dirty flags" to our form fields and only send the dirty
  // info if we want to properly test this.
  writeTask("PATCH"); // PATCH means update
}

function deleteTask() {
  id = document.getElementById("detail.id").value;
  ajax = ajaxObj();

  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      displayRawResponse(this.status, this.responseText);
      if (this.status == 200) {
        // successful - should we do something?
      }
    }
  };

  ajaxDelete(ajax, tasksURL(id));
}
