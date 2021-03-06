/*
==============================================================================
 PIM Vue Components
------------------------------------------------------------------------------
 This is our component library for the application.  It is built 
 hierarchically with the main components being:
  pim-task-list - binds to a TaskList
  pim-task      - beings to a Task
  pim-modal     - allows creation / editing of task attributes

 Lots of other sub-components represent user controls displaying attributes
 and state of the tasks, and are grouped into the previous two for display.
============================================================================*/

/*
======================================================================
 pim-start-time
----------------------------------------------------------------------
 Inputs: task    - the task whose start time we should display
         
 Standard display of start time in its own component on the
 expectation we'll be doing more with it over time.
====================================================================*/
Vue.component('pim-start-time', {
  props: ['task'],
  template: '<span class="text-dark small" v-if="task.hasStartTime()"><strong>{{task.startTimeString()}}</strong></span>'
})

/*
======================================================================
 pim-task-name
----------------------------------------------------------------------
 Inputs: task    - the task whose name we should display
 Events: clicked - let my parent know the name was clicked
         
 We provide the task as a property and pull the name, frankly making
 it easier for my parent to blindly pass the clicked event up the
 chain with the task that will need to be edited.
====================================================================*/
Vue.component('pim-task-name', {
	props: ['task'],
  methods: {
    clicked: function() {
      this.$emit('clicked', this.task);
    }
  },
  template: '<a href="#" v-on:click="clicked" class="task-name" :id="task.id">{{task.getName()}}</a>'
})

/*
======================================================================
 pim-duration
----------------------------------------------------------------------
 Inputs: duration - as a string in the way you want to show duration
         
 Standardize the display of a time estimate or duration of a task.
 For now we're keeping this simple and just showing the string
 provided.  Future: maybe provide the task and encapsulate the
 knowledge of "string-ifying" the estimate in here as well?
====================================================================*/
Vue.component('pim-duration', {
	props: ['duration'],
	template: '<span v-if="duration" class="badge badge-secondary ml-1">{{duration}}</span>' 
})

/*
======================================================================
 pim-task-settoday / pim-task-resettoday
----------------------------------------------------------------------
 Inputs: task - task that have it's today tag added / removed
         
 These components either set or reset the today tag on the task
 provided using the > or < icons.  These are intended for use in 
 planning views when taking a task from a longer timeframe (for now
 weekly) and adding/removing from a shorter one (for now today).

 TBD: if we keep this component knowledgeable about its task, it
 could probably decide for itself whether to display, rather than
 have it's parent do it as it does today.  But perhaps the parent
 should really be doing ALL the work anyway.
====================================================================*/
Vue.component('pim-task-settoday', {
  props: ['task'],
  methods: {
    settoday: function() {
      if (!this.task.isToday()) {
        this.task.setToday(true);
      }
      writeToday(this.task);
    }
  },
  template: '<span class="fa fa-chevron-right" v-on:click="settoday"></span>'
})

Vue.component('pim-task-resettoday', {
  props: ['task'],
  methods: {
    resettoday: function() {
      if (this.task.isToday()) {
        this.task.setToday(false);
      }
      writeToday(this.task);
    }
  },
  template: '<i class="fa fa-chevron-left" v-on:click="resettoday"></i>'
})

/*
======================================================================
 pim-startstop
----------------------------------------------------------------------
 Inputs: task - task what should be paused/resumed
         
 Toggles a task from in-progress to not-started using play / pause
 buttons.
====================================================================*/
Vue.component('pim-startstop', {
  props: ['task'],
  methods: {
    startstop: function() {
      // adjust the task to its new state of
      // either in-progress or not-started
      if (this.task.isInProgress()) {
        this.task.markNotStarted()
      } else {
        this.task.markInProgress();
      }

      // tell the outside world to persist this
      changeTaskState(this.task)
    }
  },
  template: '<span v-if="this.task.isInProgress()" class="fa fa-pause" v-on:click="startstop"></span> \
             <span v-else-if="!this.task.isComplete()" class="fa fa-bolt" v-on:click="startstop"></span>'
})

