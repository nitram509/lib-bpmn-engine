package bpmn

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"log"

	"github.com/bwmarrin/snowflake"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20"
	rqlite "github.com/pbinitiative/zenbpm/pkg/bpmn/persistence/rqlite"
	sql "github.com/pbinitiative/zenbpm/pkg/bpmn/persistence/rqlite/sql"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/var_holder"
)

type BpmnEnginePersistenceRqlite struct {
	snowflakeIdGenerator *snowflake.Node
	state                *BpmnEngineState
	rqlitePersistence    *rqlite.BpmnEnginePersistenceRqlite
}

func NewBpmnEnginePersistenceRqlite(snowflakeIdGenerator *snowflake.Node, state *BpmnEngineState, rqlite *rqlite.BpmnEnginePersistenceRqlite) *BpmnEnginePersistenceRqlite {
	gen := snowflakeIdGenerator

	return &BpmnEnginePersistenceRqlite{
		snowflakeIdGenerator: gen,
		state:                state,
		rqlitePersistence:    rqlite,
	}
}

// READ

func (persistence *BpmnEnginePersistenceRqlite) FindProcessById(processId string) *ProcessInfo {
	// TODO finds the latest version
	processes := persistence.FindProcessesById(processId)
	if len(processes) == 0 {
		return nil
	}
	return processes[0]
}

func (persistence *BpmnEnginePersistenceRqlite) FindProcessesById(processId string) []*ProcessInfo {
	return persistence.FindProcesses(processId, -1)
}

func (persistence *BpmnEnginePersistenceRqlite) FindProcessByKey(processKey int64) *ProcessInfo {
	processes := persistence.FindProcesses("", processKey)
	if len(processes) == 0 {
		return nil
	}
	return processes[0]
}

func (persistence *BpmnEnginePersistenceRqlite) FindProcesses(processId string, processKey int64) []*ProcessInfo {
	// TODO finds all processes with given ID sorted by version number

	processes := persistence.rqlitePersistence.FindProcesses(processId, processKey)

	resultProcesses := make([]*ProcessInfo, 0)
	for _, process := range processes {
		// map to ProcessInfo
		resultProcess := &ProcessInfo{
			ProcessKey:       process.Key,
			Version:          process.Version,
			BpmnProcessId:    process.BpmnProcessId,
			bpmnData:         process.BpmnData,
			bpmnChecksum:     [16]byte(process.BpmnChecksum),
			bpmnResourceName: process.BpmnResourceName,
		}

		var definitions bpmn20.TDefinitions

		data, err := decodeAndDecompress(string(process.BpmnData))
		if err != nil {
			log.Fatalf("Error decompressing data %s", err)
			return nil
		}
		err = xml.Unmarshal(data, &definitions)
		if err != nil {
			log.Fatalf("Error unmarshalling bpmn data: %s", err)
			return nil
		}
		resultProcess.definitions = definitions

		resultProcesses = append(resultProcesses, resultProcess)

	}

	return resultProcesses

}

func (persistence *BpmnEnginePersistenceRqlite) FindProcessInstanceByKey(processInstanceKey int64) *processInstanceInfo {
	processInstances := persistence.FindProcessInstances(processInstanceKey)
	if len(processInstances) == 0 {
		return nil
	}
	return processInstances[0]
}

func (persistence *BpmnEnginePersistenceRqlite) FindProcessInstances(processInstanceKey int64) []*processInstanceInfo {
	instances := persistence.rqlitePersistence.FindProcessInstances(processInstanceKey, -1)

	resultProcessInstances := make([]*processInstanceInfo, 0)

	for _, instance := range instances {
		// map to processInstanceInfo
		resultProcessInstance := &processInstanceInfo{
			InstanceKey: instance.Key,
			ProcessInfo: persistence.FindProcessByKey(int64(instance.ProcessDefinitionKey)),
			CreatedAt:   time.Unix(instance.CreatedAt, 0),
			State:       reverseMap(activityStateMap)[int(instance.State)],
		}

		variables := map[string]interface{}{}

		varHolder := var_holder.New(nil, nil)
		json.Unmarshal([]byte(instance.VariableHolder), &variables)
		for key, value := range variables {
			varHolder.SetVariable(key, value)
		}
		resultProcessInstance.VariableHolder = varHolder

		caughtEvents := []catchEvent{}
		json.Unmarshal([]byte(instance.CaughtEvents), &caughtEvents)
		resultProcessInstance.CaughtEvents = caughtEvents

		activities := []*activityAdapter{}
		json.Unmarshal([]byte(instance.Activities), &activities)
		recoverProcessInstanceActivitiesPartWithBaseElements(resultProcessInstance, activities)

		resultProcessInstances = append(resultProcessInstances, resultProcessInstance)

	}

	return resultProcessInstances

}

func convertActivityStatesToStrings(states []ActivityState) []string {
	result := make([]string, 0)
	for _, s := range states {
		result = append(result, string(s))
	}
	return result
}

