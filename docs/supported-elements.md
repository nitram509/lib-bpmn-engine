
These BPMN elements are supported by the latest release of lib-bpmn-engine.
There are some comments as well, which describe the level of support per each element.

## Start Event                      

![](images/start_event.png){: .width-60pt }                      

* multiple start events are supported and are triggered in order of appearance. 

## End Event                        
![](images/end_event.png){: .width-60pt }  

* multiple end events are supported as well.

## Service Task                     
![](images/service_task.png){: .width-60pt }         

* you can define task handlers for one or multiple tasks (by ID or by Type)
* get & set variables from/to context (of the instance) is possible
* variable mapping is supported, but only one scope exists

## User Task                        
![](images/user_task.png){: .width-60pt }      

* you can define task handlers for one or multiple tasks (by assignee or by candidate groups)
* equally handled like service tasks
* get & set variables from/to context (of the instance) is possible
* variable mapping is supported, but only one scope exists

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

### Parallel Gateway                 
![](images/parallel_gateway.png){: .width-60pt }        

* fully supported, incl. conditional expressions per each outgoing flow

### Event Based Gateway              
![](images/event_based_gateway.png){: .width-60pt }              

* fully supported

## Message Intermediate Catch Event 
![](images/message_intermediate_catch_event.png){: .width-60pt } 

* at the moment, just matching/correlation by name supported

## Timer Intermediate Catch Event   
![](images/time_intermediate_catch_event.png){: .width-60pt }    

* supported
* a ticker/scheduler needs to be externally provided, see [Timers](advanced-timers.md)
