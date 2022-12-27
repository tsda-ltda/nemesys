package pg

import (
	"fmt"
	"reflect"
	"strings"

	"unicode/utf8"

	"github.com/fernandotsda/nemesys/shared/types"
)

type queryFilters interface {
	ContainerType() types.ContainerType
	GetOrderBy() string
	GetOrderByFn() string
}

func mergeFilters(filters []string) string {
	if len(filters) == 0 {
		return ""
	}
	e := `WHERE `
	for i, f := range filters {
		e += f
		if i != len(filters)-1 {
			e += ` AND `
		}
	}
	return e
}

func validateOrderFn(fn string) error {
	if fn != "DESC" && fn != "ASC" {
		return ErrInvalidOrderByFn
	}
	return nil
}

func applyFilters(queryFilters queryFilters, sql string, allowedColumns []string) (sqlResult string, err error) {
	v := reflect.ValueOf(queryFilters)
	typeof := reflect.TypeOf(queryFilters)

	sqlFilters := make([]string, 0, v.NumField())

	// replacer for special postgres chars
	re := strings.NewReplacer("_", "/_", "%", "/%", "'", "''")

	for i := 0; i < v.NumField(); i++ {
		_i := []int{i}
		field := typeof.FieldByIndex(_i)
		value := v.FieldByIndex(_i)

		fieldValue, ok := getFilterFieldValue(value)
		if !ok {
			continue
		}

		// replace special chars if fieldValue is a string
		if reflect.TypeOf(fieldValue) == reflect.TypeOf("") {
			asString := fieldValue.(string)
			if !utf8.ValidString(asString) {
				return "", ErrInvalidFilterValue
			}
			fieldValue = re.Replace(asString)
		}

		switch field.Tag.Get("type") {
		case "=":
			sqlFilters = append(sqlFilters, fmt.Sprintf("%s = %v", field.Tag.Get("column"), fieldValue))
		case "<=":
			sqlFilters = append(sqlFilters, fmt.Sprintf("%s <= %v", field.Tag.Get("column"), fieldValue))
		case ">=":
			sqlFilters = append(sqlFilters, fmt.Sprintf("%s >= %v", field.Tag.Get("column"), fieldValue))
		case "ilike":
			sqlFilters = append(sqlFilters, fmt.Sprintf("%s ILIKE '%s", field.Tag.Get("column"), fieldValue)+"%'")
		}
	}

	sqlFilters = append(sqlFilters, fmt.Sprintf("type = %d", queryFilters.ContainerType()))
	frag := mergeFilters(sqlFilters)

	orderBy := queryFilters.GetOrderBy()
	orderByFn := queryFilters.GetOrderByFn()
	if orderBy != "" && orderByFn != "" {
		if err := validateOrderFn(orderByFn); err != nil {
			return "", err
		}

		var founded bool
		for _, col := range allowedColumns {
			if col == orderBy {
				founded = true
			}
		}
		if !founded {
			return "", ErrInvalidFilterValue
		}

		frag += fmt.Sprintf(` ORDER BY %s %s`, orderBy, orderByFn)
	}

	return fmt.Sprintf(sql, frag), nil
}

func getFilterFieldValue(v reflect.Value) (parsedValue any, empty bool) {
	if v.IsZero() {
		return nil, false
	}
	switch v.Kind() {
	case reflect.String:
		if v.String() != "" {
			return v.Interface().(string), true
		}
	case reflect.Pointer:
		if v.Type() == reflect.PointerTo(reflect.TypeOf(false)) {
			if *v.Interface().(*bool) {
				return "true", true
			}
			return "false", true
		}
	case reflect.Int:
		return v.Interface(), true
	case reflect.Int8:
		return v.Interface(), true
	case reflect.Int16:
		return v.Interface(), true
	case reflect.Int32:
		return v.Interface(), true
	case reflect.Int64:
		return v.Interface(), true
	}
	return nil, false
}