/*
======================================================================
 pim-add
----------------------------------------------------------------------
 Inputs: none
 Events: newtask  - the add-tag icon was clicked
         
 Just standardizes how I display the "add a new task" button.  The
 parent takes all action when clicked.
====================================================================*/
Vue.component('pim-add', {
	template: '<a href="#" class="text-white" v-on:click="$emit(\'newtask\')"> + </a>'
})

/*
======================================================================
 pim-clear
----------------------------------------------------------------------
 Inputs: none
 Events: clear - the clear-tasks icon was clicked
         
 Just standardizes how I display the "clear" button which can be used
 to clear a task list of it's items.  The parent takes all action when 
 clicked.
====================================================================*/
Vue.component('pim-clear', {
  template: '<span class="fa fa-archive" v-on:click="$emit(\'clear\')"></span>',
})

/*
======================================================================
 pim-tag
----------------------------------------------------------------------
 Inputs: label    - the string that represents this tag
         selected - show this tag as active
         xicon    - show this tag with a "x" so it can be "removed"
         task     - which task should the tag be removed from

 Events: toggle - the tag name was clicked
         remove - the x-icon on the tag was clicked
         
 Display a tag as a tile (bootstrap badge), display myself as active
 or not depending on input property, and let my parent know when
 i've been clicked on.
====================================================================*/
Vue.component('pim-tag', {
  props: ['label', 'xicon', 'active'],
  template: '<span><div \
                class="badge" \
                v-bind:class="{ \'badge-dark\': active, \'badge-light\': !active }" \
                @click="labelClick">{{ label }} \
                <a @click.stop="xClick" v-if="xicon && label != \'All\'" href="#" class="text-dark"> x </a> \
             </div>&nbsp;</span>',
  methods: {
    labelClick: function(event) {
      this.$emit('labelclick', this.label)
    },
    xClick: function(event) {
      this.$emit('xclick', this.label)
    },
  }             
})

/*
======================================================================
 pim-tag-picker
----------------------------------------------------------------------
 Inputs: tags     - the list of tags to choose from
         selected - the list of selected tags (subset of tags)
         tiles    - use tiles to show the tags
         menu     - use a menu to show the tags (being phased out)
         task     - optionally provide a task so a tag can be removed
         xicon    - show each tag with a "x" so it can be "removed"

 Events: tagclick - let parent know tile was clicked (it may filter)
         xclick   - let parent know a tile's x was clicked
         
 Display a list of tags, usually as tiles.  This is used in two
 contexts today, and uses events to allow the parent to handle the 
 behavior of either filtering lists for tags, or removing tags from
 individual tasks.

 Note that "menu mode" only supports one tag selected at a time while
 "tile mode" allow multiple tags to be selected (or de-selected)   

 Also note we have special handling for the "All" tag which makes
 itself active if no other tags are selected.
====================================================================*/
Vue.component('pim-tagpicker', {
  props: ['tags', 'selected', 'tiles', 'menu', 'xicon'],
  template: '<div> \
              <pim-tag v-for="tag in tags" v-bind:key="tag" \
                       :label="tag" \
                       :active="isActive(tag)" \
                       :xicon="xicon" \
                       @labelclick="labelclick(tag)" \
                       @xclick="xclick(tag)" \
                       /> \
              <select v-if="menu" class="form-control" v-on:change="filter"> \
                <option v-for="tag in tags">{{ tag }}</option> \
              </select> \
             </div>',
  methods: {
    labelclick: function(tag) {
      this.$emit('tagclick', tag)
    },
    xclick: function(tag) {
      this.$emit('tagremove', tag)
    },
    isActive: function(tag) {
      if (this.selected && this.selected.length > 0) {
        return (this.selected.find(t => tag == t) != undefined);
      }
      else {
        return (tag === "All")
      }
    }
  }             
})

