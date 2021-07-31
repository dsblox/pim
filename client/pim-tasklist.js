
// class list holds a list of tasks
class TaskList {
  constructor() {
    this.tasks = [];
    this.title = "";
    this.id = "";
  }

  title() {
  	return this.title;
  }

  setTitle(newTitle) {
  	this.title = newTitle;
  }

  getId() {
  	return this.id;
  }

  setId(newId) {
  	this.id = newId;
  }

  filterToTag(tag) {
    this.tasks.map(function(t){if(t.isTaskSet(tag)) { this.removeTask(t);} });
  }

  // make this look like JavaScript array filter but return a tasklist
  filter(callback, thisarg) {
    var newlist = new TaskList()
    newlist.tasks = this.tasks.filter(callback, thisarg)
    return newlist
  }
  sort(callback, thisarg) {
    var newlist = new TaskList()
    newlist.tasks = this.tasks.sort(callback, thisarg)
    return newlist    
  }

  find(callback, thisarg) {
    return this.tasks.find(callback, thisarg)
  }

  findIndex(callback, thisarg) {
    return this.tasks.findIndex(callback, thisarg)
  }

  findTaskIndex(task) {
    return this.findIndex(t => t.getId() === task.getId())
  }

  // add a task and keep in time order unless requested not to
  insertTask(task, placement = 'end') {

    // interpret placement as either:
    //   - instruction to place at 'end'
    //   - instruction to place in time order
    //   - insutrction to place after task with that 'id'
    var bStartTimeSort = (placement == 'targetstarttime')
    var bDoneTimeSort = (placement == 'actualendtime')
    var bStart = (placement == 'start')
    var bInsertAfter = (!bStartTimeSort && !bDoneTimeSort && placement != 'end' && placement != 'start')

	  // don't allow the same task 2x in this list
  	if (this.findTask(task.id) != null) {
  		return; 
  	}

    // if time sort by start time needed run the list to insert
    if (bStartTimeSort && task.hasStartTime()) {
      var list = this.tasks;
      var i = 0;
      while (i < list.length && task.getTargetStartTime() > list[i].getTargetStartTime()) {
        i++;
      }

      // if i'm off the end then add to the end otherwise insert
      if (i >= list.length) {
        list.push(task);
      } else {
        list.splice(i, 0, task);
      }    
    }

    // if time sort by end time needed run the list to insert
    else if (bDoneTimeSort && task.hasCompletionTime()) {
      var list = this.tasks;
      var i = 0;
      while (i < list.length && task.getActualCompletionTime() > list[i].getActualCompletionTime()) {
        i++;
      }

      // if i'm off the end then add to the end otherwise insert
      if (i >= list.length) {
        list.push(task);
      } else {
        list.splice(i, 0, task);
      }    
    }

    // if to be added after a specific task number
    else if (bInsertAfter) {
      var list = this.tasks;
      var i = 0;
      while (i < list.length && placement != list[i].id) {
        i++;
      }

      // if i'm off the end then add to the end otherwise insert
      if (i >= list.length) {
        list.push(task);
      } else {
        list.splice(i, 0, task);
      }    
    }

    else if (bStart) {
      this.tasks.splice(0, 0, task);
    }

    // otherwise just add to the end
    else {
      this.tasks.push(task);      
    }

  }

  // remove a task from the list
  removeTask(task) {
    // if there is no id all we can do is find the same
    // instance of the task, but if we have an id we can
    // look up the task by id - even if not the same instance
    // of the task that is in the list.
    t = task;
    if (task.hasId()) {
      t = this.findTask(task.getId())
    }
    var i = this.tasks.indexOf(t);
    if (i >= 0) {
      this.tasks.splice(i,1);
    }
  }

  // see if a task is in the list
  isHere(task) {
    return (this.findTask(task.id) != null);
  }

  // find a task by id
  findTask(id) {
    var result = this.tasks.find(t => t.id == id )
    return result==undefined?null:result
  }

  // recompute the estimated duraction of the entire list
  duration() {
    if (this.tasks == null || this.tasks.length == 0) {
      return 0;
    }
    var total = 0;
    var len = this.tasks.length;
    for (var i = 0; i < len; i++) {
      total += this.tasks[i].getEstimate();
    }
    return total;    
  }

  // format the duration as hours and minutes hh:mm
  durationFormatted() {
    var minutes = this.duration();
    return Task.formatMinutes(minutes);
  }  

  numTasks() {
    return this.tasks.length;
  }

  clean() {
    this.tasks = [];
  }

  copy(target) {
    target.clean();
    var len = this.tasks.length;
    for (var i = 0; i < len; i++) {
      target.insertTask(this.tasks[i]);
    }
    return target;
  }

  getTaskByIndex(i) {
    if (i >=0 && i < this.numTasks()) { 
      return this.tasks[i];
    }
    else {
      return null;
    }
  }

  getAsText() {
    let strings = this.tasks.map(t => t.getAsText())
    return strings.join('\n')
  }

  getAsHTML() {
    let strings = this.tasks.map(t => '<li>' + t.getAsHTML())
    return "<ul>" + strings.join('') + "</ul>"
  }  

  async clipboardCopy() {
    try {
      const textType = 'text/plain'
      const textBlob = new Blob([this.getAsText()], { type: textType })
      const htmlType = 'text/html'
      const htmlBlob = new Blob([this.getAsHTML()], { type: htmlType })
      let data = [new ClipboardItem({ [htmlType]: htmlBlob, [textType]: textBlob })]
      await navigator.clipboard.write(data)
    } catch (err) {
      console.log('clipboardCopy: copy failed with ' + err)
    }
  }

}
