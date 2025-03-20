package bpmn_engine

import (
	"fmt"
	"github.com/corbym/gocrest/is"
	"github.com/pbinitiative/feel"
	"reflect"
	"testing"

	"github.com/corbym/gocrest/then"
)

type TestStruct struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age,omitempty"`
}

func Test_MarshallEngine(t *testing.T) {
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-with_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		job.SetVariable("valueFromHandler", true)
		job.SetVariable("otherVariable", "value")
		job.Complete()
	})
	variableContext := make(map[string]interface{})
	variableContext["hello"] = "world"
	variableContext["john"] = "doe"

	_, _ = bpmnEngine.CreateAndRunInstance(process.ProcessKey, variableContext)

	data, err := bpmnEngine.Marshal()
	then.AssertThat(t, err, is.Nil())
	fmt.Println(string(data))

	bpmnEngine, _ = Unmarshal(data)
	vars := bpmnEngine.ProcessInstances()[0].VariableHolder
	then.AssertThat(t, vars.GetVariable("hello"), is.EqualTo("world"))
	then.AssertThat(t, vars.GetVariable("john"), is.EqualTo("doe"))
	then.AssertThat(t, vars.GetVariable("valueFromHandler"), is.EqualTo(true))
}

func Test_MarshallEngineComplexVars(t *testing.T) {
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-with_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		s := job.Variable("testStruct").(TestStruct)
		p := job.Variable("testPointer").(*TestStruct)
		fmt.Printf("Name: %s, Age: %d\n", s.Name, s.Age)
		fmt.Printf("Name: %s, Age: %d\n", p.Name, p.Age)
		job.Complete()
	})
	variableContext := make(map[string]interface{})
	variableContext["testStruct"] = TestStruct{
		Name: "struct",
		Age:  12,
	}
	variableContext["testPointer"] = &TestStruct{
		Name: "pointer",
		Age:  23,
	}
	variableContext["testSlicePtr"] = &[]TestStruct{{
		Name: "slice",
		Age:  1,
	}}

	_, _ = bpmnEngine.CreateAndRunInstance(process.ProcessKey, variableContext)

	data, err := bpmnEngine.Marshal(WithMarshalComplexTypes())
	then.AssertThat(t, err, is.Nil())
	fmt.Println(string(data))

	bpmnEngine, err = Unmarshal(data,
		WithUnmarshalComplexTypes(
			RegisterType(TestStruct{}),
		),
	)
	then.AssertThat(t, err, is.Empty())
	vars := bpmnEngine.ProcessInstances()[0].VariableHolder
	then.AssertThat(t, vars.GetVariable("testStruct"), is.EqualTo(TestStruct{
		Name: "struct",
		Age:  12,
	}))
	then.AssertThat(t, vars.GetVariable("testPointer"), is.EqualTo(&TestStruct{
		Name: "pointer",
		Age:  23,
	}))
	then.AssertThat(t, vars.GetVariable("testSlicePtr"), is.EqualTo(&[]TestStruct{{
		Name: "slice",
		Age:  1,
	}}))
}

func Test_MarshallEngineComplexVarsSlicePtr(t *testing.T) {
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-with_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		job.Complete()
	})
	variableContext := make(map[string]interface{})
	variableContext["testSlicePtr"] = &[]TestStruct{{
		Name: "slice",
		Age:  1,
	}}

	_, _ = bpmnEngine.CreateAndRunInstance(process.ProcessKey, variableContext)

	data, err := bpmnEngine.Marshal(WithMarshalComplexTypes())
	then.AssertThat(t, err, is.Nil())
	fmt.Println(string(data))

	bpmnEngine, err = Unmarshal(data,
		WithUnmarshalComplexTypes(
			RegisterType(TestStruct{}),
		),
	)
	then.AssertThat(t, err, is.Empty())
	vars := bpmnEngine.ProcessInstances()[0].VariableHolder
	then.AssertThat(t, vars.GetVariable("testSlicePtr"), is.EqualTo(&[]TestStruct{{
		Name: "slice",
		Age:  1,
	}}))
}

