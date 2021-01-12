/*
=========================================================================
 Task Changes
-------------------------------------------------------------------------
 The meat of our UI sits here, holding the UI-bound JavaSceript objects
 that populate all the tasks in our UI.  Right now, we have one file
 for all the objects that live in the UI, in anticipation of a true
 SPA architecture (right now, there are separate pages for the larger
 functions).

 Each UI object tends to be a list of tasks, and its bound models live
 here.  When the user does something to change a task's state, functions
 here are called to make sure:
   * the state change is reflected in the UI, and
   * the state change is reflected on the server.
========================================================================*/

/*
--------------------------------------------------------------------------
 TODAY page
------------------------------------------------------------------------*/
var scheduled = new TaskList();
var stuff = new TaskList();
var done = new TaskList();
var inprogress = new TaskList();

/*
--------------------------------------------------------------------------
 PLANNING page
------------------------------------------------------------------------*/
var planWeek = new TaskList();
planWeek.setId("PW");
var planDay = new TaskList();
planDay.setId("PD");

/*
--------------------------------------------------------------------------
 HISTORY page
------------------------------------------------------------------------*/
var days = {};
var selectedDate = null;
var currday = new TaskList();
var completedTaskDates = [];

/*
--------------------------------------------------------------------------
 GLOBALS - that tend to span pages
------------------------------------------------------------------------*/
var currentPage  = null; // recently added - not sure it is dependable
var selectedTags = [];   // list of which tags are selected for filtering
var allTags      = [];   // list of all tags in popularity order
var currTags     = [];   // list of tags in the current task for modal


/*
=========================================================================
 toggleTag(tag), selectTag(tag), deselectTag(tag)
-------------------------------------------------------------------------
 You can filter to any number of tags (tasks that have ALL selected tags
 set on them are displayed) by either calling select or deselect (to add
 or remove a tag selection), or by calling toggleTag() to turn a filter
 on/off (which is what the UI currently uses - calling toggle any time
 a tag is clicked).

 We (perhaps inefficiently) clear and reload the screen from the server 
 whenever you choose a tag.

 Inputs: tag - user-defined tag to filter to

 TBD: Consider the move from ALL tags to just one.  That should not
      require a server call - just remove the non-selected tags from
      each view.

 TBD: Data-bind the selectedTags somehow so things happen more auto-vue-
      magically? We've set it up to bind to the selector control, but
      not the lists.
========================================================================*/
function reloadOnTagChange() {
  if (currentPage == 'today') {
    scheduled.clean();
    done.clean();
    stuff.clean();
    inprogress.clean();
    loadTasksToday(selectedTags);
  }
  if (currentPage == "planning") {
    planWeek.clean()
    planDay.clean()
    loadTasksThisWeek(selectedTags);
    loadTasksThisDay(selectedTags);
  }  
}

function selectTag(tag, selectOne = false) {
  var somethingChanged = true;

  if ((selectedTags.length == 0 && tag == 'All') || 
      (selectedTags.indexOf(tag) != -1)) {
    somethingChanged = false;
  }
  else if (tag == 'All') {
    while (selectedTags.length > 0) {
      selectedTags.pop()
    }
  }
  else {
    if (selectOne) {
      while (selectedTags.length > 0) {
        selectedTags.pop()
      }
    }
    selectedTags.push(tag);    
  }

  if (somethingChanged) {
    reloadOnTagChange();
  }

  return somethingChanged;
}

function deselectTag(tag) {
  var somethingChanged;
  const idxOf = selectedTags.indexOf(tag);

  // you can't deleselect All or a tag not selected
  // just leave with false meaning no change
  if (tag == 'All' || idxOf == -1) {
    somethingChanged = false;
  }
  else {
    selectedTags.splice(idxOf, 1);
    somethingChanged = true;
  }

  if (somethingChanged) {
    reloadOnTagChange();
  }
  return somethingChanged;  
}

function toggleTag(tag) {
  if (tag == 'All' || selectedTags.indexOf(tag) == -1) {
    return selectTag(tag);
  }
  else {
    return deselectTag(tag);
  }
}

