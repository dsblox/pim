// the ajax calls will initially load tasks into these lists
// depending on the attributes on the tasks
var scheduled = new TaskList();
var stuff = new TaskList();
var done = new TaskList();


function upsertTask() {
	var t = null;
	var f = document.getElementById("newTask");
	var name = f.elements["task"].value;
    var strdate = f.elements["startdate"].value;
	var strtime = f.elements["starttime"].value;
	var time = null;
	if ((strtime.length > 0) && (strdate.length > 0)) {
		// remember time will be sent in local time zone
		// with TZ info and will be stored in GMT
	  	time = new Date(strdate + " " + strtime);
	}
	var duration = parseInt(f.elements["duration"].value);
	if (isNaN(duration)) {
		duration = null;
	}
	if (currTask == null) {
		t = new Task(null, name, time, duration, false);
    	createTask(t); // create a new task on the server
		var list = stuff;
		var sort = false;
		if (time != null) {
			list = scheduled;
			sort = true;
		}
 		list.insertTask(t, sort?'timesort':'end');
	} else {
		t = currTask;
		t.name = name;
		t.startTime = time;
		t.estimate = duration;
		moveTask(t);
		currTask = null;
    	updateTask(t); // update the task that changed
	}
}

function deleteTask() {
	if (currTask == null) {
		return;
	}

	// remove task from any list it is in
	done.removeTask(currTask);
	stuff.removeTask(currTask);
	scheduled.removeTask(currTask);

  // delete the task from the server (when we write it)
  killTask(currTask);

	// remove what we hope is the last reference to the task
	currTask = null;
}


function findTaskInList(list, id, returnType) {
	found = null;
	found = list.findTask(id);
	if (found != null) {
		if (returnType == 'task') {
			return found;
		}
		else if (returnType == 'list') {
			return list;
		}
		else {
			console.log("findTaskInList: returnType must be list or task, got: " + returnType)
			return null;
		}
	}
	else {
		return null;
	}
}

// search all the lists for the task and return the list it was in
function findTaskInPIMList(id, returnType) {
	found = null;
	found = findTaskInList(scheduled, id, returnType);
	if (!found) {
		found = findTaskInList(stuff, id, returnType);
	}
	if (!found) {
		found = findTaskInList(done, id, returnType);
	}
	return found;	
}

function listOfTask(id) {
	return findTaskInPIMList(id, 'list');
}

// search all the lists for the task
function findTask(id) {
	return findTaskInPIMList(id, 'task');
}

function extractTimeString(timestamp) {
	if (timestamp == null) {
		return null;
	}
	var hr = timestamp.getHours();
	var mn = timestamp.getMinutes();
	var hrstr = (hr < 10 ? "0" + hr : hr);
	var mnstr = (mn < 10 ? "0" + mn : mn);
	return hrstr + ":" + mnstr + ":" + '00';
}

function extractDateString(timestamp) {
  if (timestamp == null) {
    return null;
  }
  try {
  	parts = timestamp.toISOString().split("T");
  	strDate = parts[0];
  }
  catch (err) {
  	console.log("invalid date: " + err);
  	strDate = null;
  }
  return strDate;  
}

function stringToDate(strDate) {
  if (strDate == null) {
    return null;
  }
  date = new Date(strDate.substring(0,strDate.length-1));
  if (strDate.slice(-1) == "Z") { // Z as last char means UTC
	var offset = new Date().getTimezoneOffset();
  	date.setMinutes(date.getMinutes() - offset);
  }	

  return date;
}



function taskListFromID(id) {
	var result = null;
	switch (id) {
		case "C": result = scheduled; break;
		case "S": result = stuff;     break;
		case "D": result = done;      break;
	}
	return result;
}

// this function is called when a task is marked completed
// or "unmarked" completed - it moves the task from list
// to list and if vue works properly it'll be really cool
// note that we get the event the value on the task hasn't
// changed yet so we check for the opposite.
// right now it will also make an ajax call to update the task
function moveTask(task) {
  if (task == null) { return; }

  // task is done
  if (task.isComplete()) {

    // make sure it is in the done list
    done.insertTask(task, false);

    // if it was in the scheduled or stuff lists remove it
    scheduled.removeTask(task);
    stuff.removeTask(task);


  } else {

    // remove from done list if it was there
    done.removeTask(task);

    // put it into scheduled or stuff by whether it has startTime
    if (task.hasStartTime()) {
      scheduled.insertTask(task, 'timesort');
      stuff.removeTask(task);
    } else {
      stuff.insertTask(task);
      scheduled.removeTask(task);
    }
  }

  // call the server to update the task persistently
  updateTask(task);
}

function toggleTask(task) {
 	if (task == null) { return; }

  // note this used to be necessary - but now it isn't
  // i think vue fixed a bug in the data binding and is
  // now automatically changing the task value when it
  // didn't before
  // task.complete = !task.complete;

 	moveTask(task);
}