func convertTimerStatesToStrings(states []TimerState) []string {
	result := make([]string, 0)
	for _, s := range states {
		result = append(result, string(s))
	}
	return result
}

func (persistence *BpmnEnginePersistenceRqlite) FindMessageSubscription(originActivityKey int64, processInstance *processInstanceInfo, elementId string, state ...ActivityState) []*MessageSubscription {
	states := convertActivityStatesToStrings(state)
	pik := int64(-1)

	if processInstance != nil {
		pik = (*processInstance).GetInstanceKey()
	}
	subscriptions := persistence.rqlitePersistence.FindMessageSubscription(originActivityKey, pik, elementId, states)

	resultSubscriptions := make([]*MessageSubscription, 0)

	for _, subscription := range subscriptions {

		pi := processInstance

		if processInstance == nil {
			pi = persistence.FindProcessInstanceByKey(subscription.ProcessInstanceKey)
		}

		resultSubscriptions = append(resultSubscriptions, &MessageSubscription{
			ElementId:          subscription.ElementID,
			ElementInstanceKey: subscription.ElementInstanceKey,
			ProcessKey:         subscription.ProcessKey,
			ProcessInstanceKey: subscription.ProcessInstanceKey,
			Name:               subscription.MessageName,
			MessageState:       reverseMap(activityStateMap)[subscription.State],
			CreatedAt:          time.Unix(subscription.CreatedAt, 0),
			originActivity:     constructOriginActivity(subscription.OriginActivityKey, subscription.OriginActivityState, subscription.OriginActivityId, pi.ProcessInfo),
			baseElement:        bpmn20.FindBaseElementsById(&pi.ProcessInfo.definitions, subscription.ElementID)[0],
		})
	}
	return resultSubscriptions

}

func constructOriginActivity(originActivityKey int64, originActivityState int, originActivityId string, process *ProcessInfo) activity {
	activity := &elementActivity{
		key:     originActivityKey,
		state:   reverseMap(activityStateMap)[originActivityState],
		element: bpmn20.FindBaseElementsById(&process.definitions, originActivityId)[0],
	}

	return activity

}

func (persistence *BpmnEnginePersistenceRqlite) FindTimers(originActivityKey int64, processInstanceKey int64, state ...TimerState) []*Timer {
	states := convertTimerStatesToStrings(state)
	timers := persistence.rqlitePersistence.FindTimers(originActivityKey, processInstanceKey, states)

	resultTimers := make([]*Timer, 0)
	for _, timer := range timers {

		resultTimers = append(resultTimers, &Timer{
			ElementId:          timer.ElementID,
			ElementInstanceKey: timer.ElementInstanceKey,
			ProcessKey:         timer.ProcessKey,
			ProcessInstanceKey: timer.ProcessInstanceKey,
			TimerState:         reverseMap(timerStateMap)[int(timer.TimerState)],
			CreatedAt:          time.Unix(timer.CreatedAt, 0),
			DueAt:              time.Unix(timer.DueDate, 0),
			Duration:           time.Duration(timer.Duration) * time.Second,
			//originActivity:     timer.OriginActivityKey,
			//baseElement: timer.ElementID,
		})
	}
	return resultTimers
}

func (persistence *BpmnEnginePersistenceRqlite) FindJobs(elementId string, processInstance *processInstanceInfo, jobKey int64, state ...ActivityState) []*job {
	states := convertActivityStatesToStrings(state)
	processInstanceKey := int64(-1)
	if processInstance != nil {
		processInstanceKey = (*processInstance).GetInstanceKey()
	}

	jobs := persistence.rqlitePersistence.FindJobs(elementId, processInstanceKey, jobKey, states)

	if processInstance == nil && len(jobs) > 0 {
		processInstance = persistence.FindProcessInstanceByKey(jobs[0].ProcessInstanceKey)
	}

	resultJobs := make([]*job, 0)
	for _, j := range jobs {
		resultJob := &job{
			ElementId:          j.ElementID,
			ElementInstanceKey: j.ElementInstanceKey,
			ProcessInstanceKey: j.ProcessInstanceKey,
			JobKey:             j.Key,
			JobState:           reverseMap(activityStateMap)[int(j.State)],
			CreatedAt:          time.Unix(j.CreatedAt, 0),
			//baseElement:        job.ElementID,

		}

		bes := bpmn20.FindBaseElementsById(&processInstance.ProcessInfo.definitions, resultJob.ElementId)
		if len(bes) == 0 {
			continue
		}
		resultJob.baseElement = bes[0]
		resultJobs = append(resultJobs, resultJob)
	}

	return resultJobs
}

func (persistence *BpmnEnginePersistenceRqlite) FindJobByKey(jobKey int64) *job {
	jobs := persistence.FindJobs("", nil, jobKey)

	if len(jobs) == 0 {
		return nil
	}
	return jobs[0]

}

// WRITE