func Test_MarshallEngineComplexVarsSlice(t *testing.T) {
	bpmnEngine := New()
	process, _ := bpmnEngine.LoadFromFile("../../test-cases/simple_task-with_output_mapping.bpmn")
	bpmnEngine.NewTaskHandler().Id("id").Handler(func(job ActivatedJob) {
		job.Complete()
	})
	variableContext := make(map[string]interface{})
	variableContext["testSlice"] = []TestStruct{{
		Name: "slice",
		Age:  1,
	}}

	_, _ = bpmnEngine.CreateAndRunInstance(process.ProcessKey, variableContext)

	data, err := bpmnEngine.Marshal(WithMarshalComplexTypes())
	then.AssertThat(t, err, is.Nil())
	fmt.Println(string(data))

	bpmnEngine, err = Unmarshal(data,
		WithUnmarshalComplexTypes(
			RegisterType(TestStruct{}),
		),
	)
	then.AssertThat(t, err, is.Empty())
	vars := bpmnEngine.ProcessInstances()[0].VariableHolder
	then.AssertThat(t, vars.GetVariable("testSlice"), is.EqualTo([]TestStruct{{
		Name: "slice",
		Age:  1,
	}}))
}

func Test_applyUnmarshalOptions(t *testing.T) {
	type args struct {
		options []UnmarshalOption
	}
	tests := []struct {
		name    string
		args    args
		want    *unmarshalOptions
		wantErr bool
	}{
		{
			name: "no options",
			args: args{
				options: []UnmarshalOption{},
			},
			want: &unmarshalOptions{
				exportTypes: false,
				typeMapping: make(map[string]reflect.Type),
			},
			wantErr: false,
		},
		{
			name: "With unmarshal complex types",
			args: args{
				options: []UnmarshalOption{
					WithUnmarshalComplexTypes(),
				},
			},
			want: &unmarshalOptions{
				exportTypes: true,
				typeMapping: map[string]reflect.Type{
					"github.com/pbinitiative/feel.FEELDate":     reflect.TypeOf(feel.FEELDate{}),
					"github.com/pbinitiative/feel.FEELDatetime": reflect.TypeOf(feel.FEELDatetime{}),
					"github.com/pbinitiative/feel.FEELDuration": reflect.TypeOf(feel.FEELDuration{}),
					"github.com/pbinitiative/feel.FEELTime":     reflect.TypeOf(feel.FEELTime{}),
					"github.com/pbinitiative/feel.NullValue":    reflect.TypeOf(feel.NullValue{}),
					"github.com/pbinitiative/feel.Number":       reflect.TypeOf(feel.Number{}),
				},
			},
			wantErr: false,
		},
		{
			name: "With unmarshal complex types and added type",
			args: args{
				options: []UnmarshalOption{
					WithUnmarshalComplexTypes(
						RegisterType(TestStruct{}),
					),
				},
			},
			want: &unmarshalOptions{
				exportTypes: true,
				typeMapping: map[string]reflect.Type{
					"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine.TestStruct": reflect.TypeOf(TestStruct{}),
					"github.com/pbinitiative/feel.FEELDate":                           reflect.TypeOf(feel.FEELDate{}),
					"github.com/pbinitiative/feel.FEELDatetime":                       reflect.TypeOf(feel.FEELDatetime{}),
					"github.com/pbinitiative/feel.FEELDuration":                       reflect.TypeOf(feel.FEELDuration{}),
					"github.com/pbinitiative/feel.FEELTime":                           reflect.TypeOf(feel.FEELTime{}),
					"github.com/pbinitiative/feel.NullValue":                          reflect.TypeOf(feel.NullValue{}),
					"github.com/pbinitiative/feel.Number":                             reflect.TypeOf(feel.Number{}),
				},
			},
			wantErr: false,
		},
		{
			name: "With unmarshal complex types giving error",
			args: args{
				options: []UnmarshalOption{
					WithUnmarshalComplexTypes(
						RegisterType(&TestStruct{}), // Pointer not allowed
					),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyUnmarshalOptions(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyUnmarshalOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyUnmarshalOptions() got = %v, want %v", got, tt.want)
			}
		})
	}
}
