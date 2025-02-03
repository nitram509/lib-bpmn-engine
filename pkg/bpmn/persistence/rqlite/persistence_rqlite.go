package rqlite

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"log"

	"github.com/bwmarrin/snowflake"
	bpmnEngineExporter "github.com/pbinitiative/zenbpm/pkg/bpmn/exporter"
	sql "github.com/pbinitiative/zenbpm/pkg/bpmn/persistence/rqlite/sql"
	"github.com/rqlite/rqlite/v8/command/proto"
	"github.com/rqlite/rqlite/v8/store"
)

type BpmnEnginePersistenceRqlite struct {
	snowflakeIdGenerator *snowflake.Node
	ctx                  *RqliteContext
}

func NewBpmnEnginePersistenceRqlite(snowflakeIdGenerator *snowflake.Node) *BpmnEnginePersistenceRqlite {
	gen := snowflakeIdGenerator
	context := Start()

	time.Sleep(2 * time.Second)

	Init(context.Str)

	return &BpmnEnginePersistenceRqlite{
		snowflakeIdGenerator: gen,
		ctx:                  context,
	}
}

func (persistence *BpmnEnginePersistenceRqlite) RqliteStop() {
	Stop(persistence.ctx)
}

// READ

func (persistence *BpmnEnginePersistenceRqlite) FindProcesses(processId string, processKey int64) []*sql.ProcessDefinitionEntity {
	// TODO finds all processes with given ID sorted by version number

	filters := map[string]string{}

	if processId != "" {
		filters["bpmn_process_id"] = fmt.Sprintf("\"%s\"", processId)
	}
	if processKey != -1 {
		filters["key"] = fmt.Sprintf("%d", processKey)
	}

	whereClause := generateWhereClause(filters)

	queryStr := fmt.Sprintf(sql.PROCESS_DEFINITION, whereClause)

	rows, err := query(queryStr, persistence.ctx.Str)
	if err != nil {
		log.Fatalf("Error executing SQL statements %s", err)
		return nil
	}
	log.Printf("Result: %T", rows)

	var processDefinitions []*sql.ProcessDefinitionEntity = make([]*sql.ProcessDefinitionEntity, 0)

	for _, qr := range rows {

		for _, r := range qr.Values {
			def := new(sql.ProcessDefinitionEntity)

			def.Key = (*r.Parameters[0]).GetI()
			def.Version = int32((*r.Parameters[1]).GetI())
			def.BpmnProcessId = (*r.Parameters[2]).GetS()
			dataString := (*r.Parameters[3]).GetS()

			encodedXml, err := base64.StdEncoding.DecodeString(dataString)
			if err != nil {
				log.Fatalf("Error base64 decoding bpmn data: %s", err)
				return nil
			}

			def.BpmnData = string(encodedXml)

			bytes, err := base64.StdEncoding.DecodeString((*r.Parameters[4]).GetS())

			if err != nil {
				log.Fatalf("Error decoding checksum: %s", err)
				return nil
			}

			def.BpmnChecksum = bytes
			// copy(def.BpmnChecksum[:], bytes)

			def.BpmnResourceName = (*r.Parameters[5]).GetS()

			processDefinitions = append(processDefinitions, def)

		}

	}
	return processDefinitions
}

