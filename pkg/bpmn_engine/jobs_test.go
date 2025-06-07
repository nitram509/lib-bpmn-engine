package bpmn_engine

import (
	"github.com/corbym/gocrest"
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"
	"testing"

	"github.com/corbym/gocrest/has"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
)

const (
	varCounter                  = "counter"
	varEngineValidationAttempts = "engineValidationAttempts"
	varHasReachedMaxAttempts    = "hasReachedMaxAttempts"
)

func increaseCounterHandler(job ActivatedJob) {
	counter := job.Variable(varCounter).(float64)
	counter = counter + 1
	job.SetVariable(varCounter, counter)
	job.Complete()
}

func jobFailHandler(job ActivatedJob) {
	job.Fail("just because I can")
}

func jobCompleteHandler(job ActivatedJob) {
	job.Complete()
}

func Test_job_implements_Activity(t *testing.T) {
	var _ activity = &job{}
}

func Test_a_job_can_fail_and_keeps_fails_the_instance(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(jobFailHandler)

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.ActivityState, is.EqualTo(Failed))
}

// Test_simple_count_loop requires correct Task-Output-Mapping in the BPMN file
func Test_simple_count_loop(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-count-loop.bpmn")
	bpmnEngine.NewTaskHandler().Id("id-increaseCounter").Handler(increaseCounterHandler)

	vars := map[string]interface{}{}
	vars[varCounter] = 0.0
	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, vars)

	then.AssertThat(t, instance.GetVariable(varCounter), is.EqualTo(4.0))
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Completed))
}

func Test_simple_count_loop_with_message(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-count-loop-with-message.bpmn")

	vars := map[string]interface{}{}
	vars[varEngineValidationAttempts] = 0.0
	bpmnEngine.NewTaskHandler().Id("do-nothing").Handler(jobCompleteHandler)
	bpmnEngine.NewTaskHandler().Id("validate").Handler(func(job ActivatedJob) {
		attempts := job.Variable(varEngineValidationAttempts).(float64)
		foobar := attempts >= 1
		attempts++
		job.SetVariable(varEngineValidationAttempts, attempts)
		job.SetVariable(varHasReachedMaxAttempts, foobar)
		job.Complete()
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, vars) // should stop at the intermediate message catch event

	_ = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg", nil)
	_, _ = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey()) // again, should stop at the intermediate message catch event
	// validation happened
	_ = bpmnEngine.PublishEventForInstance(instance.GetInstanceKey(), "msg", nil)
	_, _ = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey()) // should finish
	// validation happened

	then.AssertThat(t, instance.GetVariable(varHasReachedMaxAttempts), is.True())
	then.AssertThat(t, instance.GetVariable(varEngineValidationAttempts), is.EqualTo(2.0))
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Completed))

	// internal State expected
	then.AssertThat(t, bpmnEngine.GetMessageSubscriptions(), has.Length(2))
	then.AssertThat(t, bpmnEngine.GetMessageSubscriptions()[0].MessageState, is.EqualTo(Completed))
	then.AssertThat(t, bpmnEngine.GetMessageSubscriptions()[1].MessageState, is.EqualTo(Completed))
}

func Test_activated_job_data(t *testing.T) {
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(aj ActivatedJob) {
		then.AssertThat(t, aj.ElementId(), is.Not(is.Empty()))
		then.AssertThat(t, aj.CreatedAt(), is.Not(is.Nil()))
		then.AssertThat(t, aj.Key(), is.Not(is.EqualTo(int64(0))))
		then.AssertThat(t, aj.BpmnProcessId(), is.Not(is.Empty()))
		then.AssertThat(t, aj.ProcessDefinitionKey(), is.Not(is.EqualTo(int64(0))))
		then.AssertThat(t, aj.ProcessDefinitionVersion(), is.Not(is.EqualTo(int32(0))))
		then.AssertThat(t, aj.ProcessInstanceKey(), is.Not(is.EqualTo(int64(0))))
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)

	then.AssertThat(t, instance.ActivityState, is.EqualTo(Active))
}

