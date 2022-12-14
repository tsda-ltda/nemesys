package container

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/pg"
	"github.com/gin-gonic/gin"
)

// Get a Flex Legacy container.
// Responses:
//   - 404 If not found.
//   - 200 If succeeded.
func GetFlexLegacyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("containerId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, container, err := api.PG.GetFlexLegacyContainer(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get flex legacy container", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgContainerNotFound))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(container))
	}
}

// Get flex legacy containers.
// Responses:
//   - 200 If succeeded.
func GetFlexLegacyContainersHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}
		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		createdAtStart, _ := strconv.ParseInt(c.Query("createdAtStart"), 0, 64)
		createdAtStop, _ := strconv.ParseInt(c.Query("createdAtStop"), 0, 64)
		model, _ := strconv.ParseInt(c.Query("model"), 0, 16)

		var e bool
		var enabled *bool
		enabledQuery := c.Query("enabled")
		if enabledQuery == "1" {
			e = true
			enabled = &e
		} else if enabledQuery == "0" {
			e = false
			enabled = &e
		}

		filters := pg.FlexLegacyContainerQueryFilters{
			Name:           c.Query("name"),
			Descr:          c.Query("descr"),
			CreatedAtStart: createdAtStart,
			CreatedAtStop:  createdAtStop,
			Enabled:        enabled,
			OrderBy:        c.Query("order-by"),
			OrderByFn:      c.Query("order-by-fn"),
			Target:         c.Query("target"),
			SerialNumber:   c.Query("serial-number"),
			Model:          int16(model),
			City:           c.Query("city"),
			Region:         c.Query("region"),
			Country:        c.Query("country"),
			Limit:          limit,
			Offset:         offset,
		}

		containers, err := api.PG.GetFlexLegacyContainers(ctx, filters)
		if err != nil {
			if err == pg.ErrInvalidOrderByColumn || err == pg.ErrInvalidFilterValue || err == pg.ErrInvalidOrderByFn {
				c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
				return
			}

			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to get containers", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(containers))
	}
}
