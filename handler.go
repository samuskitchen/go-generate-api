package generate

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/samuskitchen/go-check-tool"
	"github.com/samuskitchen/go-message-tool"
	"github.com/samuskitchen/go-message-tool/errors"
	"gorm.io/gorm"
)

const (
	ActionCreate  = "CREATE"
	ActionDelete  = "DELETE"
	ActionUpdate  = "UPDATE"
	ActionFindAll = "FIND_ALL"
	ActionFindBy  = "FIND_BY_IDENTIFIER"
	ActionPatch   = "PATCH"
)

type isString bool

type HandlerGenerate struct {
	fieldKey         FieldName
	allowActions     map[string]struct{}
	translateFields  map[string]string
	filterableFields map[string]isString
	//ignoreFields  map[string]struct{}
	group   *echo.Group
	storage *storage
}

func NewHandlerMan(g *echo.Group, conn *gorm.DB) *HandlerGenerate {
	return &HandlerGenerate{
		fieldKey: FieldName{
			TableFieldName: "id",
			ModelFieldName: "ID",
			IsNumber:       true,
		},
		filterableFields: make(map[string]isString),
		allowActions: map[string]struct{}{
			ActionCreate: {}, ActionDelete: {}, ActionUpdate: {},
			ActionFindAll: {}, ActionFindBy: {}, ActionPatch: {}},
		//ignoreFields: make(map[string]struct{}),
		group: g,
		storage: &storage{
			conn: conn,
		},
	}
}

func (hg *HandlerGenerate) Start(i interface{}, options ...options) error {
	rType := reflect.TypeOf(i)
	if rType.Kind() == reflect.Ptr {
		if rType.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("error, a structure was expected")
		}
	} else if rType.Kind() != reflect.Struct {
		return fmt.Errorf("error, a structure was expected")
	}

	hg.storage.rType = rType
	hg.translateFields = getMapJsonFieldNameWithModelFieldName(i)
	for _, op := range options {
		op.apply(hg)
	}

	hg.router()

	return nil
}
func (hg *HandlerGenerate) router() {
	if _, ok := hg.allowActions[ActionFindAll]; ok {
		hg.group.GET("", hg.findAll)
	}

	if _, ok := hg.allowActions[ActionFindBy]; ok {
		hg.group.GET(fmt.Sprintf("/:%s", hg.fieldKey.TableFieldName), hg.findByIdentifier)
	}

	if _, ok := hg.allowActions[ActionCreate]; ok {
		hg.group.POST("", hg.create)
	}

	if _, ok := hg.allowActions[ActionUpdate]; ok {
		hg.group.PUT("", hg.update)
	}

	if _, ok := hg.allowActions[ActionDelete]; ok {
		hg.group.DELETE(fmt.Sprintf("/:%s", hg.fieldKey.TableFieldName), hg.delete)
	}
}

func (hg *HandlerGenerate) findAll(c echo.Context) error {
	filterValue := c.QueryParam("filter")
	if filterValue != "" {
		splits := strings.Split(filterValue, ",")
		if len(splits) == 2 {
			nameField, ok := hg.translateFields[splits[0]]
			var value interface{}
			if !ok {
				return message_tool.ErrorResponse(c, errores.NewBadRequest(nil, "%s invalid", splits[0]))
			}

			fl, ok := hg.filterableFields[nameField]
			if !ok {
				return message_tool.ErrorResponse(c, errores.NewBadRequest(
					fmt.Errorf("findAll: filed %s invalid", nameField), "%s invalid", splits[0]),
				)
			}

			if fl {
				value = splits[1]
			} else {
				num, err := strconv.Atoi(splits[1])
				if err != nil {
					return message_tool.ErrorResponse(c, errores.NewBadRequest(nil, "%s invalid, must be numeric", splits[0]))
				}

				value = num
			}

			data, err := hg.storage.findAllEntitiesWithFilter(filter{fieldName: nameField, value: value})
			if err != nil {
				return message_tool.ErrorResponse(c, err)
			}

			return message_tool.Success(c, data)
		}
	}

	data, err := hg.storage.findAllEntities()
	if err != nil {
		return message_tool.ErrorResponse(c, err)
	}

	return message_tool.Success(c, data)
}

func (hg *HandlerGenerate) findByIdentifier(c echo.Context) error {
	identifier := c.Param(hg.fieldKey.TableFieldName)
	var key interface{} = identifier
	var err error

	if hg.fieldKey.IsNumber {
		key, err = strconv.Atoi(identifier)
		if err != nil {
			return message_tool.ErrorResponse(c, errores.NewNotFound(nil, errores.ErrRecordNotFound))
		}
	}

	data, err := hg.storage.findByIdentifier(hg.fieldKey.TableFieldName, key)
	if err != nil {
		return message_tool.ErrorResponse(c, err)
	}

	return message_tool.Success(c, data)
}

func (hg *HandlerGenerate) create(c echo.Context) error {
	newObjet := reflect.New(hg.storage.rType).Interface()
	if err := jsonBind(c, newObjet); err != nil {
		return message_tool.JSONErrorResponse(c)
	}

	if err := check_tool.Valid(newObjet); err != nil {
		return message_tool.ErrorResponse(c, errores.NewBadRequest(nil, err.Error()))
	}

	if err := hg.storage.create(newObjet); err != nil {
		return message_tool.ErrorResponse(c, err)
	}

	return message_tool.Message(c, message_tool.OperationSuccess)
}

func (hg *HandlerGenerate) update(c echo.Context) error {
	newObjet := reflect.New(hg.storage.rType).Interface()
	if err := jsonBind(c, newObjet); err != nil {
		return message_tool.JSONErrorResponse(c)
	}

	field := reflect.ValueOf(newObjet).Elem().FieldByName(hg.fieldKey.ModelFieldName)
	if !field.IsValid() {
		err := fmt.Errorf("invalid %s not found in structure %v", hg.fieldKey.ModelFieldName, hg.storage.rType)
		return message_tool.ErrorResponse(c, errores.NewInternal(err, errores.ErrDatabaseInternal))
	}

	if field.IsZero() {
		err := fmt.Errorf("%s not found in structure %v", hg.fieldKey.ModelFieldName, hg.storage.rType)
		return message_tool.ErrorResponse(c, errores.NewInternal(err, errores.ErrDatabaseInternal))
	}

	pkValue, err := getIdentifierValues(field)
	if err != nil {
		return message_tool.ErrorResponse(c, err)
	}

	if err = check_tool.Valid(newObjet); err != nil {
		return message_tool.ErrorResponse(c, errores.NewBadRequest(nil, err.Error()))
	}

	if err = hg.storage.update(hg.fieldKey.TableFieldName, pkValue, newObjet); err != nil {
		return message_tool.ErrorResponse(c, err)
	}

	return message_tool.Message(c, message_tool.OperationSuccess)
}

func (hg *HandlerGenerate) delete(c echo.Context) error {
	identifier := c.Param(hg.fieldKey.TableFieldName)
	var key interface{} = identifier
	var err error

	if hg.fieldKey.IsNumber {
		key, err = strconv.Atoi(identifier)
		if err != nil {
			return message_tool.ErrorResponse(c, errores.NewNotFound(nil, errores.ErrRecordNotFound))
		}
	}

	if err = hg.storage.delete(hg.fieldKey.TableFieldName, key); err != nil {
		return message_tool.ErrorResponse(c, err)
	}

	return message_tool.Message(c, message_tool.OperationSuccess)
}