func Test_task_InputOutput_mapping_happy_path(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-input-output.bpmn")
	bpmnEngine.NewTaskHandler().Id("service-task-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("user-task-2").Handler(cp.TaskHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	for _, job := range bpmnEngine.jobs {
		then.AssertThat(t, job.JobState, is.EqualTo(Completed))
	}
	then.AssertThat(t, cp.CallPath, is.EqualTo("service-task-1,user-task-2"))
	// id from input should not exist in instance scope
	then.AssertThat(t, pi.GetVariable("id"), is.Nil())
	// output should exist in instance scope
	then.AssertThat(t, pi.GetVariable("dstcity"), is.EqualTo("beijing"))
	then.AssertThat(t, pi.GetVariable("order"), is.EqualTo(map[string]interface{}{
		"name": "order1",
		"id":   "1234",
	}))
	then.AssertThat(t, pi.GetVariable("orderId"), is.EqualTo(1234.0))
	then.AssertThat(t, pi.GetVariable("orderName"), is.EqualTo("order1"))
}

func Test_instance_fails_on_Invalid_Input_mapping(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-input.bpmn")
	bpmnEngine.NewTaskHandler().Id("invalid-input").Handler(cp.TaskHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err.Error(), is.Not(is.Nil()))

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo(""))
	then.AssertThat(t, pi.GetVariable("id"), is.Nil())
	then.AssertThat(t, bpmnEngine.jobs[0].JobState, is.EqualTo(Failed))
	then.AssertThat(t, pi.GetState(), is.EqualTo(Failed))
}

func Test_job_fails_on_Invalid_Output_mapping(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/service-task-invalid-output.bpmn")
	bpmnEngine.NewTaskHandler().Id("invalid-output").Handler(cp.TaskHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("invalid-output"))
	then.AssertThat(t, pi.GetVariable("order"), is.Nil())
	then.AssertThat(t, bpmnEngine.jobs[0].JobState, is.EqualTo(Failed))
	then.AssertThat(t, pi.GetState(), is.EqualTo(Failed))
}

func Test_task_type_handler(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}

	// give
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-task-with-type.bpmn")
	bpmnEngine.NewTaskHandler().Type("foobar").Handler(cp.TaskHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(Completed))
}

func Test_task_type_handler_ID_handler_has_precedence(t *testing.T) {
	// setup
	bpmnEngine := New()
	calledHandler := "none"
	idHandler := func(job ActivatedJob) {
		calledHandler = "ID"
		job.Complete()
	}
	typeHandler := func(job ActivatedJob) {
		calledHandler = "TYPE"
		job.Complete()
	}
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-task-with-type.bpmn")

	// given reverse order of definition, means 'type:foobar' before 'id'
	bpmnEngine.NewTaskHandler().Type("foobar").Handler(typeHandler)
	bpmnEngine.NewTaskHandler().Id("id").Handler(idHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, calledHandler, is.EqualTo("ID"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(Completed))
}

func Test_just_one_handler_called(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple-task-with-type.bpmn")

	// given multiple matching handlers executed
	bpmnEngine.NewTaskHandler().Id("id").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Type("foobar").Handler(cp.TaskHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("id").Reason("just one execution"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(Completed))
}

func Test_assignee_and_candidate_groups_are_assigned_to_handler(t *testing.T) {
	// setup
	bpmnEngine := New()
	cp := CallPath{}
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/user-tasks-with-assignments.bpmn")

	// given multiple matching handlers executed
	bpmnEngine.NewTaskHandler().Assignee("john.doe").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().CandidateGroups("marketing", "support").Handler(cp.TaskHandler)

	// when
	pi, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())

	// then
	then.AssertThat(t, cp.CallPath, is.EqualTo("assignee-task,group-task"))
	then.AssertThat(t, pi.GetState(), is.EqualTo(Completed))
}

func Test_task_default_all_output_variables_map_to_process_instance(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-no_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		job.SetVariable("aVariable", true)
		job.Complete()
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Completed))

	then.AssertThat(t, instance.GetVariable("aVariable"), is.True())
}