/*
=========================================================================
 upsertTask()
-------------------------------------------------------------------------
 This is called from the modal when the Save button is clicked to 
 manually collect the form elements, create a Task object, and call
 the server to save / create it as specified by the state of the
 modal.  If that wasn't enough, it manually puts the modified task into
 all the right lists depending on which view was passed in.

 Inputs: view - the current page so it knows which lists to change
                TBD: just change all the lists and let the UI take
                     care of itself?

 TBD: do the modal in vue at some point so this doesn't feel like such
      a hack.
========================================================================*/
// note this only adds new tags - it doesn't make
// the tags set on the task "match" the string
// (for example, by removing tags not in the string)
function setTagsFromString(task, strTags) {
  if (strTags.length > 0) {
    var tags = strTags.split("/").map(function(e){return e.trim();});
    tags.map(function(e){return task.addTag(e);});
  }
}

function upsertTask(view) {
	var t = null;
	var f = document.getElementById("newTask");
	var name    = f.elements["task"].value;
  var strdate = f.elements["startdate"].value;
	var strtime = f.elements["starttime"].value;
	var today   = (f.elements["today"].value == "true"); // hidden field useable by every UI
	var time = null;
  var strtags = f.elements["tags"].value;

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
	// console.log("upsertTask: time=" + time);


	// we decide whether or not to set the thisweek flag based on the hidden field
	var thisWeek = (f.elements["thisweek"].value == "true"); // hidden field useable by every UI

	var duration = parseInt(f.elements["duration"].value);
	if (isNaN(duration)) {
		duration = null;
	}
	if (currTask == null) {
		t = new Task(null, name, time, null, duration, false);
    t.setToday(today);
    t.setThisWeek(thisWeek);
    setTagsFromString(t, strtags);
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
 		// console.log("upsertTask: new task id=" + t.id);
 		// tbd: put the id into the local version of the task
	} else {
		t = currTask;
		t.setName(name);
		t.setTargetStartTime(time);
		t.setEstimate(duration);
    // these may be redundant - we should clean up or separate the
    // system tags like today/thisweek from user-defined tags
    t.setToday(today);
    t.setThisWeek(thisWeek);
    setTagsFromString(t, strtags);

		moveTask(t); // adjust the task to show in all the right lists
		currTask = null;
    replaceTask(t); // set all the fields on this task
	}
}

/*
=========================================================================
 deleteTask()
-------------------------------------------------------------------------
 This always deletes the current task, which is always the one in the
 modal in the UI.  It decides not to care about the view, just removing
 the task from all of them (except history where the modal doesn't
 currently display - but probably should and will soon).
========================================================================*/
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


/*
=========================================================================
 findTaskInList()
-------------------------------------------------------------------------
 This heavily used funciton searches the specified list for an id and
 returns either:
   * the task - if it was found and the "task" return type was requested
   * the list - if it was found and the "list" return type was requested
   * null - if it was not found

 Inputs: TaskList list       - the TaskList to search
         string   id         - the id of the task being sought
         string   returnType - 'task' or 'list'
========================================================================*/
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
    found = findTaskInList(inprogress, id, returnType);
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

