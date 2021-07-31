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
=========================================================================
 Navigation Menu
-------------------------------------------------------------------------
 Our navigation component can be passed a menu, and this is what they
 all use.  Soon we'll likely go SPA so this will only be used once.
=======================================================================*/
var mainMenu = [
  { display: 'Home',     route: "vuepim.html"  },
  { display: 'Planning', route: "planning.html"},
  { display: 'History',  route: "whatidid.html"},
  { display: 'Mission',  route: "#"            },
]

/*
=========================================================================
 Alert
-------------------------------------------------------------------------
 Assumes the root Vue app instance is in a global named "v" and this
 function simply invokes the error box by setting it's message.  The
 rest of the work is done in the vue-app and pim-alert component.
 Keeping this abstraction for easy use from throughout even those it
 has the ugliness of the knowledge of the root vue instance name.
=======================================================================*/
function pimShowError(message) {
  v.warn(message)
}


/*
--------------------------------------------------------------------------
 TODAY page - load all tasks and let the component lists choose
------------------------------------------------------------------------*/
var todaylist = new TaskList();

/*
--------------------------------------------------------------------------
 PLANNING page - load all tasks and let the component lists choose
------------------------------------------------------------------------*/
var planninglist = new TaskList();

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
 a tag is clicked).  Note that the UI components are simply bound to
 the selectedTags array and automatically refresh themselves as tags are
 added or removed from the array.

 Note that tag-toggling is a UI-only concept and requires no AJAX
 AJAX calls - which means we may want to consider moving this
 functinonality into a Vue component.

 Inputs: tag - user-defined tag to add or remove from selected tags
========================================================================*/
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
 modalTaskSave()
-------------------------------------------------------------------------
 This is called from the modal when the Save button is clicked to 
 handle any automatic system tags, call the server to save / create the
 task and make sure it ends up in the list that was given to the modal
 when it was invoked.
 
 Inputs: modalTask - the task to be written to the server.  If it has an
                     id we will replace on the server, if not we will
                     create a new task
         list      - the list to add the task to - typically the list
                     for the complete list of tasks for the current page.
         sysTags   - a list of tags that should be auto-added to any
                     task created or edited on this modal.  Used to make
                     sure all new 'today' tasks have the 'today' tag
                     (for example).

 TBD: consider moving more of this function into the vue component(s)?
========================================================================*/
function modalTaskSave(modalTask, list, sysTags) {
  var t = modalTask;

  if (sysTags != null) {
    sysTags.forEach(function(sysTag) { t.addTag(sysTag) })
  }

  // write the task to the server
  createOrReplaceTask(t)

  // add the task to the list the modal wants to use
  if (list != null) {
    list.insertTask(t)
  }

  return t;
}

/*
=========================================================================
 deleteTask() - this is now in the new April 2021 approach
-------------------------------------------------------------------------
 This is called from the UI components when a task should be deleted.
 Today it simply invokes the server to delete the task.  Someday it
 may be enhanced to mark the task as dirty and synchronize at some
 interval (which is why we have this layer here instead of calling the
 server directly from the UI components).
========================================================================*/
function deleteTask(task) {
	if (task == null) {
		return;
	}

  // delete the task from the server (when we write it)
  killTask(task);
}


/*
=========================================================================
 changeTaskState() - this is the new approach April 2021!!!
-------------------------------------------------------------------------
 This function is invoked anytime the state of a task changes, with
 possible states being not-started, in-progress, or complete.  This
 function takes care of marking the completion time of the task (or
 clearing it if a completed task is marked incomplete).

 For now, we do real-time notification of the server as we do in all
 places, but in the future we may choose to collect "dirty" tasks here
 and perform intermittent synchronization to the server.

 Inputs: task - whose state has changed
 Result: task is written to the server as needed to persist the change
=======================================================================*/
function changeTaskState(task) {
  if (task == null) { return; }

  if (task.isComplete()) {
    now = new Date();
    task.setActualCompletionTime(now);

    task.dirty = ["state", "actualcompletiontime"];    
  }
  else if (task.isInProgress() || task.isNotStarted()) {
    if (task.hasCompletionTime()) {
      task.setActualCompletionTime(null); // clear completion time since not really done
      task.dirty = ["state", "actualcompletiontime"];    
    }
    else {
      task.dirty = ["state"];          
    }
  }

  // call the server to update the task persistently
  // only changing the "dirty" elements
  updateTask(task);    
}