func Test_task_no_output_variables_mapping_on_failure(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-no_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		job.SetVariable("aVariable", true)
		job.Fail("because I can")
	})

	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Failed))
	then.AssertThat(t, err.Error(), is.EqualTo("because I can"))

	then.AssertThat(t, instance.GetVariable("aVariable"), is.Nil())
}

func Test_task_just_declared_output_variables_map_to_process_instance(t *testing.T) {
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-with_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		job.SetVariable("valueFromHandler", true)
		job.SetVariable("otherVariable", "value")
		job.Complete()
	})

	instance, _ := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Completed))

	then.AssertThat(t, instance.GetVariable("valueFromHandler"), is.True())
	then.AssertThat(t, instance.GetVariable("otherVariable"), is.Nil())
}

func Test_missing_task_handlers_break_execution_and_can_be_continued_later(t *testing.T) {
	cp := CallPath{}
	// setup
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/parallel-gateway-flow.bpmn")

	// given
	bpmnEngine.NewTaskHandler().Id("id-a-1").Handler(cp.TaskHandler)
	instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Active))
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1"))

	// when
	bpmnEngine.NewTaskHandler().Id("id-b-1").Handler(cp.TaskHandler)
	bpmnEngine.NewTaskHandler().Id("id-b-2").Handler(cp.TaskHandler)
	instance, err = bpmnEngine.RunOrContinueInstance(instance.GetInstanceKey())
	then.AssertThat(t, instance, is.Not(is.Nil()))
	then.AssertThat(t, instance.ActivityState, is.EqualTo(Completed))

	// then
	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, cp.CallPath, is.EqualTo("id-a-1,id-b-1,id-b-2"))
}

