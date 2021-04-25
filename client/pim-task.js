function guid() {
  function s4() {
    return Math.floor((1 + Math.random()) * 0x10000)
      .toString(16)
      .substring(1);
  }
  return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
    s4() + '-' + s4() + s4() + s4();
}

var TaskState = {
  NOT_STARTED: 0,
  COMPLETE:    1,
  IN_PROGRESS: 2,
  ON_HOLD:     3
};



// utility date / time functions
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



// create our basic task view model
class Task {

  constructor(id, name, startTime, endTime, estimate, complete) {
    this.id = id;               // id of the task on the server
    this.name = name;           // name of the task
    this.targetStartTime = startTime; // start time or date
    this.actualCompletionTime = endTime; // completed datetime
    this.estimate = estimate;   // in minutes
    // this.today = today;
    // this.thisWeek = thisWeek;
    this.state = (complete?TaskState.COMPLETE:TaskState.NOT_STARTED);
    this.tags = null;
    // this.complete = complete;   // true if the task is done - OBSOLETE
  }

  copy(taskSource) {
    if (taskSource != null) {
      this.id = taskSource.id;
      this.name = taskSource.name;
      this.targetStartTime = taskSource.targetStartTime;
      this.actualCompletionTime = taskSource.actualCompletionTime;
      this.estimate = taskSource.estimate;
      this.state = taskSource.state;
      this.tags = taskSource.getCopyOfTags();
    }
  }

  getId() {
    return this.id;
  }

  hasId() {
    return this.id != null;
  }

  getName() {
    if (this.name && this.name.length > 0) {
      return this.name;
    }
    else {
      return "<unnamed task>";
    }
  }

  setName(newName) {
    this.name = newName;
  }

  getState() {
    return this.state;
  }

  setState(newState) {
    this.state = newState;
  }

  isComplete() {
    return (this.state == TaskState.COMPLETE);
  }

  isNotStarted() {
    return (this.state == TaskState.NOT_STARTED); 
  }  

  isInProgress() {
    return (this.state == TaskState.IN_PROGRESS); 
  }

  markComplete() {
    this.state = TaskState.COMPLETE;
  }

  markNotStarted() {
    this.state = TaskState.NOT_STARTED;    
  }

  markInProgress() {
    this.state = TaskState.IN_PROGRESS;        
  }

  markOnHold() {
    this.state = TaskState.ON_HOLD;  
  }

  tagIndex(tag) {
    if (this.tags == null) {
      return -1;
    }
    return this.tags.indexOf(tag);    
  }

  isTagSet(tag) {
    return (this.tagIndex(tag) >= 0);
  }

  hasAllTags(tags) {
    return tags.every(this.isTagSet, this)
  }

  // returns false if tag is already set
  // TBD: validate tags have no spaces or slashes
  addTag(tag) {
    if (this.tags == null) {
      this.tags = [];
    }
    if (this.isTagSet(tag)) {
      return false;
    }
    this.tags.push(tag);
    return true;
  }

  getTags(separator = null) {
    if (this.tags != null) {
      if (separator != null) {
        return this.tags.join(separator);
      }
      else {
        return this.tags;
      }
    }
    else {
      return "";
    }
  }

  getCopyOfTags() {
    var tagList = []
    if (this.tags != null) {
      this.tags.map(t => tagList.push(t))
    }
    return tagList
  }

  addTagsFromString(strTags) {
    var task = this
    if (strTags.length > 0) {
      var tags = strTags.split("/").map(function(e){return e.trim();});
      tags.map(function(e){return task.addTag(e);});
    }
  }

  // returns false if tag wasn't set
  removeTag(tag) {
    var i = this.tagIndex(tag);
    if (i >= 0) {
      this.tags.splice(i,1);
      return true;
    }
    return false;
  }

  isToday() {
    return this.isTagSet("today");
  }

  setToday(newToday) {
    if (newToday) {
      this.addTag("today");
    }
    else {
      this.removeTag("today");
    }
  }

  isThisWeek() {
    return this.isTagSet("thisweek");
  }

  isThisWeekAndToday() {
    return this.isThisWeek() && this.isToday()
  }

  isThisWeekAndNotToday() {
    return this.isThisWeek() && !this.isToday()
  }

  setThisWeek(newThisWeek) {
    if (newThisWeek) {
      this.addTag("thisweek");
    }
    else {
      this.removeTag("thisweek");
    }
  }

  // return the esimtate as a number even if null
  getEstimate() {
    return (isNaN(this.estimate) ? 0 : this.estimate);
  }

  setEstimate(e) {
    this.estimate = e;
  }

  // return true if we have a start time on this task
  hasStartTime() {
    return (this.getTargetStartTime() != null); // should check type?  typescript?
  }

  getTargetStartTime() {
    return this.targetStartTime;
  }

  setTargetStartTime(newTime) {
    this.targetStartTime = newTime;
  }

  hasCompletionTime() {
    return (this.getActualCompletionTime() != null); // should check type?  typescript?
  }

  getActualCompletionTime() {
    return this.actualCompletionTime;
  }

  setActualCompletionTime(newTime) {
    this.actualCompletionTime = newTime;
  }

  // set targetStartTime from well formatted date / time strings
  setTargetStart(strdate, strtime) {
    var time = null;
    if (strtime && strtime.length > 0) {
      // if date not specified, assume today    
      if (!strdate || strdate.length == 0) {
        var today = new Date();
        strdate = today.getFullYear()+'-'+(today.getMonth()+1)+'-'+today.getDate();     
      }
      // remember time will be sent in local time zone
      // with TZ info and will be stored in GMT   
      time = new Date(strdate + " " + strtime);
    }

    if ((strtime && strtime.length > 0) && (strdate && strdate.length > 0)) {
      // remember time will be sent in local time zone
      // with TZ info and will be stored in GMT
      time = new Date(strdate + " " + strtime);
    }

    // set the result - whatever it is on the object
    this.setTargetStartTime(time);
    return time;
  }

  // utility function to convert minutes to a more
  // concise format if over an hour or not provided
  // at all.
  static formatMinutes(minutes) {
    if (minutes == null || minutes == 0) {
      return "";
    }
    else if (minutes <= 60) {
      return minutes;
    }
    else {
      var hours = Math.floor(minutes / 60);
      var minleft = minutes - (hours * 60);
      return hours + ":" + (minleft<=9?"0":"") + minleft;
    }
  }

  // utility function to format a time from a date object
  static formatTime(date) {
    if (date == null) {
      return "";
    }
    var hr = date.getHours();
    var min = date.getMinutes();
    if (min < 10) {
      min = "0" + min;
    }
    var ampm = null;
    if (hr < 12) {
    	ampm = 'am';
      if (hr == 0) {
        hr = 12;
      }
    } else {
    	ampm = 'pm';
    	if (hr >= 13) {
    		hr -= 12
    	}
    }
    return hr + ":" + min + ampm;
  }

  // return the estimated minutes in readable format
  estimateString() {
    return Task.formatMinutes(this.getEstimate());
  }

  // return start time as formated string of just the time
  startTimeString() {
    if (this.getTargetStartTime() == null) {
      return "";
    }
    else {
      return Task.formatTime(this.getTargetStartTime()) + ' - ';
    }
  }

  // these were from the modal needs, above may be for something else
  // we should get rid of one of them
  justStartTime() {
    return extractTimeString(this.getTargetStartTime())
  }
  justStartDate() {
    return extractDateString(this.getTargetStartTime())
  }

} // class Task

