package main

import (
    "fmt"    
    "errors"
    "log"
)

/*
==============================================================================
 Commands and Undo / Redo
------------------------------------------------------------------------------
 The Command and commandHistory objects allow us to define and track what
 users are doing, track all actions in a stack, and allow for undo and redo
 functions.  The intended use is to log a command onto the stack for every
 user action that modifies something, and that each command includes enough
 information to execute (do/redo) or undo the command.

 Each command implements the Command interface to define the Exec() and
 Undo() actions for that command.  Today that are below in this file, but
 it may make sense to separate the "generic" Command and commandHistory
 objects from the PIM-specific commands.

 We don't actually expose the command stack except through the two functions
 which execute or undo, and take care of all the pushing and popping to the
 stack:
    CommandDo(cmd Command) - executes the command and pushes it to history
    CommandUndo()          - pops the most recent command and undoes it

 TBD: create a redo stack and a CommandRedo() function.

 TBD: Of course, all of this depends on the creation of new Commands that 
 exec and undo their work, and making sure these commands are actually used 
 to do all the work that changes things. 
============================================================================*/
type Command interface {
    Exec() error
    Undo() error
    Log() string
}

type commandHistory struct {
    cmds []Command
    redo []Command // TBD - not yet used
}

func (h *commandHistory) IsEmpty() bool {
    return len(h.cmds) == 0
}

func (h *commandHistory) Push(c Command) error {
    h.cmds = append(h.cmds, c)
    return nil
}

func (h *commandHistory) Pop() Command {
    if h.IsEmpty() {
        return nil
    } else {
        index := len(h.cmds) - 1 // Get the index of the top most element.
        c := h.cmds[index] // Index into the slice and obtain the element.
        h.cmds = h.cmds[:index] // Remove it from the stack by slicing it off.
        return c

    }
}

// since the pattern we chose is to keep the actual command objects
// internal, this function should only be called from a command object.
func CommandDo(cmd Command) error {
    err := cmd.Exec()
    if err == nil {
        if commands == nil {
            commands = new(commandHistory)
        }
        commands.Push(cmd)
    }
    log.Print(cmd.Log())
    return err
}

func CommandUndo() error {
    if commands == nil || commands.IsEmpty() {
        log.Print(commandLog("UNDO-ERROR", "Nothing to Undo"))
        return errors.New("nothing to undo")
    } else {
        cmd := commands.Pop()
        err := cmd.Undo()
        return err
    }
}

func commandLog(cmd string, context string) string {
    return fmt.Sprintf("CMD-%s-%s\n", cmd, context)    
}

/*
==============================================================================
 DeleteTaskCmd
------------------------------------------------------------------------------
 This command deletes a task and allows for undo.  It keeps an in-memory
 pointer to the task that was deleted so it can be undone.
============================================================================*/
type deleteTaskCmd struct {
    tDelete    *Task
    tNewParent *Task // this can be nil and usually is for now
    tOldParent *Task // save so we know where to 
    sLog       string
    // TBD: save multiple parents
    // TBD: save previous children so we can restore them
}

func (dtc *deleteTaskCmd) Exec() error {
    // save the parent (assumes one parent and no children)
    dtc.tOldParent = dtc.tDelete.FirstParent()

    // TBD: save additional parents and children

    // delete the task which will immediately delete in storage
    dtc.sLog = commandLog("EXEC-DELETE", dtc.tDelete.GetName())
    dtc.tDelete.Remove(dtc.tNewParent)

    return nil
}

func (dtc *deleteTaskCmd) Undo() error {
    // reestablish parent (assumes one parent and no children)
    dtc.tOldParent.AddChild(dtc.tDelete)

    // TBD re-establish multiple parents and children

    // save the task
    dtc.tDelete.Save(true)
    dtc.sLog = commandLog("UNDO-DELETE", dtc.tDelete.GetName())

    return nil
}

func (dtc *deleteTaskCmd) Log() string {
    return dtc.sLog
}

func CommandDeleteTask(t *Task, tNewParent *Task) error {
    var cmdDelete *deleteTaskCmd
    cmdDelete = new(deleteTaskCmd)
    cmdDelete.tDelete = t
    cmdDelete.tNewParent = nil
    CommandDo(cmdDelete)
    return nil
}


