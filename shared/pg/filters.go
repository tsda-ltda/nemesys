package pg

import (
	"fmt"
	"reflect"
	"strings"
)

type queryFilters2 interface {
	GetOrderBy() string
	GetOrderByFn() string
	GetLimit() int
	GetOffset() int
}

func mergeFilters(filters []string) string {
	if len(filters) == 0 {
		return ""
	}
	e := ` WHERE `
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

// applyFilters apply the queryFilters in the provided sql. The queryFilters may NOT have any
// unexported field, otherwise will panic.
func applyFilters(queryFilters queryFilters2, sql string, allowedColumns []string) (sqlResult string, params []any, err error) {
	v := reflect.ValueOf(queryFilters)
	typeof := reflect.TypeOf(queryFilters)

	params = make([]any, 0, v.NumField())
	statements := make([]string, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		_i := []int{i}
		field := typeof.FieldByIndex(_i)
		value := v.FieldByIndex(_i)

		fieldValue, ok := getFilterFieldValue(value)
		if !ok {
			continue
		}

		operator := field.Tag.Get("type")
		if len(operator) == 0 {
			continue
		}

		index := len(statements) + 1
		column := field.Tag.Get("column")

		statements = append(statements, getStatement(column, operator, index))

		fieldValueS, ok := fieldValue.(string)
		if operator == "ilike" && ok {
			fieldValue = fieldValueS + "%"
		}
		params = append(params, fieldValue)
	}
	statementMerged := mergeFilters(statements)

	orderBy := queryFilters.GetOrderBy()
	orderByFn := strings.ToUpper(queryFilters.GetOrderByFn())
	if orderBy != "" && orderByFn != "" {
		if err := validateOrderFn(orderByFn); err != nil {
			return "", nil, err
		}

		var founded bool
		for _, col := range allowedColumns {
			if col == orderBy {
				founded = true
			}
		}
		if !founded {
			return "", nil, ErrInvalidFilterValue
		}

		statementMerged += fmt.Sprintf(` ORDER BY %s %s`, orderBy, orderByFn)
	}

	statementMerged += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(statements)+1, len(statements)+2)
	params = append(params, queryFilters.GetLimit(), queryFilters.GetOffset())

	return sql + statementMerged, params, nil
}

func getStatement(column string, operator string, index int) string {
	return fmt.Sprintf("%s %s $%d", column, operator, index)
}

func getFilterFieldValue(v reflect.Value) (parsedValue any, empty bool) {
	if v.IsZero() {
		return nil, false
	}

	if reflect.Pointer == v.Kind() {
		if v.Type() == reflect.PointerTo(reflect.TypeOf(false)) {
			if *v.Interface().(*bool) {
				return "true", true
			}
			return "false", true
		}
	}

	return v.Interface(), true
}
