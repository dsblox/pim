/*
==============================================================================
 PIM Vue Components
------------------------------------------------------------------------------
 This is our component library for the application.  It is built 
 hierarchically with the main components being:
  pim-task-list - binds to a TaskList
  pim-task      - beings to a Task

 Lots of other sub-components represent user controls displaying attributes
 and state of the tasks, and are grouped into the previous two for display.
============================================================================*/

Vue.component('pim-start-time', {
  props: ['task'],
  template: '<span class="text-dark small" v-if="task.hasStartTime()"><strong>{{task.startTimeString()}}</strong></span>'
})

Vue.component('pim-task-name', {
	props: ['name', 'id'],
	template: '<a href="#myModal" \
				        data-toggle="modal" \
                class="task-name" \
                :id="id" \
	              :data-taskid="id">{{name}}</a>',
})

Vue.component('pim-duration', {
	props: ['duration'],
	template: '<span v-if="duration" class="badge badge-secondary ml-1">{{duration}}</span>' 
})

Vue.component('pim-task-settoday', {
  props: ['task'],
  methods: {
    settoday: function() {
      setToday(this.task);
    }
  },
  template: '<span class="fa fa-chevron-right" v-on:click="settoday"></span>'
})

Vue.component('pim-task-resettoday', {
  props: ['task'],
  methods: {
    resettoday: function() {
      resetToday(this.task);
    }
  },
  template: '<i class="fa fa-chevron-left" v-on:click="resettoday"></i>'
})

Vue.component('pim-startstop', {
  props: ['task'],
  methods: {
    startstop: function() {
      startStop(this.task);
    }
  },
  template: '<span v-if="this.task.isInProgress()" class="fa fa-pause" v-on:click="startstop"></span> \
             <span v-else-if="!this.task.isComplete()" class="fa fa-bolt" v-on:click="startstop"></span>'
})


Vue.component('pim-task', {
  props: ['task', 'week', 'day'],
  methods: {
    toggle: function() {
      this.$emit('toggle', this.task);
    }
  },
  template:'<div :id="task.id" class="d-flex justify-content-between p-1" draggable="true" ondragstart="drag(event)"  \
                 ondrop="drop(event)" ondragover="allowDrop(event)" > \
              <div :id="task.id" class="p-0"> \
                  <input type="checkbox" :id="task.id" v-model="task.state" id="inner" \
                  	   v-bind:true-value=1 v-bind:false-value=0 \
                       v-on:input="toggle"> \
                	<pim-start-time :task="task"></pim-start-time> \
                	<pim-task-name :name="task.getName()" :id="task.id"></pim-task-name> \
              </div> \
              <div class="p-0 d-flex justify-content-end align-items-baseline"> \
                <pim-duration :duration="task.estimateString()" /> \
                <pim-startstop :task="task" class="pl-1" /> \
                <pim-task-settoday :task="task" v-if="this.week" class="pl-1"></pim-task-settoday> \
                <pim-task-resettoday :task="task" v-if="this.day && this.task.isThisWeek()" class="pl-1"></pim-task-resettoday> \
              </div> \
            </div>'
})


Vue.component('pim-add', {
	props: ['id'],
	template: '<a href="#myModal" class="text-white" \
	              data-toggle="modal" \
	              :data-list="id"> + </a>'
})

Vue.component('pim-clear', {
  props: ['taskList', 'tagtoclear'],
  template: '<span class="fa fa-archive" v-on:click="clear"></span>',
  methods: {
    clear: function(event) {
      clearTagAndList(this._props.taskList, this._props.tagtoclear);
    }
  }
})

// TBD: note that "menu mode" only supports one tag selected at a time while
// "tile mode" allow multiple tags to be selected (or de-selected)
Vue.component('pim-tagpicker', {
  props: ['tags', 'selected', 'tiles', 'menu'],
  template: '<div> \
              <span v-for="tag in tags"> \
                <a href="#" v-if="tiles" \
                   class="badge" \
                   v-bind:class="{ \'badge-dark\': imselected(tag), \'badge-light\': !imselected(tag) }" \
                   v-on:click="filter">{{ tag }} \
                </a>&nbsp; \
              </span> \
              <select v-if="menu" class="form-control" v-on:change="filter"> \
                <option v-for="tag in tags">{{ tag }}</option> \
              </select> \
             </div>',
  methods: {
    imselected: function(mytext) {
      if (this.selected.length == 0 && mytext == "All") {
        return true;
      }
      else {
        return (this.selected.indexOf(mytext) != -1);
      }
    },
    filter: function(event) {
      if (event.type == "click") { // tiles were used
        toggleTag(event.target.innerText);
      }
      else { // select dropdown was used type is "change"
        selectTag(event.target.value, true);
      }
    }
  }
})

