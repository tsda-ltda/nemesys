package whitelist

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

func CreateHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var id32 models.Id32
		err := c.ShouldBind(&id32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(id32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		exists, err := api.PG.UserExists(ctx, id32.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if user exists", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgUserNotFound))
			return
		}

		err = api.PG.AddUserToCounterWhitelist(ctx, id32.Id)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to add user to counter whitelist", logger.ErrField(err))
			return
		}
		api.Counter.LoadWhitelist()
		api.Log.Debug("User added to counter whitelist, user id: " + strconv.FormatInt(int64(id32.Id), 10))

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("userId")
		id, err := strconv.ParseInt(rawId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.RemoveUserFromCounterWhitelist(ctx, int32(id))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to remove user from counter whitelist", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgUserWhitelistNotFound))
			return
		}
		api.Counter.LoadWhitelist()
		api.Log.Debug("User removed from counter whitelist, user id: " + rawId)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

func GetHandler(api *api.API) func(c *gin.Context) {
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

		ids, err := api.PG.GetCounterWhitelist(ctx, limit, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get counter whitelist", logger.ErrField(err))
			return
		}
		c.JSON(http.StatusOK, tools.DataRes(ids))
	}
}
