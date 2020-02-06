// call the server to delete a task
function killTask(task) {

  // validate inputs that we have a task and a valid id
  var id = "";
  if (task == null || task.id == null || task.id.length == 0) {
    console.log("error: id must be specified on delete.  aborting delete.");
    return;
  }
  id = task.id;

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

// call the server to create a new task or update an existing one
// pass in a task with a null id to create a new task on the server
// or one with an id to update an existing task on the server
function writeTask(task, directive) {

  // convert booleans to tags - not that updates must be done by callers
  // since this code cannot know what has actually changed.
  if (directive == "POST" || directive == "PUT") {
    task.tags = [];
    if (task.isToday()) {
      task.tags.push("today");
    }
    if (task.isThisWeek()) {
      task.tags.push("thisweek");
    }
  }

  // collect the task from the form elements
  var id = "";
  if (directive == "POST") {
    if (task.id != null || task.id == "") {
      console.log("error: id should be null on create.  aborting create.");
      return;
    }
  }
  else {
    if (task.id == null || task.id.length == 0) {
      console.log("error: id must be specified on update.  aborting update.");
      return;
    }
    id = task.id;
  }

  ajax = ajaxObj();

  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200 || this.status == 201) {
        console.log("success: task created or updated");

        // update the task with newly created id
        if (directive == "POST") {
          serverTask = JSON.parse(this.responseText);
          task.id = serverTask.id; 
        }
      }
      else {
        console.log("writeTask: task create or updated failed.");
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

function createTask(task) {
  if (task == null || (task.id != null || task.id == "")) {
    console.log("error: id should be null on create.  aborting create.");
    return;
  }
  writeTask(task, "POST");
  console.log("createTask: new task id=" + task.id);
}

// before calling this function, called should set any of:
// task.dirty[] - field names that changed
// task.setTags[] - tags to set to true
// task.resetTags[] = tags to set to false
function updateTask(task) {
  if (task == null || task.id == null || task.id == "") {
    console.log("error: id must be specified on update.  aborting update.");
    return;
  }
  writeTask(task, "PATCH");
}

function replaceTask(task) {
  if (task == null || task.id == null || task.id == "") {
    console.log("error: id must be specified on update.  aborting replace.");
    return;
  }
  writeTask(task, "PUT");
}

function clearToday(task) {
  if (task.isToday()) {
    task.setToday(false);
    task.resetTags = ["today"];
    task.setTags = [];
    updateTask(task);
  }
}

function findTag(taskJSON, tag) {
  found = false;
  if (taskJSON.tags !== null) {
    len = taskJSON.tags.length;
    idx = 0;
    while (!found && idx < len) {
      found = (taskJSON.tags[idx] == tag);
      idx++;
    }
  }  
  return found;
}

function loadTags(taskJSON, taskJS) {
  taskJS.setToday(findTag(taskJSON, "today"));
  taskJS.setThisWeek(findTag(taskJSON, "thisweek"));
}

// call the server to get our initial list of tasks
function loadTasks() {
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        tasks = JSON.parse(this.responseText);

        for (i = 0; i < tasks.length; i++) {
            task = tasks[i];
            t = new Task(task.id, 
                         task.name, 
                         null, // stringToDate(task.getTargetStartTime()), 
                         null, // actualCompletionTime
                         task.estimate/nSecPerMinute,
                         task.today);
            t.state = task.state; // see mapping in TaskState enum
            t.setTargetStartTime(stringToDate(task.targetStartTime));
            t.setActualCompletionTime(stringToDate(task.actualCompletionTime));
            loadTags(task, t);
            if (t.isComplete()) {
              done.insertTask(t);
            }
            else {
              if (t.getTargetStartTime() == null) {
                stuff.insertTask(t);
              }
              else {
                scheduled.insertTask(t, 'targetstarttime');
              }
            }
        }
      }
    }
    else {
      pimAjaxError(this.responseText);      
    }
  };
  ajaxGet(ajax, tasksURL());
}