/*
======================================================================
 pim-task
----------------------------------------------------------------------
 Inputs: task  - my task to display
         week  - show today/reset-today based on state of task

 Events: drag   - emitted when i'm being dragged
         drop   - emitted when another task has been dropped on me
         editme - emitted when the user has requested that I be edited
         
 This displays a task in its standard form, including a checkbox to
 change it's state, and "play/pause" to put it in progress, and an
 optional week/day switch, which I hope to generalize to any set of
 levels annual/quarter/month/week/day planning views.
====================================================================*/
Vue.component('pim-task', {
  props: ['task', 'week'],
  methods: {
    toggle: function() {
      // tell the outside world so they can persist the change
      changeTaskState(this.task)
    },
    starttime: function() {
      return this.task.startTimeString()
    },
    editme: function(task) {
      this.$emit('editme', task)
    },
    drag: function(ev, task) {
      ev.dataTransfer.dropEffect = 'move'
      ev.dataTransfer.effectAllowed = 'move'
      ev.dataTransfer.setData('id', task.id)
      this.$emit('drag', ev)
    },
    drop: function(ev) {
      this.$emit('drop', ev)
    }
  },
  template:'<div :id="task.id" class="d-flex justify-content-between p-1" draggable @dragstart="drag($event, task)" \
                   @drop="drop($event)" @dragover.prevent @dragenter.prevent > \
              <div :id="task.id" class="p-0"> \
                  <input type="checkbox" :id="task.id" v-model="task.state" id="inner" \
                       v-bind:true-value=1 v-bind:false-value=0 \
                       @change="toggle"> \
                  <span class="text-dark small" v-if="task.hasStartTime()"><strong>{{starttime()}}</strong></span> \
                  <pim-task-name :task="task" @clicked="editme"></pim-task-name> \
              </div> \
              <div class="p-0 d-flex justify-content-end align-items-baseline"> \
                <pim-duration :duration="task.estimateString()" /> \
                <pim-startstop :task="task" class="pl-1" /> \
                <pim-task-settoday :task="task" v-if="this.week && this.task.isThisWeekAndNotToday()" class="pl-1"></pim-task-settoday> \
                <pim-task-resettoday :task="task" v-if="this.week && this.task.isThisWeekAndToday()" class="pl-1"></pim-task-resettoday> \
              </div> \
            </div>'
})

