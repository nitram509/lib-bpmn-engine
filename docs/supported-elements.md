
These BPMN elements are supported by the latest release of lib-bpmn-engine.
There are some comments as well, which describe the level of support per each element.

## Start Event                      

![](images/start_event.png){: .width-60pt }                      

* multiple start events are supported and are triggered in order of appearance. 

## End Event                        
![](images/end_event.png){: .width-60pt }  


## Service Task                     
![](images/service_task.png){: .width-60pt }         

* you can define task handlers for one or multiple tasks (by ID or by Type)
* get & set variables from/to context (of the instance) is possible
* variable mapping is supported (for input and output, see [Variables](#variables))

## User Task                        
![](images/user_task.png){: .width-60pt }      

* you can define task handlers for one or multiple tasks (by assignee or by candidate groups)
* equally handled like service tasks
* get & set variables from/to context (of the instance) is possible
* variable mapping is supported (for input and output, see [Variables](#variables))

## Error Boundary Event
![](images/error_boundary_event.png){: .width-60pt }

* business errors raised by a task which requires modeling.
* multiple end events are supported as well.

## Sub Process
![](images/sub_process.png){: .width-60pt }    

* sub-processes can encapsulate a series of tasks and events.
* supports both embedded and reusable sub-processes.
* allows for better organization and modularization of complex workflows.
* sub-processes can have their own start and end events.
* supports variable mapping for input and output, similar to tasks.
* can be used to handle repetitive or complex logic within a process.
* 
## Event Sub Process
![](images/event_sub_process.png){: .width-60pt }    

* Only Error Event Sub Process is supported
* event subprocess is a subprocess triggered by an event
* it must have an event based start event

## Gateways

The Parallel Gateway and the Exclusive Gateway do allow fork and join logic,
as well as parallel execution. The library does not support real parallel execution,
but flows are processed in order of appearance in the BPMN file.

* Forks
    * controlled and uncontrolled forks are supported
    * parallel gateway supported
    * exclusive gateway with conditions supported
* Joins
    * uncontrolled and exclusive joins are supported
    * parallel joins are supported

### Exclusive Gateway                
![](images/exclusive_gateway.png){: .width-60pt }                

* fully supported, incl. conditional expressions per each outgoing flow

### Inclusive Gateway                
![](images/inclusive_gateway.png){: .width-60pt }                

* fully supported, incl. conditional expressions per each outgoing flow

### Parallel Gateway                 
![](images/parallel_gateway.png){: .width-60pt }        

* fully supported, incl. conditional expressions per each outgoing flow

### Event Based Gateway              
![](images/event_based_gateway.png){: .width-60pt }              

* fully supported

## Message Intermediate Catch Event 
![](images/message_intermediate_catch_event.png){: .width-60pt } 

* at the moment, just matching/correlation by name supported
* variable mapping is supported (for output, see [Variables](#variables))

## Timer Intermediate Catch Event   
![](images/time_intermediate_catch_event.png){: .width-60pt }     

* supported
* a ticker/scheduler needs to be externally provided, see [Timers](advanced-timers.md)

## Link Intermediate Throw & Catch Event

* supported
* variable mapping is supported(for output, see [Variables](#variables))

![](images/link_intermediate_throw_event.png){: .width-60pt }     
![](images/link_intermediate_catch_event.png){: .width-60pt }         
----

## Variables

### Input Variables

Input mappings can be used to create new variables. They can be defined on service tasks and user task.
When an input mapping is applied, it creates a new local variable in the scope (visible in the task handler)
where the mapping is defined.

### Output Variables

Output mappings can be used to customize how job/message variables are merged into the process instance.
They can be defined on service tasks, user tasks, and message catch events.

Within a task handler, variables are set as local variables (local scope).
When the handler completes the job, the output mappings are applied to the variables
and create new variables in this scope. The new variables are merged into the parent scope (typically the process instance).
If there is no mapping for a job/message variable, the variable is not merged.
In other words: if one or more variables are defined, only these are mapped.

If no output mappings are defined, all job/message variables are merged into the process instance.
