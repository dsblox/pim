// the ajax calls will initially load tasks into these lists
// depending on the attributes on the tasks
var scheduled = new TaskList();
var stuff = new TaskList();
var done = new TaskList();

var planWeek = new TaskList();
planWeek.setId("PW");
var planDay = new TaskList();
planDay.setId("PD");

var days = {};
var currday = new TaskList();

function upsertTask(view) {
	var t = null;
	var f = document.getElementById("newTask");
	var name    = f.elements["task"].value;
    var strdate = f.elements["startdate"].value;
	var strtime = f.elements["starttime"].value;
	var today   = (f.elements["today"].value == "true"); // hidden field useable by every UI
	var time = null;

	if (strtime.length > 0) {
		// if date not specified, assume today		
		if (strdate.length == 0) {
			var today = new Date();
			strdate = today.getFullYear()+'-'+(today.getMonth()+1)+'-'+today.getDate();			
		}
		// remember time will be sent in local time zone
		// with TZ info and will be stored in GMT		
		time = new Date(strdate + " " + strtime);
	}

	if ((strtime.length > 0) && (strdate.length > 0)) {
		// remember time will be sent in local time zone
		// with TZ info and will be stored in GMT
	  	time = new Date(strdate + " " + strtime);
	}
	console.log("upsertTask: time=" + time);


	// we decide whether or not to set the thisweek flag based on the hidden field
	var thisWeek = (f.elements["thisweek"].value == "true"); // hidden field useable by every UI

	var duration = parseInt(f.elements["duration"].value);
	if (isNaN(duration)) {
		duration = null;
	}
	if (currTask == null) {
		t = new Task(null, name, time, null, duration, false, today, thisWeek);
    	createTask(t); // create a new task on the server

    	if (view == 'planning') {
    		if (thisWeek) {
    			planWeek.insertTask(t);
    		}
    		else if (today) {
    			planDay.insertTask(t);
    		}
 		}
 		else {
			var list = stuff;
			var sort = false;
			if (time != null) {
				list = scheduled;
				sort = true;
			}
 			list.insertTask(t, sort?'targetstarttime':'end');
 		}
 		console.log("upsertTask: new task id=" + t.id);
 		// tbd: put the id into the local version of the task
	} else {
		t = currTask;
		t.setName(name);
		t.setTargetStartTime(time);
		t.setEstimate(duration);
		t.setToday(today);
		t.setThisWeek(thisWeek);
		moveTask(t);
		currTask = null;
    	replaceTask(t); // set all the fields on this task
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
    planWeek.removeTask(currTask);
    planDay.removeTask(currTask);

    // delete the task from the server (when we write it)
    killTask(currTask);

	// remove what we hope is the last reference to the task
	currTask = null;
}


function cancelModal() {
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
	if (!found) {
		findTaskInList(planWeek, id, returnType);
	}
	if (!found) {
		findTaskInList(planDay, id, returnType);
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
		case "PW": result = planWeek; break;
		case "PD": result = planDay;  break;
	}
	return result;
}

// this function is called when a task is marked completed
// or "unmarked" completed - it moves the task from list
// to list and if vue works properly it'll be really cool
// note that we get the event the value on the task hasn't
// changed yet so we check for the opposite.
// right now it will also make an ajax call to update the task
// TBD: in today view, when the date of a task changes to another
//   date then we should remove it from this view entirely.  Not
//   being done right now.  Once we have a planning view where
//.  we can see all our tasks across days we can implement that.
function moveTask(task) {
  if (task == null) { return; }

  // console.log(task.actualCompletionTime);
  console.log("in moveTask()");


  // task is done
  if (task.isComplete()) {

    // make sure it is in the done list
    done.insertTask(task, false);

    // make now it's completion time
    now = new Date();
    task.setActualCompletionTime(now.toJSON());

    // if it was in the scheduled or stuff lists remove it
    scheduled.removeTask(task);
    stuff.removeTask(task);

    // when we update change the state and completion time
    task.dirty = ["state", "actualcompletiontime"];


  } else {

    // remove from done list if it was there
    done.removeTask(task);

    // clear any completion time if its not really done
    task.setActualCompletionTime(null);

    // put it into scheduled or stuff by whether it has startTime
    if (task.hasStartTime()) {
      scheduled.insertTask(task, 'targetstarttime');
      stuff.removeTask(task);
    } else {
      stuff.insertTask(task);
      scheduled.removeTask(task);
    }

    // when we update change the state and completion time
    task.dirty = ["state", "actualcompletiontime", "targetstarttime"];
  }

  // call the server to update the task persistently
  // only changing the state and completion time

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

function clearTodayAndList(list) {
  var lenClear = list.tasks.length;
  // go backwards so we can remove items from the end
  // and do the loop for both server and UI
  for (var idxClear = lenClear - 1; idxClear >= 0; idxClear--) {
    console.log(list.tasks[idxClear].name);
    task = list.tasks[idxClear];
    clearToday(task);
    list.removeTask(task);
  }  
}

function setToday(task) {
	if (!task.isToday()) {
		task.setToday(true);
		planDay.insertTask(task);
		task.setTags = ['today'];
		updateTask(task);
	}
}

function resetToday(task) {
	if (task.isToday()) {
		task.setToday(false);
		planDay.removeTask(task);
		task.resetTags = ['today'];
		task.setTags = [];
		updateTask(task);
	}
}


