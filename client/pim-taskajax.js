/*
=========================================================================
 TaskAjax
-------------------------------------------------------------------------
 This file is the bridge between the JavaScript task world and the
 server-side Ajax calls.  It has knowledge of both Task and TaskList
 objects and the underlying JSON, and hides that mapping from its
 clients.

 It often operates with callbacks to allow access to JavaScript versions
 of Tasks and TaskLists from within Ajax calls.  For example, the
 collectTasks() method takes a URL to an API call that returns a list
 of tasks, and calls the hook for each task as it is read in, so that
 the caller can do something with it (like put it in the UI).

 Write functions typically just take a Task object and do the needed
 operation (create, replace, update, delete) against it on the server.
========================================================================*/


// set local to true to convert date to local timezone
// typically we want them in the local time zone when we are in weekly / daily views
// and we want them in UTC in historical views - though perhaps we should change that?
function stringToDate(strDate, local) {
  if (strDate == null) {
    return null;
  }
  date = new Date(strDate.substring(0,strDate.length-1));
  if ((strDate.slice(-1) == "Z") && (local)) { // Z as last char means UTC
  var offset = new Date().getTimezoneOffset();
    date.setMinutes(date.getMinutes() - offset);
  } 

  return date;
}

function dateToForm(date) {
  if (date == null) return "";
  return date.toISOString().slice(0,10);
}

function timeToForm(date) {
  if (date == null) return "";

  var day = date.getDate(),
      month = date.getMonth() + 1,
      year = date.getFullYear(),
      hour = date.getHours(),
      min  = date.getMinutes();

  month = (month < 10 ? "0" : "") + month;
  day = (day < 10 ? "0" : "") + day;
  hour = (hour < 10 ? "0" : "") + hour;
  min = (min < 10 ? "0" : "") + min;

  var displayDate = year + "-" + month + "-" + day;
  var displayTime = hour + ":" + min; 

  return displayTime;
}



/*
=========================================================================
 URL Functions
-------------------------------------------------------------------------
 These all eventually make use of makeURL in the pim-ajax package, but
 all are task specific so belong with their uses in this file.

 Functions:
  * tasksURL - refer to a task by id (used by many directives)
  * tasksTodayURL - finds all today tasks
  * tasksThisWeekURL - finds all thisweek tasks
  * tasksCompleteURL - finds all complete tasks in the system
  * tasksFundURL - finds all tasks at a certain date
========================================================================*/
// we should find a way to make this more dynamic
var baseURL = "https://localhost:4000/";

function makeURL(cmd, tags = null) {
  var params = "";
  if (tags != null) {
    // hrm - those tags better not already have a comma in them
    params = "?tags=" + tags.join(",");
  }
  return baseURL + cmd + params;
}

function serverStatusURL() {
  return makeURL("status")
}

function tasksURL(id = "") {
  var rest = "tasks";
  if (id) {
    rest += "/";
    rest += id;
  }
  return makeURL(rest)
}

function tasksDefaultSystemTagURL(systemTag, tags = null) {
  var urlTags = [];
  if (tags == null) {
    urlTags.push(systemTag);
  }
  else if (typeof(tags) == "string") {
    urlTags = [systemTag, tags]
  }
  else {
    tags.map(function(e){urlTags.push(e)});
    urlTags.push(systemTag);
  }
  return makeURL("tasks", urlTags);  
}

function tasksTodayURL(tags = null) {
  return tasksDefaultSystemTagURL("today", tags);
}

function tasksThisWeekURL(tags = null) {
  return tasksDefaultSystemTagURL("thisweek", tags);
}

function tasksCompleteURL(tags = null) {
  return makeURL("tasks/complete", tags)
}

function tasksFindURL(date, tags = null) {
  var rest = "tasks";
  if (date) {
    rest += "/date/";
    rest += date;
  }
  return makeURL(rest, tags)
}

function tagsURL() {
  return makeURL("tags")
}

/*
=========================================================================
 serverStatus
-------------------------------------------------------------------------
 Simplest call possible to just see if the server is alive, but performed
 asynch so it can be used in a dynamic UI if desired.  Mostly used in
 my test UIs.
========================================================================*/
function serverStatus() {
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.responseText != "OK") {
      pimShowError(this.responseText);      
    }
  };
  ajaxGet(ajax, serverStatusURL());  
}


