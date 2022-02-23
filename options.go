package generate

type options interface {
	apply(*HandlerGenerate)
}

type FieldName struct {
	TableFieldName string
	ModelFieldName string
	IsNumber       bool
}

type AllowActions struct {
	actions []string
}

type FilterOption struct {
	FieldTableName string
	IsString       bool
}

type Filterable struct {
	filter []FilterOption
}

func WithKeyFieldName(tableFieldName, modelFieldName string, isNumber bool) FieldName {
	return FieldName{TableFieldName: tableFieldName, ModelFieldName: modelFieldName, IsNumber: isNumber}
}

func (fn *FieldName) apply(handlerGenerate *HandlerGenerate) {
	handlerGenerate.fieldKey.TableFieldName = fn.TableFieldName
	handlerGenerate.fieldKey.ModelFieldName = fn.ModelFieldName
	handlerGenerate.fieldKey.IsNumber = fn.IsNumber
}

func WithAllowActions(actions ...string) AllowActions {
	return AllowActions{actions: actions}
}

func (a *AllowActions) apply(handlerGenerate *HandlerGenerate) {
	for _, action := range a.actions {
		handlerGenerate.allowActions[action] = struct{}{}
	}
}

func WithFieldFilter(tableFields ...FilterOption) Filterable {
	return Filterable{filter: tableFields}
}

func (f *Filterable) apply(handlerGenerate *HandlerGenerate) {
	for _, fil := range f.filter {
		handlerGenerate.filterableFields[fil.FieldTableName] = isString(fil.IsString)
	}
}