/*
======================================================================
 pim-task-list
----------------------------------------------------------------------
 Inputs: taskList      - tasks from which I should choose a subset
         title         - title of my list
         hidewhenempty - when true, hide my container when no tasks
         clear         - include icon to "clear" a tag from this list
         add           - include icon to add a new task to this list
         week          - only show tasks that have "thisweek" tag
         day           - only show tasks that have "today" tag
         state         - only show tasks that match the provded state
         time          - only show tasks that have a start time
         tags          - only show tasks that match ALL these tags
         systag        - only show tasks that have this systag

 Events: drag   - emitted when one of our tasks if being dragged
         drop   - emitted when one of out asks has been dropped upon
         modal  - emitted with appropriate context if newtask or
                  edittasks events were received form our children
         
 This baby is at the heart of the app, and flexibly selects a subset
 of tasks from the provided taskList property to display to the user
 based on other properties that choose the tags or task-attributes
 this particular list cares about showing.  Most of this magic
 happens using the matchingtasks computed property, which notices
 any time a task changes within the list, and adds/removes/reorders
 appropriately based on the week, day, state, time and tags props.
====================================================================*/
Vue.component('pim-task-list', {
  props: {
    taskList: { // all tasks on the page - not just mine
      default: null,
      type: Object
    },
    title: {
      default: "",
      type: String
    },
    hidewhenempty: {
      default: false,
      type: Boolean      
    },
    clear: {
      default: null,
      type: String
    },
    add: {
      default: false,
      type: Boolean
    },
    week: {
      default: false,
      type: Boolean
    },
    day: {
      default: false,
      type: Boolean
    },
    state: {  // only show tasks in this state if defined
      default: undefined,
      type: Number
    },
    time: { // only or never show tasks with a time if defined
      default: undefined,
      type: Boolean
    },
    tags: { // only show tasks that match these tags
      default: null,
      type: Array
    },
    systag: { // we're allowed one system tag per list
      default: null,
      type: String
    }
  },
  methods: {
    // one of my tasks has asked to be edited so ask my parent 
    // to show the modal with my list and the task to edit
    edittask: function(task) { 
      this.$emit('modal', {task: task, list: this.taskList, tags: null});
    },
    // my newtask button has been clicked
    newtask: function() {
      this.$emit('modal', {task: null, list: this.taskList, tags: [this.systag]});
    },
    // i've been requested to clear a tag from all entries visible
    // usually to clear the systag like "today" or "this week"
    cleartag: function() {
      clearTagFromList(this.matchingtasks, this.clear, true);
    },
    drag: function(ev) {
      this.$emit('drag', ev, this.title)
    },
    drop: function(ev) {
      // if dropped within the same list for reordering
      // we should handle it here and only emit the
      // even if we are dragging across lists.
      var dragged_id = ev.dataTransfer.getData("id")
      var dragged = this.matchingtasks.findTask(dragged_id);
      if (dragged != null) {
        // get the id of the item dropped onto
        var on_id = ev.target.id;
        var on = this.matchingtasks.findTask(on_id)

        // reorder here somehow within my own list
        // now that I have all the info I need
        // we need to rework the keys for these lists
        // so vue will reorder them for us.  read up
        // on this in v-for documentation.
        console.log('in the same list - we will reorder')
      }
      else {
        this.$emit('drop', ev, this.title)
      }
    },
  },
  computed: {
    // this magic computed property returns the subset of tasks provided
    // that match the critera (state, time, tags) for what to show in this
    // instance of the task list.  then, as tasks from the original
    // list change they are automatically added or removed from this component
    matchingtasks: function() {
      var filtered = this.taskList.filter(function(t) {
        return ((this.state === undefined || t.state == this.state) && 
                (this.time === undefined || t.hasStartTime() == this.time) &&
                (this.tags == null || t.hasAllTags(this.tags)) &&
                (this.systag == null || t.isTagSet(this.systag)))
      }, this)

      // sorting
      if (this.time) {  // sort by start time
        return filtered.sort((a,b) => (a.getTargetStartTime() > b.getTargetStartTime())?1:-1)
      }
      else if (this.state === TaskState.COMPLETE) { // sort by completion time
        return filtered.sort((a,b) => (a.getActualCompletionTime() > b.getActualCompletionTime())?1:-1)
      }
      return filtered // sort in natural order
    }
  },
  template: '<div v-if="this.matchingtasks.numTasks() || !hidewhenempty" class="card mt-2"> \
              <div class="card-header text-white bg-primary d-flex justify-content-between align-items-baseline pt-1 pb-0 pl-2 pr-2"> \
                <h6>{{title}} <pim-add v-if="this.add" @newtask="newtask"/></h6>\
                <div class=""> \
                  <pim-duration :duration="this.matchingtasks.durationFormatted()" /> \
                  <pim-clear v-if="this.clear" @clear="cleartag" /> \
                </div> \
              </div> \
              <div class="card-body list-group list-group-flush p-1"> \
                <pim-task v-for="task in matchingtasks.tasks" \
                              :key="task.id" :task="task" :week="week" :day="day" \
                              @drag="drag" @drop="drop" \
                              @editme="edittask" \
                              class="list-group-item" /> \
              </div> \
             </div>'
})


