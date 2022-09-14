package bpmn_engine

type variableScope struct {
	Parent   *variableScope
	Children []*variableScope
	Context map[string]interface{}
}

func NewScope(parent *variableScope, context map[string]interface{}) *variableScope {
	if context == nil {
		return &variableScope{
			Context: make(map[string]interface{}),
			Parent: parent,
		}
	}
	return &variableScope{
		Context: context,
		Parent: parent,
	}
}

func (s *variableScope) GetVariable(key string) interface{} {
	cur := s
	for cur != nil {
		if v, ok := cur.Context[key]; ok {
			return v
		}
		cur = cur.Parent
	}
	return nil
}

func (s *variableScope)GetContext() map[string]interface{} {
	return s.Context
}

func (s *variableScope)SetVariable(key string, val interface{}) {
	s.Context[key] = val
}