/*
=========================================================================
 Write Functions for Tasks
-------------------------------------------------------------------------
 This section of the code includes the Ajax functions to create, update
 and delete a task.  Each expects a task to delete, executes its
 function with a server call, and (TBD) can have a callback function 
 invoked on completion with "OK" or any error message.

 Functions:
  * createTask - creates the task on the server (POST)
  * replaceTask - re-writes all fields from client to server (PUT)
  * updateTask - writes only specified fields to server (PATCH)
  * killTask - deletes the task on the server (DELETE)
  * writeTask - worker that does all the work for create, update, replace
========================================================================*/

/*
-------------------------------------------------------------------------
 killTask()
-----------------------------------------------------------------------*/
function killTask(task) {

  // validate inputs that we have a task and a valid id
  var id = "";
  if (task == null || task.getId() == null || task.getId().length == 0) {
    console.log("error: id must be specified on delete.  aborting delete.");
    return;
  }
  id = task.getId();

  // prepare the callback
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        console.log("success: task deleted");
      }
      else {
        console.log("failed: task not deleted http response: " + this.status);
      }
    }
  };
  
  // make the ajax call to delete
  ajaxDelete(ajax, tasksURL(id), "DELETE");
}


/*
-------------------------------------------------------------------------
 writeTask()
-------------------------------------------------------------------------
 Does the work for POST, PUT and PATCH of a task to the server.  Note we
 enforce the task provided must not have an id already if this is a POST
 and it must have an id if a PUT or PATCH.  Almost all the code in here
 is just checking the right combo of input parameters for the directive.
 The only real code posts the request to the server, and (importantly)
 after a POST it will set the id of the JS task originally provided.

 There are semantics to know about PATCH calls to let the server know
 which fields are to be set.  See updateTask() below for details.

 Inputs:
  * task - JavaScript task object to be written to the server
  * directiive - POST to create, PUT to replace, PATCH to update

 TBD: Clean up the strange system-tag support that auto converts the 
      booleans into server tags.  I don't think it is needed anymore
      since we got rid of the booleans on the JS side!
-----------------------------------------------------------------------*/
function writeTask(task, directive) {

  if (task == null) {
    console.log("writeTask() error: null task provided");
    return;
  }

  // convert booleans to tags - note that updates must be done by callers
  // since this code cannot know what has actually changed.
  if (directive == "POST" || directive == "PUT") {
    if (task.isToday()) {
      task.addTag("today");
    }
    if (task.isThisWeek()) {
      task.addTag("thisweek");
    }
  }

  // collect the task from the form elements
  var id = "";
  if (directive == "POST") {
    if (task.getId() != null || task.getId() == "") {
      console.log("error: id should be null on create.  aborting create.");
      return;
    }
  }
  else {
    if (task.getId() == null || task.getId().length == 0) {
      console.log("error: id must be specified on update.  aborting update.");
      return;
    }
    id = task.id;
  }

  ajax = ajaxObj();

  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200 || this.status == 201) {
        // console.log("success: task created or updated");

        // update the task with newly created id
        if (directive == "POST") {
          serverTask = JSON.parse(this.responseText);
          task.id = serverTask.id; 
        }
      }
      else {
        // console.log("writeTask: task create or updated failed.");
        // TBD: undo any writes in the UI such as
        // remove an added task that didn't change
        // or undo a modification that didn't "take"
        // or at least notify the user that the task
        // was not saved.
        alert("Task not properly saved.  Please refresh and try again.");
      }
    }
  };
  // a little confusing but it saves a lot of code
  // wrapper task passes in the right directive and based
  // on the directive (POST, PUT, PATCH) we've set up
  // the id to be valid or not to build the right URL
  ajaxPayload(ajax, tasksURL(id), task, directive);
}

/*
-------------------------------------------------------------------------
 updateTask()
-------------------------------------------------------------------------
 While just a wrapper to writeTask, there are semantics to understand
 regarding how the server determines which fields to actually update.
 If a field name is in the "dirty" array on the task, then it will be
 written, otherwise it will be left in place on the server.  There is 
 special handling for tags, which can be set, reset or both.

 Fields on task to instruct the server on a PATCH:
  * task.dirty[] - field names that changed and should be patched
  * task.setTags[] - tags to write to the task on the server
  * task.resetTags[] = tags to turn OFF on the server
-----------------------------------------------------------------------*/
function updateTask(task) {
  writeTask(task, "PATCH");
}

function createTask(task) {
  writeTask(task, "POST");
}

