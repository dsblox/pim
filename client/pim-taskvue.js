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

Vue.component('pim-duration', {
	props: ['duration'],
	template: '<span v-if="duration" class="badge">{{duration}}</span>' 
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

Vue.component('pim-startstop', {
  props: ['task'],
  methods: {
    startstop: function() {
      console.log("startstop");      
      startStop(this.task);
    }
  },
  template: '<span v-if="this.task.isInProgress()" class="glyphicon glyphicon-pause" v-on:click="startstop"></span> \
             <span v-else-if="!this.task.isComplete()" class="glyphicon glyphicon-flash" v-on:click="startstop"></span>'
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
                       v-on:input="toggle"> \
                	<pim-start-time :task="task"></pim-start-time> \
                	<pim-task-name :name="task.getName()" :id="task.id"></pim-task-name> \
              </label> \
              <div class="pull-right"> \
                <pim-startstop :task="task" /> \
                <pim-duration :duration="task.estimateString()" /> \
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
  template: '<div v-if="this.taskList.numTasks() || !hidewhenempty" class="panel panel-primary"> \
               <div class="panel-heading"> \
                 <h4 class="panel-title">{{title}} \
                 <pim-add v-if="this.add" :id="id" /> \
                 <div class="pull-right"> \
                   <pim-duration :duration="this.taskList.durationFormatted()" /> \
                   <pim-clear v-if="this.clear" :taskList="taskList" /> \
                 </div> \
                 </h4> \
               </div> \
               <div class="list-group container-fluid"> \
                  <pim-task v-for="task in taskList.tasks" :key="task.id" :task="task" :week="week" :day="day" class=list-group-item v-on:toggle="toggle" /> \
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
  template: '<div class="panel panel-primary"> \
               <div class="panel-heading"> \
                 <h4 class="panel-title">Date</h4> \
               </div> \
               <v-date-picker mode="single" color="blue" v-on:input="onInput($event)" is-inline v-model="date" \
                :available-dates="availableDates" is-required /> \
            </div>'
})