/*
=========================================================================
 moveTask()
-------------------------------------------------------------------------
 This heavily used function is called when a task is marked completed
 or "unmarked" completed - it moves the task from list to list in the UI
 based on it's previous and new state, and updates the task on the
 server to reflect its new state.
========================================================================*/
function moveTask(task) {
  if (task == null) { return; }

  // console.log("in moveTask(): complete="+task.isComplete());

  if (task.isComplete()) {


    // if it is in the weekly list then move it to the bottom
    if (planWeek.isHere(task)) {
        planWeek.removeTask(task);
        planWeek.insertTask(task);
    }

    // make sure it is in the done list
    done.insertTask(task, false);

    // make now it's completion time
    now = new Date();
    task.setActualCompletionTime(now.toJSON());

    // if it was in the scheduled or stuff lists remove it
    scheduled.removeTask(task);
    stuff.removeTask(task);
    inprogress.removeTask(task);

    // when we update change the state and completion time
    task.dirty = ["state", "actualcompletiontime"];

  } else if (task.isInProgress()) {

    inprogress.insertTask(task, false);
    scheduled.removeTask(task);
    stuff.removeTask(task);
    done.removeTask(task);

    // when we update change the state and completion time
    task.dirty = ["state"];

  // else back to not started
  } else {

    // remove from done list if it was there
    done.removeTask(task);
    inprogress.removeTask(task);

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

    // change the state of the task
    // not sure why double-binding doesn't
    // make this happen automatically
    if (task.isComplete()) {
        task.markNotStarted();
    }
    else {
        task.markComplete();
    }
    // console.log("toggleTask(): complete="+task.complete)

    moveTask(task);
}

// badly named function because we only clear tasks that
// are in the list, marked done, and clear them of the specified tag
function clearTagAndList(list, tagToClear) {
  // only today and thisweek are currently supported
  // once we properly implement tags in the JS Task
  // object we can generalize this better
  if (!(tagToClear == 'today' || tagToClear == 'thisweek')) {
    return;
  }

  var lenClear = list.tasks.length;
  // go backwards so we can remove items from the end
  // and do the loop for both server and UI
  for (var idxClear = lenClear - 1; idxClear >= 0; idxClear--) {
    task = list.tasks[idxClear];
    if (task.isComplete()) {
        console.log(task.name);
        switch (tagToClear) {
            case 'today': clearToday(task);
            case 'thisweek': clearThisWeek(task);
        }
        list.removeTask(task);
    } // only clear the completed tasks
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

function startStop(task) {
    if (task.isInProgress()) {
        task.markNotStarted();
    }
    else if (task.isNotStarted()) {
        task.markInProgress();
    }
    else {
        return;
    }
    moveTask(task);
}


/*
=========================================================================
 Today View - Loading into Various Lists
-------------------------------------------------------------------------
 These two functions make sure various lists that make up the today
 view are populated properly from the server, with each list presumably
 bound to the UI.  It distributes the items into lists based on their
 status and existence of a time (it puts timed items into a calendar).
 One function kicks off the server call and the other is called back
 for each task returned by the server.

 Structure:
  * loadTasksToday() - kicks off ajax call for all 'today' tasks
  * todayTaskIntoLists() - figures where to put each task
========================================================================*/
function todayTaskIntoLists(task) {
  if (task.isComplete()) {
    done.insertTask(task, 'actualendtime');
  }
  else if (task.isInProgress()) {
    inprogress.insertTask(task);
  }
  else {
    if (task.getTargetStartTime() == null) {
      stuff.insertTask(task);
    }
    else {
      scheduled.insertTask(task, 'targetstarttime');
    }
  }
}
function loadTasksToday(tags = null) {
  collectTasks(tasksTodayURL(tags), todayTaskIntoLists)
}


/*
=========================================================================
 Planning View - Loading the Week and Day Lists
-------------------------------------------------------------------------
 These two sets of functions load the "thisweek" and "today" lists
 so they can be displayed side-by-side (will later be augmented with
 thismonth and thisyear and foreever lists???).  They mostly rely on
 the server supporting their system tags of "thisweek" and "today" which
 magically also check their dates and auto-include them even if the tag
 is not set explicitly on them.

 Structure:
  * loadTasksThisDay() - kicks off ajax call for all 'today' tasks
  * loadTasksThisWeek() - kicks off ajax call for all 'thisweek' tasks
  * planTaskIntoWeek() - adds task to proper place in week list
  * planTaskIntoDay() - added task to day list

 TBD - Funny use of a global to track where to put things in the week
       view so completed weekly items auto-populate ar the bottom of the
       list.
========================================================================*/
var planningIdLastIncomplete = null;
function planTaskIntoWeek(task) {
  if (task.isComplete()) {
    planWeek.insertTask(task, 'end');
  }
  else {
    planWeek.insertTask(task, planningIdLastIncomplete!=null?planningIdLastIncomplete:'start');
    planningIdLastIncomplete = task.id;
  } 
}
function loadTasksThisWeek(tags = null) {
  planningIdLastIncomplete = null;
  collectTasks(tasksThisWeekURL(tags), planTaskIntoWeek)
}

function planTaskIntoDay(task) {
  planDay.insertTask(task)
}
function loadTasksThisDay(tags = null) {
  collectTasks(tasksTodayURL(tags), planTaskIntoDay)
}


/*
=========================================================================
 History View - Loading the Currently Selected Day
-------------------------------------------------------------------------
 Given a date, these functions load the history currDay task list global
 with the tasks completed on that day.  The global is typically bound
 to a UI control to show the tasks.  The list is cleared on each run
 so it starts fresh.

 Structure:
  * loadTasksByDay() - kicks off ajax call for all done on that day
  * historyTaskIntoDay() - just adds each found task to the list

 TBD - It seems weird that we're calling tasksFindURL with the date but
       we don't specify here that we only want completed tasks.  Is
       the "find" function currently hard-coded to just completed tasks?
========================================================================*/
function dateToString(date) {
  var result = "";
  if (date != null) {
    result += date.getFullYear();
    result += "-";
    result += (date.getMonth() < 9 ? "0" : "") + (date.getMonth() + 1);
    result += "-";
    result += (date.getDate() < 10 ? "0" : "") + date.getDate();
  }
  return result;
}
function historyTaskIntoDay(task) {
  currday.insertTask(task)
}
function loadTasksByDay(date) {
  currday.clean();
  if (date == null) {
    return;
  }
  collectTasks(tasksFindURL(dateToString(date)), historyTaskIntoDay)

  // selectedDate is bound to vue calendar control so will select the date
  // this is bullshit - not sure why I have to reach into the view model
  // and can't just change my JS Date - but I can't.
  v["selectedDate"] = date;
}


/*
=========================================================================
 History View - Loading the Calendar Control
-------------------------------------------------------------------------
 These functions are useful only to populate the calendar control so th
 user knows which days have completed tasks on them, and which do not so
 she know which to click to find things.  Each task is added to the
 completedTaskDates global which gets bound to the calendar control.
 When all are done we invoke loadTasksByDay() to load the most recent
 day with the actual task display.

 Structure:
  * findAllCompletedTaskDates() - kicks it all off with 2 callbacks
  * historyTaskIntoCalendar() - is called for each collected task
  * historyDoneLoadingTasks() - is called when all tasks are processed

 TBD - These dates are off because of timezones!!!  They used to be
       closer (but still off a little) when I used the raw completion
       time from the server JSON, but when I converted it here I lost
       something.

 TBD - ugly use of global variables for availability across callbacks.
========================================================================*/
var historyUniqueDates = {};
var historyMaxDate = null;
function historyTaskIntoCalendar(task) {
  var timestamp = task.getActualCompletionTime();
  if (timestamp != null) {
    // chopping off time is ignoring timezone and resulting in
    // the wrong date?  Or is it working and "FindByCompletionDate()"
    // is ignoring the TZ?
    var completionDate = new Date(timestamp.toDateString()); 
    if (completionDate != null) {
      if (!historyUniqueDates[completionDate]) {
        historyUniqueDates[completionDate] = true;
        completedTaskDates.push(completionDate);
        if (historyMaxDate == null || completionDate > historyMaxDate) {
          historyMaxDate = completionDate;
        } // if the high water mark wasn't already set
      } // if it isn't already in the list of unique dates
    } // if the date call didn't fail
  } // if we have a completio time
}
function historyDoneLoadingTasks(status) {
  loadTasksByDay(historyMaxDate)
}
function findAllCompletedTaskDates() {
  historyUniqueDates = {}
  historyMaxDate = null
  collectTasks(tasksCompleteURL(), historyTaskIntoCalendar, historyDoneLoadingTasks)
}

/*
=========================================================================
 Tags
-------------------------------------------------------------------------
 Functions to collect all the tags in all the tasks on the server so
 we can present them as choices for the user to filter on.  They are
 stored on a global sorted by most-used (but we drop the counts since
 the UI doesn't actually need them).
========================================================================*/
function tagsDoneFinding(mapTags) {
  // these system tags shouldn't be displayed to the user
  delete mapTags.today  
  delete mapTags.thisweek

  while (allTags.length) {
    allTags.pop()
  }
  allTags.push("All")
  aTags = Object.entries(mapTags)
  aTags.sort(function(a,b){b[1]-a[1]})
  aTags.map(function(e){allTags.push(e[0])})
}
function tagsFindAll() {
  collectTags(tagsDoneFinding)
}

