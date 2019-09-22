var nSecPerMinute = 60000000000;

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

// create our basic task view model
class Task {

  constructor(id, name, startTime, endTime, estimate, complete, today, thisWeek) {
    this.id = id;               // id of the task on the server
    this.name = name;           // name of the task
    this.targetStartTime = startTime; // start time or date
    this.actualCompletionTime = endTime; // completed datetime
    this.estimate = estimate;   // in minutes
    this.today = today;
    this.thisWeek = thisWeek;
    this.state = (complete?TaskState.COMPLETE:TaskState.NOT_STARTED);
    // this.complete = complete;   // true if the task is done - OBSOLETE
  }

  getName() {
    return this.name;
  }

  setName(newName) {
    this.name = newName;
  }

  isComplete() {
    return (this.state == TaskState.COMPLETE);
  }

  isToday() {
    return this.today;
  }

  setToday(newToday) {
    this.today = newToday;
  }

  isThisWeek() {
    return this.thisWeek;
  }

  setThisWeek(newThisWeek) {
    this.thisWeek = newThisWeek;
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

} // class Task