/*
======================================================================
 pim-title-bar
----------------------------------------------------------------------
 Inputs: title    - the major heading to show in the title bar
         add      - whether or not to show an "add task" icon
         tags     - list of tags to show in the bar for selection
         selected - list of tags currently selected

 Events: modal   - emitted if add button was clicked

         
 Allow the user to create or edit a task.  This is invoked from my
 parent by simply setting / changing the task property.  Note that
 my parent needs to clear / reset the property in order for me to
 notice that a new task has been set on me.
====================================================================*/
Vue.component('pim-title-bar', {
  props: ['tags', 'selected', 'title', 'add'],
  methods: {
    newtask: function() { // our add-a-task button was clicked
      this.$emit('modal', null);      
    },
    tagClick: function(tag) {
      toggleTag(tag)      
    }
  },
  template: '<div class="col-12 card bg-primary text-white"> \
                 <div class="card-header lead pt-0 pb-0 pl-0 pr-0 d-flex justify-content-between"> \
                   <div class="pl-0 m-0"> \
                     <strong>{{title}}</strong> \
                     <pim-add v-if="this.add" @newtask="newtask"/>\
                   </div> \
                   <div class="small"> \
                     <pim-tagpicker :tags="tags" :selected="selected" :tiles=true :menu=false @tagclick="tagClick" /> \
                   </div> \
                 </div> \
             </div>'
})