function replaceTask(task) {
  writeTask(task, "PUT");
}


/*
=========================================================================
 System Tag Handling
-------------------------------------------------------------------------
 These workers probably belong in pim-task-changes, and are used to reset
 the today and thisweek system tags.

 TBD: We should see if these are still needed now that we've made the
 system tags "just another tag" on the JavaScript Task object.
========================================================================*/
function clearToday(task) {
  if (task.isToday()) {
    task.setToday(false);
    task.resetTags = ["today"];
    task.setTags = [];
    updateTask(task);
  }
}

function clearThisWeek(task) {
  if (task.isThisWeek()) {
    task.setThisWeek(false);
    task.resetTags = ["thisweek"];
    task.setTags = [];
    updateTask(task);
  }
}


/*
=========================================================================
 Query Functions for Tasks - collectTasks()
-------------------------------------------------------------------------
 This function makes the requested AJAX call to the server expecting
 back a list of tasks, which it processes one by one by invoking the
 provided callback function, so the invoker can do whatever it wants
 with the returned tasks.  It also makes an additional callback (if
 provided) when done so the caller can get any errors or perform any
 followup work.

 Note that if taskCallback is not provided, then we don't bother to
 loop over the tasks at all, and we return the entire parsed JSON
 response to the caller.

 Inputs:
  * url          - full url to invoke, should return a task list
  * taskCallback - callback to be provided each task
  * doneCallback - callback invoked with error code or "OK" when done
========================================================================*/
function loadTags(taskJSON, taskJS) {
  if (taskJSON.tags != null && taskJSON.tags.length > 0) {
    taskJSON.tags.map(function(tag){taskJS.addTag(tag)});
  }
}

var nSecPerMinute = 60000000000;
function taskJsonToJs(jsonTask) {
  t = new Task(jsonTask.id, 
               jsonTask.name, 
               null, // stringToDate(task.getTargetStartTime()), 
               null, // actualCompletionTime
               jsonTask.estimate/nSecPerMinute,
               false); // we set complete state later
  t.state = jsonTask.state; // see mapping in TaskState enum
  t.setTargetStartTime(stringToDate(jsonTask.targetStartTime, true));
  t.setActualCompletionTime(stringToDate(jsonTask.actualCompletionTime, true));
  loadTags(jsonTask, t); // will set system tags like today and thisweek 
  return t;
}

function collectTasks(url, taskCallback, doneCallback = null) {
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        var taskList = new TaskList();
        jsonTasks = JSON.parse(this.responseText);

        for (i = 0; i < jsonTasks.length; i++) {

          // create a javascript task from the JSON
          t = taskJsonToJs(jsonTasks[i]);

          // let the caller do what it wants with the task
          if (taskCallback != null) {
            taskCallback(t);
          } // if a callback (and this the individual tasks) was requested

          // if caller wants a full list back then we'll collect here
          if (doneCallback != null) {
            taskList.insertTask(t);
          }

        } // for each task that came from the server
  
        // if caller asked for a callback at the end then assume
        // it wants the list of tasks (not actually always true that they do)
        if (doneCallback != null) {
          doneCallback(taskList);
        }
      } // if server returned a 200
      else {
        pimAjaxError(this.responseText); 
        if (doneCallback != null) {
          doneCallback(null);
        }
        // if we got an empty list, let's ask the server for it's status to
        // see if there is a better explanation besides "empty list"
        // that we share with the user.
        // serverStatus(); // probably should not always do this!
      }
    }
  };
  ajaxGet(ajax, url);
}

/*
-----------------------------------------------------------------------
 collectTask - in use only in our tests, but gets one task from server
---------------------------------------------------------------------*/
function collectTask(id, doneCallback) {
  if (doneCallback == null || id == null || id == "") {
    return false;
  }
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        var jsonTask = JSON.parse(this.responseText);
        task = taskJsonToJs(jsonTask);
        doneCallback(task);
      }
    }
  };
  ajaxGet(ajax, tasksURL(id));
  return true;
}

/*
-----------------------------------------------------------------------
 collectTags - returns all the tags across all tasks (with counts)
---------------------------------------------------------------------*/
function collectTags(doneCallback) {
  if (doneCallback == null) {
    return false;
  }
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        var jsonTags = JSON.parse(this.responseText);
        doneCallback(jsonTags);
      }
    }
  };
  ajaxGet(ajax, tagsURL());
  return true;
}