func (persistence *BpmnEnginePersistenceRqlite) FindProcessInstances(processInstanceKey int64, processDefinitionKey int64) []*sql.ProcessInstanceEntity {
	filters := map[string]string{}

	if processInstanceKey != -1 {
		filters["key"] = fmt.Sprintf("%d", processInstanceKey)
	}

	if processDefinitionKey != -1 {
		filters["process_definition_key"] = fmt.Sprintf("%d", processDefinitionKey)
	}

	whereClause := generateWhereClause(filters)

	queryStr := fmt.Sprintf(sql.PROCESS_INSTANCE_SELECT, whereClause)

	rows, err := query(queryStr, persistence.ctx.Str)
	if err != nil {
		log.Fatalf("Error executing SQL statements %s", err)
		return nil
	}
	log.Printf("Result: %T", rows)

	var processInstances []*sql.ProcessInstanceEntity = make([]*sql.ProcessInstanceEntity, 0)

	for _, qr := range rows {

		for _, r := range qr.Values {
			def := new(sql.ProcessInstanceEntity)

			def.Key = (*r.Parameters[0]).GetI()
			def.ProcessDefinitionKey = int64((*r.Parameters[1]).GetI())
			def.CreatedAt = (*r.Parameters[2]).GetI()

			def.State = int((*r.Parameters[3]).GetI())
			def.VariableHolder = (*r.Parameters[4]).GetS()
			def.CaughtEvents = (*r.Parameters[5]).GetS()

			def.Activities = (*r.Parameters[6]).GetS()

			processInstances = append(processInstances, def)

		}

	}
	return processInstances
}

func (persistence *BpmnEnginePersistenceRqlite) FindMessageSubscription(originActivityKey int64, processInstanceKey int64, elementId string, state []string) []*sql.MessageSubscriptionEntity {
	filters := map[string]string{}

	if originActivityKey != -1 {
		filters["origin_activity_key"] = fmt.Sprintf("%d", originActivityKey)
	}
	if processInstanceKey != -1 {
		filters["process_instance_key"] = fmt.Sprintf("%d", processInstanceKey)
	}
	if elementId != "" {
		filters["element_id"] = fmt.Sprintf("\"%s\"", elementId)
	}

	statesClause := ""
	if len(state) > 0 {
		states := map[string]string{}
		// for each state
		for _, s := range state {
			states["state"] = fmt.Sprintf("%d", activityStateMap[s])
		}
		statesClause = fmt.Sprintf(" AND (%s)", whereClauseBuilder(states, "OR"))
	}

	whereClause := generateWhereClause(filters) + statesClause

	queryStr := fmt.Sprintf(sql.MESSAGE_SUBSCRIPTION_SELECT, whereClause)

	rows, err := query(queryStr, persistence.ctx.Str)
	if err != nil {
		log.Fatalf("Error executing SQL statements %s", err)
		return nil
	}
	log.Printf("Result: %T", rows)

	var messageSubscriptions []*sql.MessageSubscriptionEntity = make([]*sql.MessageSubscriptionEntity, 0)

	for _, qr := range rows {

		for _, r := range qr.Values {
			def := new(sql.MessageSubscriptionEntity)
			def.ElementInstanceKey = (*r.Parameters[0]).GetI()
			def.ElementID = (*r.Parameters[1]).GetS()
			def.ProcessKey = (*r.Parameters[2]).GetI()
			def.ProcessInstanceKey = (*r.Parameters[3]).GetI()
			def.MessageName = (*r.Parameters[4]).GetS()
			def.State = int((*r.Parameters[5]).GetI())
			def.CreatedAt = (*r.Parameters[6]).GetI()
			def.OriginActivityKey = int64((*r.Parameters[7]).GetI())
			def.OriginActivityState = int((*r.Parameters[8]).GetI())
			def.OriginActivityId = (*r.Parameters[9]).GetS()

			messageSubscriptions = append(messageSubscriptions, def)

		}

	}
	return messageSubscriptions
}