/*
======================================================================
 pim-modal
----------------------------------------------------------------------
 Inputs: task    - task to edit, or null if creating a new task
         list    - list into which task should be added into client
         systags - list of current context system tags.  today these
                   are emitted with save event to auto-create these
                   tags on new tasks, but may later be used to filter
                   system tags from user views.
 Events: dismiss - emitted when canceled - no args
         save    - edmitted when saved with {task, list, systags}
         
 Allow the user to create or edit a task.  This is invoked from my
 parent by simply setting / changing the task property.  Note that
 my parent needs to clear / reset the property in order for me to
 notice that a new task has been set on me.
====================================================================*/
Vue.component('pim-modal', {
  props: ['task', 'list', 'page', 'systags'],
  data: function() {
    return {
      t: new Task(),
      strtime: "",
      strdate: "",
      strtags: "",
    }
  },
  methods: {
    cancel: function() { // reset the box to where it was on load
      // left as separate method, but for now cancel and load do
      // the same thing because we want to "undo" any changes
      // made while we were editing.  Needed in case a task is
      // edited in the box, canceled, then clicked again
      this.load() 

      // tell my parent so it can clear the box for next time
      this.$emit('dismiss')
    },
    save: function() { 
      // bring the tags and date form fields together
      this.t.addTagsFromString(this.strtags) // tags combined from t and text box
      this.t.setTargetStart(this.strdate, this.strtime)      

      // tell my parent so it can persist and update actual tasks in the lists
      this.$emit('save', {task: this.t, list: this.list, systags: this.systags})

      // clear the box manually since it won't trigger on load for some reason
      this.t = new Task()
      this.strtime = ""
      this.strdate = ""
      this.strtags = "" // always clear the "add tags" box
    },
    deltask: function() { // remove the task from the list and call the server
      this.list.removeTask(this.task) // we'll take care of the JS objects
      deleteTask(this.task); // call a helper to persist the change      
    },
    load: function() { // prepare box for use (call to initialize task)
      if (this.task) {
        this.t.copy(this.task)
        this.strtime = this.t.justStartTime()
        this.strdate = this.t.justStartDate()
      }
      else {
        this.t = new Task()
        this.strtime = ""
        this.strdate = ""
      }
      this.strtags = "" // always clear the "add tags" box

      // make myself visible anytime i'm loaded
      $('#myModal').modal('show')      
    },
    addtag: function() { // adds to our local copy - save will write to server
      this.t.addTagsFromString(this.strtags)
      this.strtags = ""
    },
    removetag: function(tag) { // remove from local copy - save will persist
      console.log("modal: removetag "+ tag)
      this.t.removeTag(tag)
    }
  },
  watch: {
    // anytime the task prop is set (when invoking modal) load it up
    // note modal "holds" last task until another is clicked
    task: function() {
      this.load();
    },
  },
  computed: {
    title: function() { // set the modal title based on whether a task was provided
      if (this.task) {
        return "Edit Task" //  + this.task.getName();
      } else {
        return "Create New Task"
      }
    }
  },
  template: '<div class="modal fade" id="myModal" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true"> \
              <div class="modal-dialog" role="document"> \
                <div class="modal-content"> \
                  <div class="modal-header"> \
                    <h5 class="modal-title" id="myModalLabel">{{title}}</h5> \
                    <button type="button" class="close" v-on:click="cancel" data-dismiss="modal" aria-label="Close"> \
                      <span aria-hidden="true">&times;</span> \
                    </button> \
                  </div> \
                  <form id="newTask" action="createTask"> \
                    <div class="modal-body"> \
                      <div class="form-group"> \
                        <div class="input-group input-group mb-3"> \
                          <div class="input-group-prepend"> \
                            <label class="input input-group-text" for="task">Task Name:</label> \
                          </div> \
                          <input v-model="t.name" type="text" class="form-control" id="task" aria-describedby="taskHelp" placeholder="Short summary for your task"> \
                        </div> \
                      </div> \
                      <div class="form-group"> \
                        <div class="input-group input-group mb-3"> \
                          <div class="input-group-prepend"> \
                            <label class="input input-group-text" for="startdate">Start Date:</label> \
                          </div> \
                          <input type="date" class="form-control" id="startdate"  placeholder="Optional future start date" v-model="strdate"> \
                        </div> \
                      </div> \
                      <div class="form-group"> \
                        <div class="input-group input-group mb-3"> \
                          <div class="input-group-prepend"> \
                            <label class="input input-group-text" for="starttime">Start Time:</label> \
                          </div> \
                          <input type="time" class="form-control" id="starttime" placeholder="Optional start time" v-model="strtime"> \
                        </div> \
                      </div> \
                      <div class="form-group"> \
                        <div class="input-group input-group mb-3"> \
                          <div class="input-group-prepend"> \
                            <label class="input input-group-text" for="duration">Estimate (min):</label> \
                          </div> \
                          <input type="number" class="form-control" id="duration" placeholder="Optional minutes to complete" v-model.number="t.estimate"> \
                        </div> \
                      </div> \
                      <div><span>&nbsp;</span></div> \
                      <div class="form-group"> \
                        <pim-tagpicker :tags="t.tags" :tiles=true :menu=false :xicon=true @tagremove="removetag"/> \
                        <div class="input-group input-group-sm mb-3"> \
                          <div class="input-group-prepend"> \
                            <label class="input input-group-text" for="tags">Tags:</label> \
                          </div> \
                          <input type="text" class="form-control" placeholder="Add one at a time or separate by /s" v-model="strtags" id="tags"> \
                          <div class="input-group-append"> \
                            <button class="btn btn-primary" type="button" v-on:click="addtag">Add</button> \
                          </div> \
                        </div> \
                      </div> \
                      <input type="hidden" id="today" value="true"> \
                      <input type="hidden" id="thisweek" value="false"> \
                    </div> \
                    <div class="modal-footer"> \
                      <button type="button" class="btn btn-secondary" data-dismiss="modal" v-on:click="deltask" id="delete">Delete Task</button> \
                      <!-- button type="button" class="btn btn-secondary" data-dismiss="modal" id="cancel">Cancel</button --> \
                      <button type="submit" class="btn btn-primary" data-dismiss="modal" v-on:click="save" id="save">Save Task</button> \
                    </div> \
                  </form> \
                </div> \
              </div> \
            </div>'
})

/*
======================================================================
 pim-navbar
----------------------------------------------------------------------
 Inputs: items    - list with nav items, each item = {display, route}
         selected - display name of the item that should be active
         
 Display the nav bar which today just navigates to all the new pages
 and does nothing on login.  TBD: emit events so the parent can
 actually navigate and so other things like login / logout.
====================================================================*/
Vue.component('pim-navitem', {
  props: ['name', 'target', 'selected'],
  template: '<li v-bind:class="{ \'nav-item\': !selected, \'nav-item active\': selected }" \> \
               <a class="nav-link" :href="target">{{name}}</a> \
             </li>'
})