// call the server to get our initial list of tasks
function loadTasksToday() {
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        tasks = JSON.parse(this.responseText);

        for (i = 0; i < tasks.length; i++) {
            task = tasks[i];
            t = new Task(task.id, 
                         task.name, 
                         null, // stringToDate(task.getTargetStartTime()), 
                         null, // actualCompletionTime
                         task.estimate/nSecPerMinute,
                         false, // we set complete state later
                         task.today);
            t.state = task.state; // see mapping in TaskState enum
            t.setTargetStartTime(stringToDate(task.targetStartTime));
            t.setActualCompletionTime(stringToDate(task.actualCompletionTime));
            loadTags(task, t);
            if (t.isComplete()) {
              done.insertTask(t, 'actualendtime');
            }
            else {
              if (t.getTargetStartTime() == null) {
                stuff.insertTask(t);
              }
              else {
                scheduled.insertTask(t, 'targetstarttime');
              }
            }
        }
      }
      else {
        pimAjaxError(this.responseText); 
      }
    }
  };
  ajaxGet(ajax, tasksTodayURL());
}


// call the server to get the list of tasks for this week
function loadTasksThisWeek() {
  console.log("loadTasksThisWeek() - entry");
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        tasks = JSON.parse(this.responseText);

        for (i = 0; i < tasks.length; i++) {
          task = tasks[i];
          console.log(task.id);
          t = new Task(task.id, 
                       task.name, 
                       null, // stringToDate(task.getTargetStartTime()), 
                       null, // actualCompletionTime
                       task.estimate/nSecPerMinute,
                       false, // we set complete state later
                       task.today);
          t.state = task.state; // see mapping in TaskState enum
          t.setTargetStartTime(stringToDate(task.targetStartTime));
          t.setActualCompletionTime(stringToDate(task.actualCompletionTime));
          loadTags(task, t);
          planWeek.insertTask(t);
        }
      }
      else {
        pimAjaxError(this.responseText); 
      }
    }
  };
  ajaxGet(ajax, tasksThisWeekURL());
  console.log("loadTasksThisWeek() - exit");  
}

// call the server to get the list of tasks for this day's planning view
function loadTasksThisDay() {
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        tasks = JSON.parse(this.responseText);

        for (i = 0; i < tasks.length; i++) {
          task = tasks[i];
          t = new Task(task.id, 
                       task.name, 
                       null, // stringToDate(task.getTargetStartTime()), 
                       null, // actualCompletionTime
                       task.estimate/nSecPerMinute,
                       false, // we set complete state later
                       task.today);
          t.state = task.state; // see mapping in TaskState enum
          t.setTargetStartTime(stringToDate(task.targetStartTime));
          t.setActualCompletionTime(stringToDate(task.actualCompletionTime));
          loadTags(task, t);
          planDay.insertTask(t);
        }
      }
      else {
        pimAjaxError(this.responseText); 
      }
    }
  };
  ajaxGet(ajax, tasksTodayURL());
}


// call the server to get just the tasks completed on the date specified
function loadTasksByDay(date) {
  currday.clean();
  if (date == null) {
    return;
  }
  ajax = ajaxObj();
  ajax.onreadystatechange = function() {
    if (this.readyState == 4) {
      if (this.status == 200) {
        tasks = JSON.parse(this.responseText);
        if (tasks !== null) {
          for (i = 0; i < tasks.length; i++) {
            task = tasks[i];
            t = new Task(task.id, 
                         task.name, 
                         null, // stringToDate(task.getTargetStartTime()), 
                         null, // actualCompletionTime
                         task.estimate/nSecPerMinute,
                         task.today);
            t.state = task.state; // see mapping in TaskState enum
            t.setTargetStartTime(stringToDate(task.targetStartTime));
            t.setActualCompletionTime(stringToDate(task.actualCompletionTime));
            loadTags(task, t);

            // currday is bound to vue so updating it updates the display
            currday.insertTask(t);
          } /* for each task returned */
        } /* if we could parse the response into tasks */
        else {
          pimShowError("internal error: could not parse response: " + this.responseText);
        }
      }
      else {
        pimAjaxError(this.responseText);
      }
    }
  };
  ajaxGet(ajax, tasksFindURL(date));
}


