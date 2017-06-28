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
// or one with an  id to update an existing task on the server
function writeTask(task, directive) {

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
      if (this.status == 200) {
        console.log("success: task created or updated")
        serverTask = JSON.parse(this.responseText);
        task.id = serverTask.id;
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
}

function updateTask(task) {
  if (task == null || task.id == null || task.id == "") {
    console.log("error: id must be specified on update.  aborting update.");
    return;
  }
  writeTask(task, "PUT");
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
                         null, // stringToDate(task.startTime), 
                         task.estimate/nSecPerMinute);
            t.state = task.state; // see mapping in TaskState enum
            t.startTime = stringToDate(task.startTime);
            if (t.isComplete()) {
              done.insertTask(t);
            }
            else {
              if (t.startTime == null) {
                stuff.insertTask(t);
              }
              else {
                scheduled.insertTask(t, 'timesort');
              }
            }
        }
      }
    }
  };
  ajaxGet(ajax, tasksURL());
}
