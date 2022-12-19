package var_holder

type VariableHolder struct {
	parent    *VariableHolder
	variables map[string]interface{}
}

// New creates a new VariableHolder with a given parent and variables map.
// All variables from parent holder are copied into this one.
// (hint: the copy is necessary, due to the fact we need to have all variables for expression evaluation in one map)
func New(parent *VariableHolder, variables map[string]interface{}) VariableHolder {
	if variables == nil {
		variables = make(map[string]interface{})
	}
	if parent != nil {
		for k, v := range parent.variables {
			variables[k] = v
		}
	}
	return VariableHolder{
		parent:    parent,
		variables: variables,
	}
}

func (vh *VariableHolder) GetVariable(key string) interface{} {
	if v, ok := vh.variables[key]; ok {
		return v
	}
	return nil
}

func (vh *VariableHolder) SetVariable(key string, val interface{}) {
	vh.variables[key] = val
}

// PropagateVariable set a value with given key to the parent VariableHolder
func (vh *VariableHolder) PropagateVariable(key string, value interface{}) {
	if vh.parent != nil {
		vh.parent.SetVariable(key, value)
	}
}

// Variables return all variables within this holder
func (vh *VariableHolder) Variables() map[string]interface{} {
	return vh.variables
}
