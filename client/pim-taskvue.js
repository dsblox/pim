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
	              data-target="#myModal" \
	              :data-taskid="id">{{name}}</a>',
})

Vue.component('pim-task-estimate', {
	props: ['estimate'],
	template: '<span class="badge">{{estimate}}</span>' 
})

Vue.component('pim-task', {
  props: ['task'],
  methods: {
    toggle: function() {
      this.$emit('toggle', this.task);
    }
  },
  template:'<div :id="task.id" class="checkbox" draggable="true" ondragstart="drag(event)" \
                 ondrop="drop(event)" ondragover="allowDrop(event)"> \
              <label> \
                <input type="checkbox" :id="task.id" v-model="task.state" \
                	   v-bind:true-value=1 v-bind:false-value=0 \
                     v-on:click="toggle"> \
              	<pim-start-time :task="task"></pim-start-time> \
              	<pim-task-name :name="task.getName()" :id="task.id"></pim-task-name> \
              </label> \
              <pim-task-estimate :estimate="task.estimateString()"></pim-task-estimate> \
            </div>'
})

Vue.component('pim-add', {
	props: ['id'],
	template: '<a href="#myModal" \
	              data-toggle="modal" \
	              data-target="#myModal" \
	              :data-list="id"> + </a>'
})


Vue.component('pim-task-list', {
  props: ['taskList', 'title', 'add'],
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
                <span v-if="this.taskList.duration()" class="badge pull-right" >{{this.taskList.durationFormatted()}}</span></h4> \
               </div> \
               <div class="list-group">\
                <template v-for="task in taskList.tasks">\
                 <pim-task :task="task" class=list-group-item v-on:toggle="toggle" /> \
               </template> \
              </div> \
            </div>'
})