func (persistence *BpmnEnginePersistenceRqlite) PersistNewProcess(processDefinition *ProcessInfo) error {

	return persistence.rqlitePersistence.PersistNewProcess(&sql.ProcessDefinitionEntity{
		Key:              processDefinition.ProcessKey,
		Version:          processDefinition.Version,
		BpmnProcessId:    processDefinition.BpmnProcessId,
		BpmnData:         base64.StdEncoding.EncodeToString([]byte(processDefinition.bpmnData)),
		BpmnChecksum:     []byte(base64.StdEncoding.EncodeToString(processDefinition.bpmnChecksum[:])),
		BpmnResourceName: processDefinition.bpmnResourceName,
	})

}

func (persistence *BpmnEnginePersistenceRqlite) PersistProcessInstance(processInstance *processInstanceInfo) error {
	varaiblesJson, err := json.Marshal(processInstance.VariableHolder.Variables())
	if err != nil {
		log.Fatalf("Error serializing variables: %s", err)
	}

	caughtEvents, err := json.Marshal(processInstance.CaughtEvents)
	if err != nil {
		log.Fatalf("Error serializing caught events: %s", err)
	}

	activityAdapters := make([]*activityAdapter, 0)

	for _, a := range processInstance.activities {
		switch activity := a.(type) {
		case *gatewayActivity:
			activityAdapters = append(activityAdapters, createGatewayActivityAdapter(activity))
		case *eventBasedGatewayActivity:
			activityAdapters = append(activityAdapters, createEventBasedGatewayActivityAdapter(activity))
		default:
			panic(fmt.Sprintf("[invariant check] missing activity adapter for the type %T", a))
		}
	}

	activities, err := json.Marshal(activityAdapters)
	if err != nil {
		log.Fatalf("Error serializing activities: %s", err)
	}

	return persistence.rqlitePersistence.PersistProcessInstance(&sql.ProcessInstanceEntity{
		Key:                  processInstance.InstanceKey,
		ProcessDefinitionKey: processInstance.ProcessInfo.ProcessKey,
		CreatedAt:            processInstance.CreatedAt.Unix(),
		State:                activityStateMap[processInstance.State],
		VariableHolder:       string(varaiblesJson),
		CaughtEvents:         string(caughtEvents),
		Activities:           string(activities),
	})

}

func (persistence *BpmnEnginePersistenceRqlite) PersistNewMessageSubscription(subscription *MessageSubscription) error {

	ms :=
		&sql.MessageSubscriptionEntity{
			ElementID:          subscription.ElementId,
			ElementInstanceKey: subscription.ElementInstanceKey,
			ProcessKey:         subscription.ProcessKey,
			ProcessInstanceKey: subscription.ProcessInstanceKey,
			MessageName:        subscription.Name,
			State:              activityStateMap[subscription.State()],
			CreatedAt:          subscription.CreatedAt.Unix(),
		}

	if subscription.originActivity != nil {
		ms.OriginActivityKey = subscription.originActivity.Key()
		ms.OriginActivityState = activityStateMap[subscription.originActivity.State()]
		ms.OriginActivityId = (*subscription.originActivity.Element()).GetId()
	}

	return persistence.rqlitePersistence.PersistNewMessageSubscription(ms)
}

func (persistence *BpmnEnginePersistenceRqlite) PersistNewTimer(timer *Timer) error {

	return persistence.rqlitePersistence.PersistNewTimer(&sql.TimerEntity{
		ElementID:          timer.ElementId,
		ElementInstanceKey: timer.ElementInstanceKey,
		ProcessKey:         timer.ProcessKey,
		ProcessInstanceKey: timer.ProcessInstanceKey,
		TimerState:         int64(timerStateMap[timer.TimerState]),
		CreatedAt:          timer.CreatedAt.Unix(),
		DueDate:            timer.DueAt.Unix(),
		Duration:           int64(timer.Duration.Seconds()),
	})
}

func (persistence *BpmnEnginePersistenceRqlite) PersistJob(job *job) error {
	return persistence.rqlitePersistence.PersistJob(&sql.JobEntity{
		Key:                job.JobKey,
		ElementID:          job.ElementId,
		ElementInstanceKey: job.ElementInstanceKey,
		ProcessInstanceKey: job.ProcessInstanceKey,
		State:              int64(activityStateMap[job.JobState]),
		CreatedAt:          job.CreatedAt.Unix(),
	})

}

func (persistence *BpmnEnginePersistenceRqlite) GetPersistence() *rqlite.BpmnEnginePersistenceRqlite {
	return persistence.rqlitePersistence
}

var activityStateMap = map[ActivityState]int{
	Active:       1,
	Compensated:  2,
	Compensating: 3,
	Completed:    4,
	Completing:   5,
	Failed:       6,
	Failing:      7,
	Ready:        8,
	Terminated:   9,
	Terminating:  10,
	WithDrawn:    11,
}

// reverse the map
func reverseMap[K comparable, V comparable](m map[K]V) map[V]K {
	rm := make(map[V]K)
	for k, v := range m {
		rm[v] = k
	}
	return rm
}

var timerStateMap = map[TimerState]int{
	TimerCreated:   1,
	TimerTriggered: 2,
	TimerCancelled: 3,
}