Vue.component('pim-navbar', {
  props: ['items', 'selected'],
  template: ' \
              <nav class="navbar navbar-default navbar-expand-sm navbar-light bg-light rounded"> \
                <a class="navbar-brand" href="#">PIM</a> \
                <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarsCollapsible" aria-controls="navbarsCollapsible" aria-expanded="false" aria-label="Toggle navigation"> \
                  <span class="navbar-toggler-icon"></span> \
                </button> \
                <div class="collapse navbar-collapse" id="navbarsCollapsible"> \
                  <ul class="navbar-nav mr-auto"> \
                    <pim-navitem v-for="item in items" :key="item.display" :name="item.display" :target="item.route" :selected="item.display==selected" /> \
                  </ul> \
                  <ul class="nav navbar-nav ml-auto"> \
                    <li class="nav-item"> \
                      <a class="nav-link" href="#"><span class="fa fa-user"></span> Sign Up</a> \
                    </li> \
                    <li class="nav-item"> \
                      <a class="nav-link" href="#"><span class="fa fa-sign-in"></span> Login</a> \
                    </li> \
                  </ul> \
                </div> \
              </nav>'
})

/*
======================================================================
 pim-alert
----------------------------------------------------------------------
 Inputs: message - string to display in the warning box
         show    - visibility flag: set to true to display the box
 Events: dismiss - emitted when dismissed so parent can track

 This component displays an error message to the user.  It is invoked
 from it's parent by setting the message and setting the show prop
 to true.  Note that it only triggers when the show prop actually
 changes value, so the parent should keep the show property in sync
 when the box is dismissed.
====================================================================*/
Vue.component('pim-alert', {
  props: ['message', 'show'],
  data: function () {
    return {
      visible: false, // start hidden
    }
  },  
  methods: {
    dismiss: function() {
      this.visible = false // hide myself and tell my parent i'm gone
      this.$emit('dismiss')
    }
  },
  watch: {
    show: function() {
      // note this can cause visibility and my parent to
      // get out of sync - but i wanted the box to be able
      // to dismiss itself without the parent's help
      // parent SHOULD update whatever state is tracking
      // this component's visibility on dismiss event
      this.visible = this.show
    }
  },
  template: ' \
    <div v-show="visible" class="alert-float" style="position: absolute; top: 0; left: 10px; right: 10px; z-index: 9999; width: 100%" id="alert"> \
      <div class="alert alert-danger alert-dismissible show" role="alert" id="alerttext"> \
        {{ message }} \
        <button type="button" class="close" aria-label="Close" @click="dismiss"> \
          <span aria-hidden="true">&times;</span> \
        </button> \
      </div> \
    </div> \
    '
})


/*
======================================================================
 pim-selector-date
----------------------------------------------------------------------
 Inputs: availableDates - array of dates that even have tasks
         value          - selected date

 Events: input - emitted any time the user chooses a new date
 Requires: v-calendar component library
         
 Vue wrapper for my calendar control that wrappers the v-date-picker 
 component.
====================================================================*/
Vue.component('pim-selector-date', {
  props: {
    availableDates: Array,
    value: Date,
  },
  data: function() {
    return {
      timestamp: null,
    }
  },
  computed: {
    date: {
      get: function() {
        return this.value;
      },
      set: function(newDate) {
        this.timestamp = newDate;
      },
    },
  },
  created: function() {
    this.timestamp = this.value;
  },
  methods: {
    onInput: function(newDate) {
      this.$emit('input', newDate)
    },
  },
  template: '<div class="card mt-2"> \
               <div class="card-header text-white bg-primary d-flex justify-content-between align-items-baseline pt-1 pb-0 pl-2 pr-2"> \
                 <h6 class="">Date</h6> \
               </div> \
               <v-date-picker mode="single" color="blue" v-on:input="onInput($event)" is-inline v-model="date" \
                :available-dates="availableDates" is-required /> \
            </div>'
})

