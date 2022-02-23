package generate

import (
	"fmt"
	"reflect"

	"github.com/samuskitchen/go-message-tool/errors"
	"gorm.io/gorm"
)

type filter struct {
	fieldName string
	value     interface{}
}

type storage struct {
	rType reflect.Type
	conn  *gorm.DB
}

func (s *storage) findAllEntitiesWithFilter(flt filter) (interface{}, error) {
	container := reflect.MakeSlice(reflect.SliceOf(s.rType), 0, 0).Interface()
	if err := s.conn.Find(&container, fmt.Sprintf("%s = ?", flt.fieldName), flt.value).Error; err != nil {
		return nil, errores.NewInternalDB(err)
	}

	return container, nil
}

func (s *storage) findAllEntities() (interface{}, error) {
	container := reflect.MakeSlice(reflect.SliceOf(s.rType), 0, 0).Interface()
	if err := s.conn.Find(&container).Error; err != nil {
		return nil, errores.NewInternalDB(err)
	}

	return container, nil
}

func (s *storage) findByIdentifier(fileName string, value interface{}) (interface{}, error) {
	newObjet := reflect.New(s.rType).Interface()
	tx := s.conn.Limit(1)
	rs := tx.Where(fmt.Sprintf("%s = ?", fileName), value).Find(&newObjet)

	if rs.Error != nil {
		return nil, errores.NewInternalDB(rs.Error)
	}

	if rs.RowsAffected == 0 {
		return nil, errores.NewNotFound(nil, errores.ErrRecordNotFound)
	}

	return newObjet, nil
}

func (s *storage) create(i interface{}) error {
	if err := s.conn.Create(i).Error; err != nil {
		return errores.NewInternalDB(err)
	}

	return nil
}

func (s *storage) update(pkName string, pkValue interface{}, i interface{}) error {
	if err := s.conn.Where(fmt.Sprintf("%s = ?", pkName), pkValue).Updates(i).Error; err != nil {
		return errores.NewInternalDB(err)
	}

	return nil
}

func (s *storage) delete(pkName string, pkValue interface{}) error {
	rs := s.conn.Delete(reflect.New(s.rType).Interface(), fmt.Sprintf("%s = ?", pkName), pkValue)
	if rs.Error != nil {
		return errores.NewInternalDB(rs.Error)
	}

	if rs.RowsAffected == 0 {
		return errores.NewBadRequest(nil, errores.ErrRecordNotFound)
	}

	return nil
}
