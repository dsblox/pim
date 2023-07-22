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
  UNSPECIFIED:-1,
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

// note - this still doesn't check that the date is valid
// ie - a Date() object can be created with in invalid date ???
function isDate(newTime) {
  if (newTime != null && !(newTime instanceof Date)) {
    console.log('setTargetStartTime() - caller provided non-date')
    return false
  } else {
    return true
  }
}

function hasNOTTagPrefix(tag) {
  return (tag.charAt(0) == '!')
}

function isLegalTag(tag) {
  return !hasNOTTagPrefix(tag)
}



class Hyperlink {
  constructor(url, label, startIndex, endIndex) {
    this.url = url
    this.label = label
    this.start = startIndex
    this.end = endIndex
  }

  isIndexed() {
    return (this.start != -1 && this.end != -1)
  }

  compare(another) {
    if (this.start < another.start) {
      return -1
    }
    if (this.start > another.start) {
      return 1
    }
    return 0
  }

  getLabel() {
    return this.label?this.label:"^"
  }
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
    this.tags = []; // must start empty vs. null so we can bind to it
    this.links = []; // must start empty vs. null so we can bind to it
    // this.complete = complete;   // true if the task is done - OBSOLETE
  }

  clone(taskSource) {
    this.id = null
    // console.log('task.clone(): taskSource='+taskSource)
    if (taskSource != null) {
      this.name = taskSource.name;
      this.targetStartTime = taskSource.targetStartTime;
      this.actualCompletionTime = taskSource.actualCompletionTime;
      this.estimate = taskSource.estimate;
      this.state = taskSource.state;
      this.tags = taskSource.getCopyOfTags();
      this.links = taskSource.getLinks();
    }
  }

  copy(taskSource) {
    if (taskSource != null) {
      this.clone(taskSource)
      this.id = taskSource.id;
    }
  }

  getId() {
    return this.id;
  }

  hasId() {
    return this.id != null;
  }

  // get the name and if there is a single link and it is requested
  // then return the name as a hyperlink to that one link
  getName(html = false) {
    let result = null
    if (this.name && this.name.length > 0) {
      result = this.name;
    }
    else {
      result = "<unnamed task>";
    }

    let links = this.getLinks()
    if (html && links.length == 1) {
      result = "<a href=\"" + links[0].url + "\">" + this.name + "</a>"
    }
    return result
  }

  setName(newName) {
    this.name = newName;
  }

  getState() {
    return this.state;
  }

  // for UI sorting reorder state enum as follows
  //   1 = not started
  //   2 = in progress
  //   3 = on hold
  //   4 = complete
  // This was sort functions can compare on this "natural" ordering
  getSortState() {
    let result = 0
    switch (this.state) {
      case TaskState.NOT_STARTED: result = 1; break;
      case TaskState.IN_PROGRESS: result = 2; break;
      case TaskState.ON_HOLD:     result = 3; break;
      case TaskState.COMPLETE:    result = 4; break;
      default:                    result = 0; break;
    }
    return result;
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
    return this.tags.indexOf(tag)  
  }

  isTagSet(tag) {
    return (this.tagIndex(tag) >= 0)
  }

  isTagNotSet(tag) {
    return (this.tagIndex(tag) < 0)
  }

  isSysTagSet(tag) {
    if (tag == "dontforget") {
      return (!this.isTagSet("thisweek") && !this.isTagSet("today") && !this.isTagSet("reuse"))
    }
    else {
      return this.isTagSet(tag)
    }
  }

  hasAllTags(tags) {
    return tags.every(this.isTagSet, this)
  }

  doesntHaveTags(tags) {
    return tags.every(this.isTagNotSet, this)    
  }

  // return true if all tag-specs provided match the tags
  // set on this task.  Any provided tags prefixed with !
  // must NOT be on the task, any provided tags without a
  // prefix MUST BE ON the task for this function to return
  // true.
  matchTags(tags) {
    let stillAMatch = true
    for (let i = 0; stillAMatch && i < tags.length; i++) {
      let tag = tags[i]
      if (hasNOTTagPrefix(tag)) {
        stillAMatch = this.isTagNotSet(tag.substring(1))
      }
      else {
        stillAMatch = this.isTagSet(tag)
      }      
    }
    return stillAMatch
  }

  // returns false if tag is already set
  // TBD: validate tags have no spaces or slashes
  addTag(tag) {
    if (this.tags == null) {
      this.tags = []
    }
    if (this.isTagSet(tag)) {
      return false
    }
    if (!isLegalTag(tag)) {
      return false // error - a tag can't start with !
    }
    this.tags.push(tag);
    return true;
  }

  getTags(separator = null, except = null) {
    if (this.tags != null) {

      // if tags-to-exclude specified then exclude them
      let tagsToUse = this.tags
      if (except != null) {
        tagsToUse = this.tags.filter(t => !except.includes(t))
      }

      // if separator specified then I want the tags as a string
      if (separator != null) {
        return tagsToUse.join(separator)
      }
      else {
        return tagsToUse
      }
    }
    else {
      return ""
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
    if (isDate(newTime)) {
      this.targetStartTime = newTime;
    }
  }

  hasCompletionTime() {
    return (this.getActualCompletionTime() != null); // should check type?  typescript?
  }

  getActualCompletionTime() {
    return this.actualCompletionTime;
  }

  setActualCompletionTime(newTime) {
    if (isDate(newTime)) {
      this.actualCompletionTime = newTime;
    }
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

  /*
  ===============================================================
   string()
  ---------------------------------------------------------------
   Inputs: html (default false) - include HTML hyperlinks

   Return the task in string form for easy display in either HTML
   or as text.  If HTML is specified the only change is the
   addition of hyperlinks in the form of <a> tags.  Note that
   hyperlinks are not included at all in the plain text version
   of the task-as-string.

   Basic format of the output is:
      <time>: <name> (<duration>) [links]

   Hyperlinks are formatted as follows based on the number of
   links associated with the task:
     * no hyperlinks - returns same a plain text
     * 1 hyperlink   - link the entire name of the task
     * >1 hyperlink  - include each link after duration

   TBD: Support for name offsets to link multiple hyperlinks
   within the name.  Holding off because although coded in link
   object there is not yet support in the API or UI for this
   anyway (8/21/21)
  =============================================================*/  
  string(html = false) {
    // magic: in the case of html and 1 hyperlink, getName() creates the link
    let result = this.startTimeString() + this.getName(html)
    let duration = this.estimateString()
    if (duration) {
      result += " (" + duration + ")"
    }

    // in the case of multiple hyperlinks we append them to the end
    if (html && this.getLinks().length > 1) {
      result += this.getLinks().map(l => "&nbsp;<a href=\"" + l.url + "\">" + l.getLabel() + "</a>")
    }

    return result
  }

  getAsText() {
    return this.string()
  }

  getAsHTML() {
    return this.string(true)
  }

  /*
  ===============================================================
   addLink()
  ---------------------------------------------------------------
   Links allow one or more hyperlinks to be associated with a
   a task, and optionally allows you to hyperlink within the
   current name of the task (TBD: how will we handle adjusting
   the link offsets into the name when the name is editted?)

   The intended use of links is to allow the UI to handle them
   flexibly in one of two ways:
     1. a list of one or more hyperlinks associated with the task
     2. embedded into the name of the task, with the idea being
        that portions of the name make sense to be hyperlinked
        (e.g. - Respond to <email> about <document>)

   We support this by supporting two ways to specify the link's
   labels, numbered by use cases above:
     1. provide a link label with each link OR
     2. provide indexes into the task's name

   So there are two ways to get links back off a task:
     1. getLinks() - provides a list of links and labels
     2. getLinkedHTMLName() - provides a DIV tag with the name
        and embedded A tags to the links.

   HOWEVER: as of 8/16/21 we've not built the UI to manage
   multiple links.  Well, kinda.  We've built a pim-link 
   component that can (poorly) show a menu of links when set
   on the task, but we have no UI in the modal or elsewhere to
   manage the list of links.  So for now the modal always
   maintains a single link - even though the underlying JS and
   the server can maintain a list of links.  Also, as of today
   we have no implemented a way to pass the text offsets
   described above to the server.  So for now we support one
   link per task, shown as an icon next to the task when it is
   there.
  =============================================================*/
  getLinks() {
    if (typeof this.links == 'undefined') {
      this.links = []
    }
    return this.links
  }

  addLink(url, label = null, startIndex = -1, endIndex = -1) {

    // for now the only illegal character to reject is single
    // quote since it breaks our yaml persistence - but perhaps
    // in the future escape the url (encodeURIComponent)?
    if (!url || url.indexOf("'") != -1) {
      console.log("task.addLink() - unable to add URL " + url + " to task " + this.name)
      return // ugh - no easy way to indicate an error to caller
    }


    // create the list of links if none exists
    if (typeof this.links == 'undefined') {
      this.links = []
    }

    // create a hyperlink object to manage and add it
    let link = new Hyperlink(url, label, startIndex, endIndex)

    // TBD: don't allow multiple links to overlap - reject or
    // adjust a link that would overlap

    // add it to my array of links
    this.links.push(link)
  }

  // clean out all the links - used for now since we don't
  // have good UI to manage multiple links so we needed a
  // way to keep the list of links clean
  clearLinks() {
    // splice to empty so vue doesn't lose its bindings
    this.links.splice(0, this.links.length)
  }

  /*
  =====================================================
   hasEmbeddedHyperlinks()
  -----------------------------------------------------
   Determines if this task has any hyperlinks that are
   "indexed" - meaning that the user intends for them
   to be linked within the display of the name of the
   task.  Note that it is possible for links to exist
   and not be embedded, in which case they are not
   displayed in the name, but rather are displayed in
   some other manner (tbd).
  ===================================================*/
  hasEmbeddedHyperlinks() {
    if (typeof this.links == 'undefined') {
      return false
    }
    else {
      return this.links.reduce(function(isIndexed, curr) {
                                  if (!isIndexed) {
                                    isIndexed = curr.isIndexed()
                                  }
                                  return isIndexed
                                }, 
                                false)
    }
  }


  /*
  ===============================================================
   getLinkedHTMLName()
  ---------------------------------------------------------------
   Returns a <span> with the task name and embedded HTML <a> tags 
   for each link stored on this task in the proper positions.  
   Note that this returns NULL if there are no links in the name, 
   indicating to the caller that they are free to link the name 
   in other ways if they so desire.

   Note that if links are overlapping, then (as coded today) the
   first link will be included, and anything overlapping will
   simply be silently skipped.

   UNTESTED AS OF 8/7/21!
  =============================================================*/
  getLinkedHTMLName() {
    if (this.hasEmbeddedHyperlinks()) {
      const taskName = this.getName()
      const orderedLinks = this.links.sort((a,b)=>{return a.compare(b)})
      let linkedName = "<span>"
      let currLen = 0
      for (let i = 0; i < this.links.length; i++) {
        curr = this.links[i]
        if (curr.isIndexed()) {
          if (curr.start > currLen) {
            linkedName += taskName.substring(currLen, curr.start)
            linkedName += '<a href="' + curr.url + '">'
            linkedName += taskName.substring(curr.start, curr.end)
            linkedName += '</a>'
            currLen = curr.start
          } // if need to "fill in" from last link (also checks for overlap)
        } // if the link has indexes into the name
      } // for each link

      // add anything not linked from the name to the end and close out
      linkedName += taskName.substring(currLen)
      linkedName += "</span>"
      return linkedName
    }
    else {
      return null
    }
  }

} // class Task

