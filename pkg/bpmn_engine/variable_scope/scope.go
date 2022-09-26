package variable_scope

type VariableContext interface {
	GetContext() map[string]interface{}
	GetParent() VariableContext
	SetVariable(key string, val interface{})
	GetVariable(key string) interface{}
}

type Scope struct {
	Parent   VariableContext
	Children []VariableContext
	Context  map[string]interface{}
}

func NewScope(parent VariableContext, context map[string]interface{}) VariableContext {
	if context == nil {
		return &Scope{
			Context: make(map[string]interface{}),
			Parent: parent,
		}
	}
	return &Scope{
		Context: context,
		Parent: parent,
	}
}

func MergeScope(local VariableContext, parent VariableContext) VariableContext {
	dst := parent.GetContext()
	for k, v := range local.GetContext() {
		dst[k] = v
	}
	return &Scope{
		Context: dst,
	}
}

func (s *Scope)GetContext() map[string]interface{} {
	var dst = make(map[string]interface{})
	for k, v := range s.Context {
		dst[k] = v
	}
	return dst
}

func (s *Scope) GetVariable(key string) interface{} {
	cur := s
	for cur != nil {
		if v, ok := cur.GetContext()[key]; ok {
			return v
		}
		if cur.GetParent() == nil {
			break
		}
		cur = cur.GetParent().(*Scope)
	}
	return nil
}

func (s *Scope)SetVariable(key string, val interface{}) {
	if s.Context == nil {
		s.Context = make(map[string]interface{})
	}
	s.Context[key] = val
}


func (s *Scope)GetParent() VariableContext {
	return s.Parent
}

