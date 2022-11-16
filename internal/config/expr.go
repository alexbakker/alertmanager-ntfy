package config

import (
	"fmt"
	"reflect"

	"github.com/PaesslerAG/gval"
)

var (
	exprLang = gval.NewLanguage(
		gval.Arithmetic(),
		gval.Text(),
		gval.PropositionalLogic(),
		gval.JSON(),
		gval.InfixOperator("in", func(a, b interface{}) (interface{}, error) {
			col, ok := b.([]interface{})
			if !ok {
				return nil, fmt.Errorf("expected type []interface{} for in operator but got %T", b)
			}
			for _, value := range col {
				if reflect.DeepEqual(a, value) {
					return true, nil
				}
			}
			return false, nil
		}),
	)
)
