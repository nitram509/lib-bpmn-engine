## Supported BPMN elements

These BPMN elements are supported by the latest release of lib-bpmn-engine.

* Start Event
* End Event
* Service Task
    * Get & Set variables from/to context (of the instance)
    * variable mapping
      * input mapping(local scope)
      * output mapping(instance scope)
* User Task
    * equally handled like service tasks
    * variable mapping
      * input mapping(local scope)
      * output mapping(instance scope)
* Forks
    * controlled and uncontrolled forks are supported
    * parallel gateway supported
    * exclusive gateway with conditions
* Joins
    * uncontrolled and exclusive joins are supported
    * parallel joins are supported
* Message Intermediate Catch Event
    * at the moment, just matching/correlation by name supported
    * TODO: introduce correlation key
* Timer Intermediate Catch Event
