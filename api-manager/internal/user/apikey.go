package user

import (
	"net/http"
	"strconv"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/auth"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// Creates a API Key.
// Responses:
//   - 400 If invalid body.
//   - 400 If json fields are invalid.
//   - 404 If user not found.
//   - 200 If succeeded.
func CreateAPIKeyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawUserId := c.Param("userId")
		userId, err := strconv.ParseInt(rawUserId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var apikeyInfo models.APIKeyInfo
		err = c.ShouldBind(&apikeyInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(apikeyInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}

		exists, role, err := api.PG.GetUserRole(ctx, int32(userId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get user role on database", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}
		if userId != int64(meta.UserId) && role >= int16(meta.Role) {
			c.Status(http.StatusForbidden)
			return
		}

		apikeyInfo.UserId = int32(userId)
		apikeyInfo.CreatedAt = time.Now().Unix()
		id, tx, err := api.PG.CreateAPIKey(ctx, apikeyInfo)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create api key on database", logger.ErrField(err))
			return
		}

		var apikeyMeta auth.APIKeyMeta
		apikeyMeta.UserId = int32(userId)
		apikeyMeta.Role = uint8(role)
		apikeyMeta.Id = id
		apikey, err := api.Auth.NewAPIKey(ctx, apikeyMeta, time.Duration(apikeyInfo.TTL)*time.Hour)
		if err != nil {
			tx.Rollback()
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create API Key on Auth handler", logger.ErrField(err))
			return
		}

		err = tx.Commit()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to create commit tx", logger.ErrField(err))
			return
		}
		api.Log.Info("API Key created, id:" + strconv.FormatInt(int64(id), 10))

		c.JSON(http.StatusOK, models.APIkey{
			APIKey: apikey,
		})
	}
}

// Deletes a API Key.
// Responses:
//   - 404 If not found.
//   - 204 If succeeded.
func DeleteAPIKeyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userId, err := strconv.ParseInt(c.Param("userId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		rawId := c.Param("apikeyId")
		id, err := strconv.ParseInt(rawId, 0, 16)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}

		exists, role, err := api.PG.GetUserRole(ctx, int32(userId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get user role on database", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}
		if userId != int64(meta.UserId) && role >= int16(meta.Role) {
			c.Status(http.StatusForbidden)
			return
		}

		exists, tx, err := api.PG.DeleteAPIKey(ctx, int16(id), int32(userId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete api key on database", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAPIKeyNotFound))
			return
		}

		err = api.Auth.RemoveAPIKey(ctx, int32(id))
		if err != nil && err != redis.Nil {
			tx.Rollback()
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to remove api key from auth", logger.ErrField(err))
			return
		}
		err = tx.Commit()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to create commit tx", logger.ErrField(err))
			return
		}
		api.Log.Info("API Key deleted, id: " + rawId)

		c.Status(http.StatusNoContent)
	}

}

// Deletes a API Key.
// Responses:
//   - 404 If not found.
//   - 204 If succeeded.
func MGetAPIKeyHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		userId, err := strconv.ParseInt(c.Param("userId"), 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get session metadata", logger.ErrField(err))
			return
		}

		exists, role, err := api.PG.GetUserRole(ctx, int32(userId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get user role on database", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}

		if userId != int64(meta.UserId) && role >= int16(meta.Role) {
			c.Status(http.StatusForbidden)
			return
		}

		keys, err := api.PG.GetAPIKeys(ctx, int32(userId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get api keys on database", logger.ErrField(err))
			return
		}

		now := time.Now()
		keysAlived := make([]models.APIKeyInfo, 0, len(keys))
		keysRemoved := make([]int16, 0, len(keys))
		for _, k := range keys {
			if k.TTL > 0 && time.Unix(k.CreatedAt, 0).Add(time.Hour*time.Duration(k.TTL)).Sub(now) <= 0 {
				keysRemoved = append(keysRemoved, k.Id)
			} else {
				keysAlived = append(keysAlived, k)
			}
		}
		err = api.PG.DeleteAPIKeys(ctx, keysRemoved)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to delete api key on database", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, keysAlived)
	}
}
