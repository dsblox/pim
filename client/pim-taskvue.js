// create a component with the property "task" which is expected to be an object with
// attributes for whether it is "complete", some "text" and an "id".  The task is
// displayed as a checkbox bound to the "state" field of the task object provided

Vue.component('pim-start-time', {
  props: ['task'],
  template: '<span v-if="task.hasStartTime()"><strong><small>{{task.startTimeString()}}</small></strong></span>'
})

Vue.component('pim-task-name', {
	props: ['name', 'id'],
	template: '<a href="#myModal" \
				        data-toggle="modal" \
                :id="id" \
	              :data-taskid="id">{{name}}</a>',
})

Vue.component('pim-task-estimate', {
	props: ['estimate'],
	template: '<span class="badge">{{estimate}}</span>' 
})

Vue.component('pim-task-settoday', {
  props: ['task'],
  methods: {
    settoday: function() {
      setToday(this.task);
    }
  },
  template: '<span class="glyphicon glyphicon-chevron-right" v-on:click="settoday"></span>'
})

Vue.component('pim-task-resettoday', {
  props: ['task'],
  methods: {
    resettoday: function() {
      resetToday(this.task);
    }
  },
  template: '<span class="glyphicon glyphicon-chevron-left" v-on:click="resettoday"></span>'
})

Vue.component('pim-task', {
  props: ['task', 'week', 'day'],
  methods: {
    toggle: function() {
      this.$emit('toggle', this.task);
    }
  },
  template:'<div :id="task.id" class="checkbox clearfix" draggable="true" ondragstart="drag(event)"  \
                 ondrop="drop(event)" ondragover="allowDrop(event)" > \
              <label :id="task.id" class="col-sm-10"> \
                  <input type="checkbox" :id="task.id" v-model="task.state" id="inner" \
                  	   v-bind:true-value=1 v-bind:false-value=0 \
                       v-on:click="toggle"> \
                	<pim-start-time :task="task"></pim-start-time> \
                	<pim-task-name :name="task.getName()" :id="task.id"></pim-task-name> \
              </label> \
              <div class="pull-right"> \
                <pim-task-estimate :estimate="task.estimateString()"></pim-task-estimate> \
                <pim-task-settoday :task="task" v-if="this.week"></pim-task-settoday> \
                <pim-task-resettoday :task="task" v-if="this.day && this.task.isThisWeek()"></pim-task-resettoday> \
              </div> \
            </div>'
})


Vue.component('pim-add', {
	props: ['id'],
	template: '<a href="#myModal" \
	              data-toggle="modal" \
	              :data-list="id"> + </a>'
})

Vue.component('pim-clear', {
  props: ['taskList'],
  template: '<a href="#" v-on:click="clear"> \
                <span class="glyphicon glyphicon-transfer"></span> \
             </a>',
  methods: {
    clear: function(event) {
      clearTodayAndList(this._props.taskList);
    }
  }
})

// TBD: make the task list data-bound to the isToday() field on the
// tasks so they can automatically clear out when the clear link is
// clicked.

Vue.component('pim-task-list', {
  props: ['taskList', 'title', 'add', 'clear', 'week', 'day'],
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
  template: '<div class="panel panel-primary"> \
               <div class="panel-heading"> \
                 <h4 class="panel-title">{{title}} \
                 <pim-add v-if="this.add" :id="id" /> \
                 <span v-if="this.taskList.duration()" class="badge pull-right" >{{this.taskList.durationFormatted()}}</span> \
                 <pim-clear v-if="this.clear" :taskList="taskList" /> \
                 </h4> \
               </div> \
               <div class="list-group container-fluid"> \
                  <pim-task v-for="task in taskList.tasks" :key="task.id" :task="task" :week="week" :day="day" class=list-group-item v-on:toggle="toggle" /> \
              </div> \
             </div>'
})




// trying to create a general datetime control that I can customize for PIM
// but also be generic for re-use                 

Vue.component('dab-datetime', {
  props: ['onchange', 'initTimestamp'],
  data: function() {
    return {
      // type: "date",
      timestamp: null,
      changeHandler: null
    }
  },
  methods: {
    onChanged: function(event) {
      if (this.changeHandler != null && 
          typeof this.changeHandler === "function") {
        this.changeHandler(this.timestamp); // make this send the new date instead of the event
      }
    }
  },
  created: function () {
    if (this.onchange !== undefined && this.onchange !== undefined) {
      this.changeHandler = this.onchange;
    }
    console.log(this.initTimestamp);
    this.timestamp = this.initTimestamp;
  },
  template: '<div class="form-group"> \
               <input v-on:change="onChanged" v-model="timestamp" type="date" class="form-control"> \
             </div>'
})

Vue.component('pim-selector-date', {
  props: ['onchange', 'initTimestamp'],
  created: function() {
    console.log(this.initTimestamp);
  },
  template: '<div class="panel panel-primary"> \
               <div class="panel-heading"> \
                 <h4 class="panel-title">Date</h4> \
               </div> \
               <dab-datetime :onchange="onchange" :init-timestamp="initTimestamp" /> \
            </div>'
})


