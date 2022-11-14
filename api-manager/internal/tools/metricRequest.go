package tools

import (
	"errors"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

var ErrMetricRequestNotExists = errors.New("metric request not exists on context")
var ErrFailToMakeMetricRequestAssertion = errors.New("fail to make metric request type assertion")

// GetMetricRequest returns the metric request saved on context. Returns an error
// if metric request does not exists or is invalid.
func GetMetricRequest(c *gin.Context) (r models.MetricRequest, err error) {
	v, e := c.Get("metric_request")
	if !e {
		return r, ErrMetricRequestNotExists
	}
	r, ok := v.(models.MetricRequest)
	if !ok {
		return r, ErrFailToMakeMetricRequestAssertion
	}
	return r, nil
}