// TBD: make the task list data-bound to the isToday() field on the
// tasks so they can automatically clear out when the clear link is
// clicked.

Vue.component('pim-task-list', {
  props: ['taskList', 'title', 'add', 'clear', 'week', 'day', 'hidewhenempty'],
  data: function () {
    return {
      numTasks: this.taskList.tasks.length,
      id: this.taskList.id,
    }
  },
  methods: {
    toggle: function(task) {
      this.$emit('toggle', task);
    },

    newtask: function (event) {
      alert('Hello ' + this.title + '!')
      // `event` is the native DOM event
      alert(event.target.tagName)
    },
  },
  template: '<div v-if="this.taskList.numTasks() || !hidewhenempty" class="card mt-2"> \
              <div class="card-header text-white bg-primary d-flex justify-content-between align-items-baseline pt-1 pb-0 pl-2 pr-2"> \
                <h6>{{title}} <pim-add v-if="this.add" :id="id" /> </h6>\
                <div class=""> \
                  <pim-duration :duration="this.taskList.durationFormatted()" /> \
                  <pim-clear v-if="this.clear" :tagtoclear="this.clear" :taskList="taskList" /> \
                </div> \
              </div> \
              <div class="card-body list-group list-group-flush p-1"> \
                <pim-task v-for="task in taskList.tasks" \
                              :key="task.id" :task="task" :week="week" :day="day" \
                              class="list-group-item" \
                              v-on:toggle="toggle" /> \
              </div> \
             </div>'
})

                 


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


Vue.component('pim-new-button', {
  template: '<a href="#" class="text-white" data-toggle="modal" data-target="#myModal"> + </a>'
})

Vue.component('pim-title-bar', {
  props: ['tags', 'selected', 'title', 'add'],
  template: '<div class="col-12 card bg-primary text-white"> \
                 <div class="card-header lead pt-0 pb-0 pl-0 pr-0 d-flex justify-content-between"> \
                   <div class="pl-0 m-0"> \
                     <strong>{{title}}</strong> \
                     <pim-new-button v-if="this.add" />\
                   </div> \
                   <div class="small"> \
                     <pim-tagpicker :tags="tags" :selected="selected" :tiles=true :menu=false /> \
                   </div> \
                 </div> \
             </div>'
})

// could not get this to work - bootstrap lost its positioning and collapse functionality
// when I tried to cross over with vue components.  Something to fix later.
Vue.component('pim-navitem', {
  props: ['name', 'target', 'selected'],


  template: '<li v-bind:class="{ \'nav-item\': !selected, \'nav-item active\': selected }" \> \
               <a class="nav-link" :href="target">{{name}}</a> \
             </li>'
})

Vue.component('pim-navbar', {
  props: ['items', 'selected'],
  template: '<div class="navbar navbar-default navbar-expand-sm navbar-light bg-light rounded"> \
              <a class="navbar-brand" href="#">PIM</a> \
              <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarsCollapsible" aria-controls="navbarsCollapsible" aria-expanded="false" aria-label="Toggle navigation"> \
                <span class="navbar-toggler-icon"></span> \
              </button> \
              \
              <div class="collapse navbar-collapse d-flex justify-content-between" id="navbarsCollapsible"> \
                <ul class="navbar-nav"> \
                  <pim-navitem v-for="item in items" :name="item.display" :target="item.route" :selected="item.display==selected" /> \
                </ul> \
              <div class="nav navbar-nav nav-center"> \
                <div class="alert-float" style="position: absolute; top: 0; left: 10px; right: 10px; z-index: 9999; width: 100%" id="err"> \
                  <div class="alert alert-danger alert-dismissible show" role="alert" id=errortext> \
                    OK \
                  <button type="button" class="close" aria-label="Close" onclick="$(\'#err\').hide()"> \
                    <span aria-hidden="true">&times;</span> \
                  </button> \
                </div> \
                <ul class="navbar-nav ml-auto"> \
                  <li class="nav-item"> \
                    <a class="nav-link" href="#"><span class="fa fa-user"></span> Sign Up</a> \
                  </li> \
                  <li class="nav-item"> \
                    <a class="nav-link" href="#"><span class="fa fa-sign-in"></span> Login</a> \
                  </li> \
                </ul> \
              </div> \
             </div>'
})

