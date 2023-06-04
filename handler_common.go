package rabbit

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type onRenderFunc[T any] func(ctx *gin.Context, v *T) error
type onDeleteFunc[T any] func(ctx *gin.Context, v *T) error
type onCreateFunc[T any] func(ctx *gin.Context, v *T) error
type onUpdateFunc[T any] func(ctx *gin.Context, v *T, vals map[string]any) error

func HandleGet[T any](c *gin.Context, db *gorm.DB, onRender onRenderFunc[T]) {
	key := c.Param("key")

	pkName := GetPkColumnName[T]()
	val := new(T)

	result := db.Where(pkName, key).First(val)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithError(http.StatusNotFound, result.Error)
		} else {
			c.AbortWithError(http.StatusInternalServerError, result.Error)
		}
		return
	}

	if onRender != nil {
		if err := onRender(c, val); err != nil {
			c.AbortWithError(http.StatusNotFound, result.Error)
			return
		}
	}

	c.JSON(http.StatusOK, val)
}

func HandleDelete[T any](c *gin.Context, db *gorm.DB, onDelete onDeleteFunc[T]) {
	key := c.Param("key")

	pkName := GetPkColumnName[T]()
	val := new(T)

	// form gorm delete hook, need to load model first
	result := db.Where(pkName, key).First(val)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, true)
		} else {
			c.AbortWithError(http.StatusInternalServerError, result.Error)
		}
		return
	}

	if onDelete != nil {
		if err := onDelete(c, val); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	result = db.Delete(val)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, true)
}

func HandleCreate[T any](c *gin.Context, db *gorm.DB, onCreate onCreateFunc[T]) {
	val := new(T)

	if err := c.BindJSON(val); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if onCreate != nil {
		if err := onCreate(c, val); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	result := db.Create(val)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, val)
}

func HandleEdit[T any](c *gin.Context, db *gorm.DB, editables []string, onUpdate onUpdateFunc[T]) {
	key := c.Param("key")

	var formVals map[string]any
	if err := c.BindJSON(&formVals); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rt := reflect.TypeOf(new(T)).Elem()

	// cannot edit primarykey
	delete(formVals, getPkJsonName(rt))

	var vals map[string]any = map[string]any{}
	for k, v := range formVals {
		field, ok := getFieldByJsonTag(rt, k)
		if !ok {
			continue
		}

		// check type
		kind := field.Type.Kind()
		if v == nil && kind != reflect.Ptr {
			continue
		}
		if v != nil && !checkType(kind, reflect.TypeOf(v).Kind()) {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("%s type not match", field.Name))
			return
		}

		vals[field.Name] = v
	}

	if len(editables) > 0 {
		stripVals := make(map[string]any)
		for _, k := range editables {
			if v, ok := vals[k]; ok {
				// TODO:
				// columnName, _ := getColumnNameByField(rt, k)
				// stripVals[columnName] = v
				stripVals[k] = v
			}
		}
		vals = stripVals
	} else {
		vals = map[string]any{}
	}

	if len(vals) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("not changed"))
		return
	}

	pkColumnName := GetPkColumnName[T]()

	if onUpdate != nil {
		val := new(T)
		if err := db.First(val, pkColumnName, key).Error; err != nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if err := onUpdate(c, val, formVals); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	var model = new(T)
	result := db.Model(model).Where(pkColumnName, key).Updates(vals)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, true)
}

// TODO:
func HandleQuery[T any](c *gin.Context, db *gorm.DB) {
	var form QueryForm
	if err := c.BindQuery(&form); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if form.Page < 1 {
		form.Page = 1
	}
	if form.Limit <= 0 || form.Limit > 150 {
		form.Limit = DefaultQueryLimit
	}

	// TODO: add filterables
	// TODO: add orderables
	// TODO: add searchable

	qr, err := ExecuteQuery[T](db, form)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, qr)
}

func ExecuteQuery[T any](db *gorm.DB, form QueryForm) (r QueryResult[[]T], err error) {
	// tableName := GetTableName[T](db)

	// TODO: form.Filters
	// TODO: form.Orders
	// TODO: form.Keyword

	r.Page = form.Page
	r.Limit = form.Limit
	r.Keyword = form.Keyword
	r.Items = make([]T, 0, form.Limit)

	var c int64
	result := db.Model(new(T)).Count(&c)
	if result.Error != nil {
		return r, result.Error
	}

	if c == 0 {
		return r, nil
	}

	r.TotalCount = int(c)
	result = db.Limit(form.Limit).Offset((form.Page - 1) * form.Limit).Find(&r.Items)
	if result.Error != nil {
		return r, result.Error
	}

	return r, nil
}
