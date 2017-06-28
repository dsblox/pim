

function allowDrop(ev) {
    ev.preventDefault();
}

function drag(ev) {
    ev.dataTransfer.setData("id", ev.target.id);
}

function drop(ev) {
    ev.preventDefault();
    var id = ev.dataTransfer.getData("id");
    dragDropTask(id, ev.target.id)
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
    console.log("this drop will soon complete the task");
    fromList.removeTask(dropped);
    dropped.state = TaskState.COMPLETE;
    toList.insertTask(dropped, idOn);
    updateTask(dropped);
  }

  // if the same and not the calendar then reorder the tasks
  else if (toList == fromList && toList != scheduled) {
    fromList.removeTask(dropped);
    fromList.insertTask(dropped, idOn);
  }

}

