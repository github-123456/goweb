package goweb

import "github.com/microcosm-cc/bluemonday"

func init()  {
	bluemondayPolicy=bluemonday.NewPolicy()
}

var bluemondayPolicy *bluemonday.Policy