function moveInArray(arr, from, to) {

  // Make sure a valid array is provided
  if (Object.prototype.toString.call(arr) !== '[object Array]') {
    throw new Error('Please provide a valid array');
  }

  // Delete the item from it's current position
  var item = arr.splice(from, 1);

  // Make sure there's an item to move
  if (!item.length) {
    throw new Error('There is no item in the array at index ' + from);
  }

  // Move the item to its new position
  arr.splice(to, 0, item[0]);

};

function moveBeforeTask(list, from, to) {

  // reorder the list items
  list.removeTask(from)
  list.insertTask(from, to.getId())

  // call server to reorder
  reorderTask(from, to)
}

/*
=========================================================================
 clearTagFromList() - this is the new approach April 2021!!! but isn't fully clean yet
-------------------------------------------------------------------------
 This function is invoked from a task-list when the user has asked to
 clear a particular tag from all items in the list.  This is typically
 used for a system tag to "archive" a bunch of items off the screen,
 such as clearing your "done" list on the today view to get ready for
 the next day.

 For now, we do real-time notification of the server as we do in all
 places, but in the future we may choose to collect "dirty" tasks here
 and perform intermittent synchronization to the server.

 Inputs: list          - of tasks that want the tag removed
         tagToClear    - this tag should be taken off all tasks
         onlyCompleted - a safety feature since usually only completed
                         tasks will have a tag cleared

 Result: tag is cleared from all tasks, all tasks are persisted
=======================================================================*/
function clearTagFromList(list, tagToClear, onlyCompleted) {
  list.tasks.forEach(function(t) { 
    var attemptToRemove = (t.isComplete() || !onlyCompleted)
    if (attemptToRemove && t.removeTag(tagToClear)) {
      t.resetTags = [tagToClear]
      t.setTags = []
      updateTask(t)
    }
  } )
}

/*
=========================================================================
 writeToday - in April 2021 approach
-------------------------------------------------------------------------
 This function is specifically for the today tag when it is turned on
 or off on a task to write the change to the server.  Today it simply
 writes the change, but in the future it may enqueue changes and sync
 intermittently, which is why we haven't moved persistence into the 
 UI components.

 Input:  task that has been changed to have today on/off
 Result: the current state of the today tag is written to the server
========================================================================*/
// called from vue component when "move to/from today" control clicked
function writeToday(task) {
  if (task.isToday()) {
    task.setTags = ['today'];
  }
  else {
    task.resetTags = ['today'];
    task.setTags = [];
  }
	updateTask(task);
}



/*
=========================================================================
 Today View - Loading into The One List
-------------------------------------------------------------------------
 This function calls the server to get all the "today" view tasks, and
 provides as its callback the todayTasks() function which is called for
 each task and drops it into the single todaylist which is used for the
 entire page.  This used to drop into separate lists for each UI
 component, but we made the components smart enough to pick their own
 subsets from the larger list.

 Structure:
  * loadTasksToday() - kicks off ajax call for all 'today' tasks
  * todayTasks()     - puts it into the list
========================================================================*/
function todayTasks(eachTask) {
  todaylist.insertTask(eachTask)
}
function loadTasksToday(tags = null) {
  collectTasks(tasksTodayURL(tags), todayTasks)
}


/*
=========================================================================
 Planning View - Loading the Week and Day Lists
-------------------------------------------------------------------------
 We're in the process of converting this code into smarter UI components
 that select the tasks they care about by looking at the state and
 tags.  Right now, we will LOSE the functionality to include a task in
 one of our UI lists based on their date - we'll only look at the
 explicit tags set on them.  We could:
  - fake the thisweek or today tags based on the dates for the UI
  - enhance the UI controls to "match" on dates as well as tags

 Still, it works pretty well just with the tags at the moment.  Note that
 the UI will not show everything the server returns because of this.

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
       view so completed weekly items auto-populate at the bottom of the
       list.  This functionality may not be working well with the new
       UI components so may need to be fixed.
========================================================================*/
var planningIdLastIncomplete = null;
function planTaskIntoWeek(task) {
  if (task.isComplete()) {
    planninglist.insertTask(task, 'end');
  }
  else {
    planninglist.insertTask(task, planningIdLastIncomplete!=null?planningIdLastIncomplete:'start');
    planningIdLastIncomplete = task.id;
  } 
}
function loadTasksThisWeek(tags = null) {
  planningIdLastIncomplete = null;
  collectTasks(tasksThisWeekURL(tags), planTaskIntoWeek)
}

function planTaskIntoDay(task) {
  planninglist.insertTask(task)
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