func Test_error_boundary_event(t *testing.T) {

	type handler struct {
		fn func(job ActivatedJob)
		id string
	}

	type varAssertion struct {
		assertion *gocrest.Matcher
		key       string
	}

	type args struct {
		file     string
		handlers []handler
	}

	type wants struct {
		instanceState ActivityState
		processError  *gocrest.Matcher
		varAssertions []varAssertion
		paths         []string
	}

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "Single boundary error event",
			args: args{
				file: "../../test-cases/error-boundary-event.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error1")
						},
					},
					{
						id: "handle-error-task",
						fn: func(job ActivatedJob) {
							job.Complete()
						},
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"error-boundary-event(ELEMENT_ACTIVATED)",
					"error-boundary-event(ELEMENT_COMPLETED)",
					"flow-handle-error(SEQUENCE_FLOW_TAKEN)",
					"handle-error-task(ELEMENT_ACTIVATED)",
					"handle-error-task(ELEMENT_COMPLETED)",
					"flow-handled-error(SEQUENCE_FLOW_TAKEN)",
					"handled-error-end(ELEMENT_ACTIVATED)",
					"handled-error-end(ELEMENT_COMPLETED)",
					"handled-error-end(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Catchall boundary error event with unknown error",
			args: args{
				file: "../../test-cases/error-boundary-event-catchall.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("unknown_error")
						},
					},
					{
						id: "handle-error-task",
						fn: func(job ActivatedJob) {
							job.Complete()
						},
					},
				},
			},
			wants: wants{
				instanceState: Failed,
				processError:  is.EqualTo(newEngineErrorf("Could not find error definition \"unknown_error\"")),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Catchall boundary error event",
			args: args{
				file: "../../test-cases/error-boundary-event-catchall.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error1")
						},
					},
					{
						id: "handle-error-task",
						fn: func(job ActivatedJob) {
							job.Complete()
						},
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"error-boundary-event(ELEMENT_ACTIVATED)",
					"error-boundary-event(ELEMENT_COMPLETED)",
					"flow-handle-error(SEQUENCE_FLOW_TAKEN)",
					"handle-error-task(ELEMENT_ACTIVATED)",
					"handle-error-task(ELEMENT_COMPLETED)",
					"flow-handled-error(SEQUENCE_FLOW_TAKEN)",
					"handled-error-end(ELEMENT_ACTIVATED)",
					"handled-error-end(ELEMENT_COMPLETED)",
					"handled-error-end(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Single boundary error event with output",
			args: args{
				file: "../../test-cases/error-boundary-event-outputs.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.SetVariable("errorValue", "something broke")
							job.ThrowError("error1")
						},
					},
					{
						id: "handle-error-task",
						fn: func(job ActivatedJob) {
							job.Complete()
						},
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
					{
						key:       "errorValue",
						assertion: is.EqualTo("something broke"),
					},
					{
						key:       "mappedErrorValue",
						assertion: is.EqualTo("something broke"),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"error-boundary-event(ELEMENT_ACTIVATED)",
					"error-boundary-event(ELEMENT_COMPLETED)",
					"flow-handle-error(SEQUENCE_FLOW_TAKEN)",
					"handle-error-task(ELEMENT_ACTIVATED)",
					"handle-error-task(ELEMENT_COMPLETED)",
					"flow-handled-error(SEQUENCE_FLOW_TAKEN)",
					"handled-error-end(ELEMENT_ACTIVATED)",
					"handled-error-end(ELEMENT_COMPLETED)",
					"handled-error-end(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Single boundary error event unknown error",
			args: args{
				file: "../../test-cases/error-boundary-event.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("unknown_error")
						},
					},
				},
			},
			wants: wants{
				instanceState: Failed,
				processError:  is.EqualTo(newEngineErrorf("Could not find error definition \"unknown_error\"")),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Multi boundary error event unknown error",
			args: args{
				file: "../../test-cases/error-boundary-event-multiple.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("unknown_error")
						},
					},
				},
			},
			wants: wants{
				instanceState: Failed,
				processError:  is.EqualTo(newEngineErrorf("Could not find error definition \"unknown_error\"")),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Multi boundary error event",
			args: args{
				file: "../../test-cases/error-boundary-event-multiple.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error1")
						},
					},
					{
						id: "handle-error-task",
						fn: func(job ActivatedJob) {
							job.Complete()
						},
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"error-boundary-event(ELEMENT_ACTIVATED)",
					"error-boundary-event(ELEMENT_COMPLETED)",
					"flow-handle-error(SEQUENCE_FLOW_TAKEN)",
					"handle-error-task(ELEMENT_ACTIVATED)",
					"handle-error-task(ELEMENT_COMPLETED)",
					"flow-handled-error(SEQUENCE_FLOW_TAKEN)",
					"handled-error-end(ELEMENT_ACTIVATED)",
					"handled-error-end(ELEMENT_COMPLETED)",
					"handled-error-end(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Multi boundary error event catchall",
			args: args{
				file: "../../test-cases/error-boundary-event-multiple.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error2")
						},
					},
					{
						id: "handle-all-task",
						fn: func(job ActivatedJob) {
							job.Complete()
						},
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"all-boundary-event(ELEMENT_ACTIVATED)",
					"all-boundary-event(ELEMENT_COMPLETED)",
					"flow-handle-all(SEQUENCE_FLOW_TAKEN)",
					"handle-all-task(ELEMENT_ACTIVATED)",
					"handle-all-task(ELEMENT_COMPLETED)",
					"flow-handled-all(SEQUENCE_FLOW_TAKEN)",
					"handled-all-end(ELEMENT_ACTIVATED)",
					"handled-all-end(ELEMENT_COMPLETED)",
					"handled-all-end(ELEMENT_COMPLETED)",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			bpmnEngine := New()
			pathRecordingExporter := NewPathRecordingEventExporter()
			bpmnEngine.AddEventExporter(pathRecordingExporter)
			bpmnEngine.AddEventExporter(exporter.NewEventLogExporter())
			process, _ := bpmnEngine.LoadFromFile(tt.args.file)
			for _, handler := range tt.args.handlers {
				bpmnEngine.NewTaskHandler().Id(handler.id).Handler(handler.fn)
			}
			instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
			then.AssertThat(t, instance.ActivityState, is.EqualTo(tt.wants.instanceState))
			then.AssertThat(t, err, tt.wants.processError)
			then.AssertThat(t, pathRecordingExporter.String(), is.EqualTo(pathString(tt.wants.paths)))

			for _, varAssert := range tt.wants.varAssertions {
				then.AssertThat(t, instance.GetVariable(varAssert.key), varAssert.assertion)
			}
		})
	}
}