func (persistence *BpmnEnginePersistenceRqlite) FindTimers(originActivityKey int64, processInstanceKey int64, state []string) []*sql.TimerEntity {
	filters := map[string]string{}

	if originActivityKey != -1 {
		filters["origin_activity_key"] = fmt.Sprintf("%d", originActivityKey)
	}
	if processInstanceKey != -1 {
		filters["process_instance_key"] = fmt.Sprintf("%d", processInstanceKey)
	}

	statesClause := ""
	if len(state) > 0 {
		states := map[string]string{}
		// for each state
		for _, s := range state {
			states["state"] = fmt.Sprintf("%d", timerStateMap[s])
		}
		statesClause = fmt.Sprintf(" AND (%s)", whereClauseBuilder(states, "OR"))
	}

	whereClause := generateWhereClause(filters) + statesClause

	queryStr := fmt.Sprintf(sql.TIMER_SELECT, whereClause)

	rows, err := query(queryStr, persistence.ctx.Str)
	if err != nil {
		log.Fatalf("Error executing SQL statements %s", err)
		return nil
	}
	log.Printf("Result: %T", rows)

	var timers []*sql.TimerEntity = make([]*sql.TimerEntity, 0)

	for _, qr := range rows {

		for _, r := range qr.Values {
			def := new(sql.TimerEntity)
			def.ElementID = (*r.Parameters[0]).GetS()
			def.ElementInstanceKey = (*r.Parameters[1]).GetI()
			def.ProcessKey = (*r.Parameters[2]).GetI()
			def.ProcessInstanceKey = (*r.Parameters[3]).GetI()
			def.TimerState = (*r.Parameters[4]).GetI()
			def.CreatedAt = (*r.Parameters[5]).GetI()
			def.DueDate = (*r.Parameters[6]).GetI()
			def.Duration = (*r.Parameters[7]).GetI()

			timers = append(timers, def)

		}

	}
	return timers
}

func (persistence *BpmnEnginePersistenceRqlite) FindJobs(elementId string, processInstanceKey int64, jobKey int64, state []string) []*sql.JobEntity {
	filters := map[string]string{}

	if elementId != "" {
		filters["element_id"] = fmt.Sprintf("\"%s\"", elementId)
	}
	if processInstanceKey != -1 {
		filters["process_instance_key"] = fmt.Sprintf("%d", processInstanceKey)
	}
	if jobKey != -1 {
		filters["key"] = fmt.Sprintf("%d", jobKey)
	}

	statesClause := ""
	if len(state) > 0 {
		states := map[string]string{}
		// for each state
		for _, s := range state {
			states["state"] = fmt.Sprintf("%d", activityStateMap[s])
		}
		statesClause = fmt.Sprintf(" AND (%s)", whereClauseBuilder(states, "OR"))
	}

	whereClause := generateWhereClause(filters) + statesClause

	queryStr := fmt.Sprintf(sql.JOB_SELECT, whereClause)

	rows, err := query(queryStr, persistence.ctx.Str)
	if err != nil {
		log.Fatalf("Error executing SQL statements %s", err)
		return nil
	}
	log.Printf("Result: %T", rows)

	var jobs []*sql.JobEntity = make([]*sql.JobEntity, 0)

	for _, qr := range rows {

		for _, r := range qr.Values {
			def := new(sql.JobEntity)

			def.Key = (*r.Parameters[0]).GetI()
			def.ElementID = (*r.Parameters[1]).GetS()
			def.ElementInstanceKey = int64((*r.Parameters[2]).GetI())
			def.ProcessInstanceKey = int64((*r.Parameters[3]).GetI())
			def.State = (*r.Parameters[4]).GetI()
			def.CreatedAt = (*r.Parameters[5]).GetI()

			jobs = append(jobs, def)

		}

	}
	return jobs
}

