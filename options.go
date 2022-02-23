package generate

type options interface {
	apply(*HandlerGenerate)
}

type fieldName struct {
	TableFieldName string
	ModelFieldName string
	IsNumber       bool
}

func WithKeyFieldName(tableFieldName, modelFieldName string, isNumber bool) fieldName {
	return fieldName{TableFieldName: tableFieldName, ModelFieldName: modelFieldName, IsNumber: isNumber}
}

func (fn *fieldName) apply(handlerGenerate *HandlerGenerate) {
	handlerGenerate.fieldKey.TableFieldName = fn.TableFieldName
	handlerGenerate.fieldKey.ModelFieldName = fn.ModelFieldName
	handlerGenerate.fieldKey.IsNumber = fn.IsNumber
}

type allowActions struct {
	actions []string
}

func WithAllowActions(actions ...string) allowActions {
	return allowActions{actions: actions}
}

func (a *allowActions) apply(handlerGenerate *HandlerGenerate) {
	for _, action := range a.actions {
		handlerGenerate.allowActions[action] = struct{}{}
	}
}

type FilterOption struct {
	FieldTableName string
	IsString       bool
}

type filterable struct {
	filter []FilterOption
}

func WithFieldFilter(tableFields ...FilterOption) filterable {
	return filterable{filter: tableFields}
}

func (f *filterable) apply(handlerGenerate *HandlerGenerate) {
	for _, fil := range f.filter {
		handlerGenerate.filterableFields[fil.FieldTableName] = isString(fil.IsString)
	}
}
