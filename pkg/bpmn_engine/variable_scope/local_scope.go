package variable_scope

type LocalScope struct {
	Context map[string]interface{}
}

func NewLocalScope(context map[string]interface{}) VarScope {
	if context == nil {
		return &LocalScope{
			Context: make(map[string]interface{}),
		}
	}
	return &LocalScope{
		Context: context,
	}
}

func (l *LocalScope) GetContext() map[string]interface{} {
	return l.Context
}

func (l *LocalScope) SetVariable(key string, val interface{}) {
	if l.Context == nil {
		l.Context = make(map[string]interface{})
	}
	l.Context[key] = val
}

func (l *LocalScope) GetVariable(key string) interface{} {
	if l.Context == nil {
		l.Context = make(map[string]interface{})
	}
	return l.Context[key]
}

func (l *LocalScope)GetParent() VarScope {
	return nil
}

func (l *LocalScope) Propagation()  {

}