func (persistence *BpmnEnginePersistenceRqlite) FindActivitiesByProcessInstanceKey(processInstanceKey int64) []*sql.ActivityInstanceEntity {
	filters := map[string]string{}

	if processInstanceKey != -1 {
		filters["process_instance_key"] = fmt.Sprintf("%d", processInstanceKey)
	}

	whereClause := generateWhereClause(filters)

	queryStr := fmt.Sprintf(sql.ACTIVITY_INSTANCE_SELECT, whereClause)

	rows, err := query(queryStr, persistence.ctx.Str)
	if err != nil {
		log.Fatalf("Error executing SQL statements %s", err)
		return nil
	}
	log.Printf("Result: %T", rows)

	var activites []*sql.ActivityInstanceEntity = make([]*sql.ActivityInstanceEntity, 0)

	for _, qr := range rows {

		for _, r := range qr.Values {
			def := new(sql.ActivityInstanceEntity)
			// key, process_instance_key, process_definition_key, created_at, state, element_id, bpmn_element_type

			def.Key = (*r.Parameters[0]).GetI()
			def.ProcessInstanceKey = int64((*r.Parameters[1]).GetI())
			def.ProcessDefinitionKey = int64((*r.Parameters[2]).GetI())
			def.CreatedAt = (*r.Parameters[3]).GetI()
			def.State = (*r.Parameters[4]).GetS()
			def.ElementId = (*r.Parameters[5]).GetS()
			def.BpmnElementType = (*r.Parameters[6]).GetS()

			activites = append(activites, def)

		}

	}
	return activites
}

// WRITE

func (persistence *BpmnEnginePersistenceRqlite) PersistNewProcess(processDefinition *sql.ProcessDefinitionEntity) error {

	sql := sql.BuildProcessDefinitionUpsertQuery(processDefinition)

	log.Printf("Creating process definition: %s", sql)

	_, err := execute([]string{sql}, persistence.ctx.Str)
	if err != nil {
		log.Fatalf("Error executing SQL statements: %s", err)
		return err
	}
	return nil

}

func (persistence *BpmnEnginePersistenceRqlite) PersistProcessInstance(processInstance *sql.ProcessInstanceEntity) error {

	sql := sql.BuildProcessInstanceUpsertQuery(processInstance)

	log.Printf("Creating process instance: %s", sql)
	_, err := execute([]string{sql}, persistence.ctx.Str)

	if err != nil {
		log.Panicf("Error executing SQL statements")
		return err
	}
	return nil

}

func (persistence *BpmnEnginePersistenceRqlite) PersistNewMessageSubscription(subscription *sql.MessageSubscriptionEntity) error {
	sql := sql.BuildMessageSubscriptionUpsertQuery(subscription)

	log.Printf("Creating message subscription: %s", sql)
	_, err := execute([]string{sql}, persistence.ctx.Str)

	if err != nil {
		log.Panicf("Error executing SQL statements")
		return err
	}
	return nil
}

func (persistence *BpmnEnginePersistenceRqlite) PersistNewTimer(timer *sql.TimerEntity) error {
	sql := sql.BuildTimerUpsertQuery(timer)

	log.Printf("Creating timer: %s", sql)
	_, err := execute([]string{sql}, persistence.ctx.Str)

	if err != nil {
		log.Panicf("Error executing SQL statements")
		return err
	}
	return nil
}

func (persistence *BpmnEnginePersistenceRqlite) PersistJob(job *sql.JobEntity) error {
	sql := sql.BuildJobUpsertQuery(job)

	log.Printf("Creating job: %s", sql)
	_, err := execute([]string{sql}, persistence.ctx.Str)

	if err != nil {
		log.Panicf("Error executing SQL statements")
		return err
	}
	return nil
}

func (persistence *BpmnEnginePersistenceRqlite) PersistActivity(event *bpmnEngineExporter.ProcessInstanceEvent, elementInfo *bpmnEngineExporter.ElementInfo) error {
	sql := sql.BuildActivityInstanceUpsertQuery(persistence.snowflakeIdGenerator.Generate().Int64(), event.ProcessInstanceKey, event.ProcessKey, time.Now().Unix(), elementInfo.Intent, elementInfo.ElementId, elementInfo.BpmnElementType)

	log.Printf("Creating activity log: %s", sql)
	_, err := execute([]string{sql}, persistence.ctx.Str)

	if err != nil {
		log.Panicf("Error executing SQL statements")
		return err
	}
	return nil
}

func (persistence *BpmnEnginePersistenceRqlite) IsLeader() bool {
	return persistence.ctx.Str.IsLeader()
}

