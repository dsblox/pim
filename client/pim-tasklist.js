
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

  // add a task and keep in time order unless requested not to
  insertTask(task, placement = 'end') {

    // interpret placement as either:
    //   - instruction to place at 'end'
    //   - instruction to place in time order
    //   - insutrction to place after task with that 'id'
    var bStartTimeSort = (placement == 'targetstarttime')
    var bDoneTimeSort = (placement == 'actualendtime')
    var bInsertAfter = (!bStartTimeSort && !bDoneTimeSort && placement != 'end')

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

    // otherwise just add to the end
    else {
      this.tasks.push(task);      
    }

  }

  // remove a task from the list
  removeTask(task) {
    var i = this.tasks.indexOf(task);
    if (i >= 0) {
      this.tasks.splice(i,1);
    }
  }

  // find a task by id
  findTask(id) {
    var len = this.tasks.length;
    var result = null;
  	for (var i = 0; i < len && result == null; i++) {
  		if (this.tasks[i].id == id) {
  			result = this.tasks[i];
  		}
  	}
  	return result;
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

}
