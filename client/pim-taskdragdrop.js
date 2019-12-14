

function allowDrop(ev) {
    ev.preventDefault();
}

// TBD: this currently works only if you drag and drop the <a> tag that
// is the task name itself.  We need the task-id to be set on any element
// that might happen to be dropped from / to such as the checkbox itself
// or the estimate glyph
function drag(ev) {
    var id = ev.target.id;
    ev.dataTransfer.setData("id", id);
}

function drop(ev) {
    console.log(ev);

    ev.preventDefault();

    var from_id = ev.dataTransfer.getData("id");
    var to_id = ev.target.id;
    // console.log(from_id);
    // console.log(to_id);
    dragDropTask(from_id, to_id);
}

// for now we have no way to drop something to the end of a list
// drag/drop functions allowed:
//   drag from another list to DONE to mark an item completed
//   drag within the STUFF or DONE list to reorder
function dragDropTask(idDropped, idOn) {

  // find which list each item is in
  var fromList = listOfTask(idDropped);
  var toList = listOfTask(idOn);
  var dropped = fromList.findTask(idDropped);

  // validate our inputs
  if (fromList == null || toList == null || dropped == null) {
    console.log("dropDropTask: invalid input ids provided");
    return;
  }

  // if to/from are different and target list is "DONE" then complete the task
  if (toList == done && toList != fromList) {
    // complete the task
    fromList.removeTask(dropped);
    dropped.state = TaskState.COMPLETE;
    toList.insertTask(dropped, idOn);
    dropped.dirty = ["state"];
    updateTask(dropped);
  }

  // if the same and not the calendar then reorder the tasks
  else if (toList == fromList && toList != scheduled) {
    fromList.removeTask(dropped);
    fromList.insertTask(dropped, idOn);
  }

}