func (persistence *BpmnEnginePersistenceRqlite) GetLeaderAddress() string {
	leaderAddr, err := persistence.ctx.Str.LeaderAddr()

	if err != nil {
		log.Panicf("Error while reading Leader Address: %s", err)
		return ""
	}
	return leaderAddr

}

func Init(str *store.Store) {
	log.Printf("Is leader: %v", str.IsLeader())
	if !str.IsLeader() {
		log.Println("Not a leader, skipping init")
		return
	}

	log.Println("Initing database!")

	statements := make([]string, 0)
	statements = append(statements, sql.PROCESS_DEFINITION_TABLE_CREATE)
	statements = append(statements, sql.PROCESS_INSTANCE_TABLE_CREATE)
	statements = append(statements, sql.JOB_TABLE_CREATE)
	statements = append(statements, sql.MESSAGE_SUBSCRIPTION_TABLE_CREATE)
	statements = append(statements, sql.TIMER_TABLE_CREATE)
	statements = append(statements, sql.ACTIVITY_INSTANCE_TABLE_CREATE)

	_, err := execute(statements, str)
	if err != nil {
		log.Fatalf("Error executing SQL statements %s", err)
	}

}

func generateWhereClause(mappings map[string]string) string {
	where := whereClauseBuilder(mappings, "AND")

	if where == "" {
		return "1=1"
	}
	return where
}

func whereClauseBuilder(mappings map[string]string, operator string) string {
	wheres := make([]string, 0)
	for k, v := range mappings {
		wheres = append(wheres, fmt.Sprintf("%s = %s", k, v))
	}
	return strings.Join(wheres, " "+operator+" ")

}

func execute(statements []string, str *store.Store) ([]*proto.ExecuteQueryResponse, error) {
	stmts := generateStatments(statements)

	er := &proto.ExecuteRequest{
		Request: &proto.Request{
			Transaction: true,
			DbTimeout:   int64(0),
			Statements:  stmts,
		},
		Timings: false,
	}

	results, resultsErr := str.Execute(er)

	if resultsErr != nil {
		log.Panicf("Error executing SQL statements %s", resultsErr)
		return nil, resultsErr
	}
	log.Printf("Result: %v", results)
	return results, nil
}

func generateStatments(statements []string) []*proto.Statement {
	stmts := make([]*proto.Statement, len(statements))
	for i := range statements {
		stmts[i] = &proto.Statement{
			Sql: statements[i],
		}
	}
	return stmts
}

func query(query string, str *store.Store) ([]*proto.QueryRows, error) {

	stmts := generateStatments([]string{query})

	qr := &proto.QueryRequest{
		Request: &proto.Request{
			Transaction: false,
			DbTimeout:   int64(0),
			Statements:  stmts,
		},
		Timings: false,
		Level:   proto.QueryRequest_QUERY_REQUEST_LEVEL_NONE,
		// TODO: this needs to be revised
		Freshness:       1000000000,
		FreshnessStrict: false,
	}

	results, resultsErr := str.Query(qr)
	if resultsErr != nil {
		log.Fatalf("Error executing SQL statements %s", resultsErr)
		return nil, resultsErr
	}
	log.Printf("Result: %v", results)
	return results, nil

}

var activityStateMap = map[string]int{
	"ACTIVE":       1,
	"COMPENSATED":  2,
	"COMPENSATING": 3,
	"COMPLETED":    4,
	"COMPLETING":   5,
	"FAILED":       6,
	"FAILING":      7,
	"READY":        8,
	"TERMINATED":   9,
	"TERMINATING":  10,
	"WITHDRAWN":    11,
}

// reverse the map
func reverseMap[K comparable, V comparable](m map[K]V) map[V]K {
	rm := make(map[V]K)
	for k, v := range m {
		rm[v] = k
	}
	return rm
}

var timerStateMap = map[string]int{
	"TIMERCREATED":   1,
	"TIMERTIGGERED":  2,
	"TIMERCANCELLED": 3,
}