/*
==============================================================================
 CreateTaskCmd
------------------------------------------------------------------------------
 This command creates a task and allows for undo.  It keeps an in-memory
 pointer to the task that was created so it can be undone by deleting it.
============================================================================*/
type createTaskCmd struct {
    tCreate    *Task
    sLog       string
    // TBD: save multiple parents
    // TBD: save previous children so we can restore them
}

func (ctc *createTaskCmd) Exec() error {
    // save the parent (assumes one parent and no children)
    // dtc.tOldParent = dtc.tDelete.FirstParent()

    // TBD: save additional parents and children

    // delete the task which will immediately delete in storage
    // fmt.Printf("createTaskCmd.Exec(): creating %s\n", ctc.tCreate.GetName())
    ctc.sLog = commandLog("EXEC-CREATE", ctc.tCreate.GetName())         
    err := ctc.tCreate.Save(true)   

    return err
}

func (ctc *createTaskCmd) Undo() error {
    // TBD - consider how we would redo multiple parents and children
    // delete the task
    // fmt.Printf("createTaskCmd.Undo(): undoing create of %s\n", ctc.tCreate.GetName())    
    ctc.tCreate.Remove(nil) // nil -> orphan any of my children ???
    ctc.sLog = commandLog("UNDO-CREATE", ctc.tCreate.GetName())    
    return nil
}

func (ctc *createTaskCmd) Log() string {
    return ctc.sLog
}

func CommandCreateTask(t *Task) error {
    var cmdCreate *createTaskCmd
    cmdCreate = new(createTaskCmd)
    cmdCreate.tCreate = t
    return CommandDo(cmdCreate)
}


/*
==============================================================================
 UpdateTaskCmd
------------------------------------------------------------------------------
 This command modifies a task and allows for undo.  It keeps an in-memory
 copy of the task as it looked prior to the operation so it can be undone.
============================================================================*/
type updateTaskCmd struct {
    tPrior    *Task
    tUpdate   *Task
    bPrepared bool
    sLog      string
    // TBD: save multiple parents
    // TBD: save previous children so we can restore them
}

// allow updates to be broken into two parts in case the client of
// this code is modifying a task "in place".  ExecPrepare(), then, saves
// the task before it is motified, and Exec() actually persists the
// changed version of the task
func (utc *updateTaskCmd) ExecPrepare() error {
    // make a copy of the old task and keep the original to update
    utc.tPrior = utc.tPrior.Copy(nil)
    utc.bPrepared = true
    return nil // TBD: errors from copy
}

func (utc *updateTaskCmd) Exec() error {

    // by the time Exec is being called we should have two copies
    // of tasks: what the task looks like before and should look like
    // after modification
    if utc.tPrior == nil || utc.tUpdate == nil {
        return errors.New("could not update - insufficient data provided")
    }

    // save the parent (assumes one parent and no children)
    // dtc.tOldParent = dtc.tDelete.FirstParent()

    // TBD: save additional parents and children

    // modify the task which will immediately change in storage
    // all we need to do is save the modified task
    // fmt.Printf("updateTaskCmd.Exec(): changing %s\n", utc.tPrior.GetName())
    utc.sLog = commandLog("EXEC-UPDATE", utc.tUpdate.GetName())             
    err := utc.tUpdate.Save(false)    
    return err
}

func (utc *updateTaskCmd) Undo() error {
    // copy the object values back and resave
    utc.sLog = commandLog("UNDO-UPDATE", utc.tUpdate.GetName())
    utc.tPrior.Copy(utc.tUpdate)
    err := utc.tUpdate.Save(true)
    return err
}

func (utc *updateTaskCmd) Log() string {
    return utc.sLog
}


/*
func CommandModifyTask(t *Task) error { // not there yet with handler - need 2 tasks to be passed in?
    var cmdUpdate *updateTaskCmd
    cmdUpdate = new(updateTaskCmd)
    cmdUpdate.tPrior = t
    return CommandDo(cmdUpdate)
}
*/

func CommandModifyTaskBegin(t *Task) *updateTaskCmd { 
    var cmdUpdate *updateTaskCmd
    cmdUpdate = new(updateTaskCmd)
    cmdUpdate.tPrior = t
    cmdUpdate.ExecPrepare()
    return cmdUpdate
}

func CommandModifyTaskEnd(cmdUpdate *updateTaskCmd, t *Task) error { 
    cmdUpdate.tUpdate = t
    return CommandDo(cmdUpdate)
}