func Test_error_event_subprocess(t *testing.T) {

	type handler struct {
		fn func(job ActivatedJob)
		id string
	}

	type varAssertion struct {
		assertion *gocrest.Matcher
		key       string
	}

	type args struct {
		file     string
		handlers []handler
	}

	type wants struct {
		instanceState ActivityState
		processError  *gocrest.Matcher
		varAssertions []varAssertion
		paths         []string
	}

	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "Single boundary error event",
			args: args{
				file: "../../test-cases/error-event-subprocess.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error1")
						},
					},
					{
						id: "handle-error-task",
						fn: jobCompleteHandler,
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"error-sub-process(ELEMENT_ACTIVATED)",
					"error1-start-event(ELEMENT_ACTIVATED)",
					"error1-start-event(ELEMENT_COMPLETED)",
					"flow-error1(SEQUENCE_FLOW_TAKEN)",
					"handle-error-task(ELEMENT_ACTIVATED)",
					"handle-error-task(ELEMENT_COMPLETED)",
					"flow-error1-end(SEQUENCE_FLOW_TAKEN)",
					"end-error1(ELEMENT_ACTIVATED)",
					"end-error1(ELEMENT_COMPLETED)",
					"end-error1(ELEMENT_COMPLETED)",
					"error-sub-process(ELEMENT_COMPLETED)",
					"error-event-subprocess(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			name: "Single boundary error event catchall",
			args: args{
				file: "../../test-cases/error-event-subprocess.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error2")
						},
					},
					{
						id: "handle-catchall-task",
						fn: jobCompleteHandler,
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"catchall-sub-process(ELEMENT_ACTIVATED)",
					"catchall-start-event(ELEMENT_ACTIVATED)",
					"catchall-start-event(ELEMENT_COMPLETED)",
					"flow-catchall(SEQUENCE_FLOW_TAKEN)",
					"handle-catchall-task(ELEMENT_ACTIVATED)",
					"handle-catchall-task(ELEMENT_COMPLETED)",
					"flow-catchall-end(SEQUENCE_FLOW_TAKEN)",
					"end-catchall(ELEMENT_ACTIVATED)",
					"end-catchall(ELEMENT_COMPLETED)",
					"end-catchall(ELEMENT_COMPLETED)",
					"catchall-sub-process(ELEMENT_COMPLETED)",
					"error-event-subprocess(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			// In this test we expect that the error boundary event is triggered
			name: "Multiple boundary error events and Subprocesses error1",
			args: args{
				file: "../../test-cases/error-boundary-event-and-subprocess.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error1")
						},
					},
					{
						id: "handle-error1-task",
						fn: jobCompleteHandler,
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"error-boundary-event(ELEMENT_ACTIVATED)",
					"error-boundary-event(ELEMENT_COMPLETED)",
					"flow-handle-error(SEQUENCE_FLOW_TAKEN)",
					"handle-error1-task(ELEMENT_ACTIVATED)",
					"handle-error1-task(ELEMENT_COMPLETED)",
					"flow-handled-error(SEQUENCE_FLOW_TAKEN)",
					"handled-error-end(ELEMENT_ACTIVATED)",
					"handled-error-end(ELEMENT_COMPLETED)",
					"handled-error-end(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			// In this test we expect that the error2 event subprocess is triggered
			name: "Multiple boundary error events and Subprocesses error2",
			args: args{
				file: "../../test-cases/error-boundary-event-and-subprocess.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error2")
						},
					},
					{
						id: "handle-error2-sub-task",
						fn: jobCompleteHandler,
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"error-sub-process(ELEMENT_ACTIVATED)",
					"error1-start-event(ELEMENT_ACTIVATED)",
					"error1-start-event(ELEMENT_COMPLETED)",
					"flow-error2(SEQUENCE_FLOW_TAKEN)",
					"handle-error2-sub-task(ELEMENT_ACTIVATED)",
					"handle-error2-sub-task(ELEMENT_COMPLETED)",
					"flow-error2-end(SEQUENCE_FLOW_TAKEN)",
					"end-error1(ELEMENT_ACTIVATED)",
					"end-error1(ELEMENT_COMPLETED)",
					"end-error1(ELEMENT_COMPLETED)",
					"error-sub-process(ELEMENT_COMPLETED)",
					"error-boundary-event-and-subprocess(ELEMENT_COMPLETED)",
				},
			},
		},
		{
			// In this test we expect that the catchall error boundary event is triggered
			name: "Multiple boundary error events and Subprocesses error3",
			args: args{
				file: "../../test-cases/error-boundary-event-and-subprocess.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("error3")
						},
					},
					{
						id: "handle-all-task",
						fn: jobCompleteHandler,
					},
				},
			},
			wants: wants{
				instanceState: Completed,
				processError:  is.Nil(),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
					"all-boundary-event(ELEMENT_ACTIVATED)",
					"all-boundary-event(ELEMENT_COMPLETED)",
					"flow-handle-all(SEQUENCE_FLOW_TAKEN)",
					"handle-all-task(ELEMENT_ACTIVATED)",
					"handle-all-task(ELEMENT_COMPLETED)",
					"flow-handled-all(SEQUENCE_FLOW_TAKEN)",
					"handled-all-end(ELEMENT_ACTIVATED)",
					"handled-all-end(ELEMENT_COMPLETED)",
					"handled-all-end(ELEMENT_COMPLETED)",
				},
			},
		}, {
			name: "Single boundary error event unknown error",
			args: args{
				file: "../../test-cases/error-event-subprocess.bpmn",
				handlers: []handler{
					{
						id: "error-task",
						fn: func(job ActivatedJob) {
							job.SetVariable("aVariable", true)
							job.ThrowError("unknown_error")
						},
					},
					{
						id: "handle-error-task",
						fn: jobCompleteHandler,
					},
				},
			},
			wants: wants{
				instanceState: Failed,
				processError:  is.EqualTo(newEngineErrorf("Could not find error definition \"unknown_error\"")),
				varAssertions: []varAssertion{
					{
						key:       "aVariable",
						assertion: is.True(),
					},
				},
				paths: []string{
					"start(ELEMENT_ACTIVATED)",
					"start(ELEMENT_COMPLETED)",
					"flow1(SEQUENCE_FLOW_TAKEN)",
					"error-task(ELEMENT_ACTIVATED)",
					"error-task(ELEMENT_COMPLETED)",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			bpmnEngine := New()
			pathExporter := NewPathRecordingEventExporter()
			bpmnEngine.AddEventExporter(pathExporter)
			bpmnEngine.AddEventExporter(exporter.NewEventLogExporter())
			process, _ := bpmnEngine.LoadFromFile(tt.args.file)
			for _, handler := range tt.args.handlers {
				bpmnEngine.NewTaskHandler().Id(handler.id).Handler(handler.fn)
			}
			instance, err := bpmnEngine.CreateAndRunInstance(process.ProcessKey, nil)
			then.AssertThat(t, instance.ActivityState, is.EqualTo(tt.wants.instanceState))
			then.AssertThat(t, err, tt.wants.processError)
			then.AssertThat(t, pathExporter.String(), is.EqualTo(pathString(tt.wants.paths)))

			for _, varAssert := range tt.wants.varAssertions {
				then.AssertThat(t, instance.GetVariable(varAssert.key), varAssert.assertion)
			}
		})
	}

}

// TODO boundaryEvent and eventSubProcesses where specific boundary event is selected
// TODO boundaryEvent and eventSubProcesses where specific sub process is selected
// TODO boundaryEvent and eventSubProcesses where boundaryevent catch all is selected
