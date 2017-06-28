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

  constructor(id, name, startTime, estimate, complete) {
    this.id = id;               // id of the task on the server
    this.name = name;           // name of the task
    this.startTime = startTime; // start time or date
    this.estimate = estimate;   // in minutes
    this.state = (complete?TaskState.COMPLETE:TaskState.NOT_STARTED);
    // this.complete = complete;   // true if the task is done - OBSOLETE
  }

  isComplete() {
    return (this.state == TaskState.COMPLETE);
  }

  // return the esimtate as a number even if null
  getEstimate() {
    return (isNaN(this.estimate) ? 0 : this.estimate);
  }

  // return true if we have a start time on this task
  hasStartTime() {
    return (this.startTime != null); // should check type?  typescript?
  }

  // utility function to convert minutes to a more
  // concise format if over an hour or not provided
  // at all.
  static formatMinutes(minutes) {
    if (minutes == null) {
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
    return Task.formatMinutes(this.estimate);
  }

  // return start time as formated string of just the time
  startTimeString() {
    if (this.startTime == null) {
      return "";
    }
    else {
      return Task.formatTime(this.startTime) + ' - ';
    }
  }

} // class Task

