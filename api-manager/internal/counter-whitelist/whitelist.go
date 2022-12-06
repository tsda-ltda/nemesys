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

		var userId models.UserId
		err := c.ShouldBind(&userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		exists, err := api.PG.UserExists(ctx, userId.UserId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to check if user exists", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}

		err = api.PG.AddUserToCounterWhitelist(ctx, userId.UserId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to add user to counter whitelist", logger.ErrField(err))
			return
		}
		api.Counter.LoadWhitelist()
		api.Log.Debug("User added to counter whitelist, user id: " + strconv.FormatInt(int64(userId.UserId), 10))

		c.Status(http.StatusOK)
	}
}

func DeleteHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("userId")
		id, err := strconv.ParseInt(rawId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
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
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserWhitelistNotFound))
			return
		}
		api.Counter.LoadWhitelist()
		api.Log.Debug("User removed from counter whitelist, user id: " + rawId)

		c.Status(http.StatusOK)
	}
}

func GetHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}
		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
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
		c.JSON(http.StatusOK, ids)
	}
}
