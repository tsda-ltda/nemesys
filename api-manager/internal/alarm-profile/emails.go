package profile

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Add an email to alarm profile.
// Responses:
//   - 400 If invalid params.
//   - 404 If alarm profile not found.
//   - 200 If succeeded.
func CreateEmailHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("profileId")
		id, err := strconv.ParseInt(rawId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var email models.AlarmProfileEmail
		err = c.ShouldBind(&email)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(email)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		exists, err := api.PG.AlarmProfileExists(ctx, int32(id))
		if err != nil {
			api.Log.Error("Fail to check if alarm profile exists", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmProfileNotFound))
			return
		}

		err = api.PG.CreateAlarmProfileEmail(ctx, int32(id), email.Email)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to add email to alarm profile", logger.ErrField(err))
			return
		}
		api.Log.Debug("Email added to alarm profile, profile id: " + rawId)

		c.Status(http.StatusOK)
	}
}

// Remove an email from alarm profile.
// Responses:
//   - 400 If invalid params.
//   - 404 If relation not found.
//   - 204 If succeeded.
func DeleteEmailHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawEmailId := c.Param("emailId")
		emailId, err := strconv.ParseInt(rawEmailId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.DeleteAlarmProfileEmails(ctx, int32(emailId))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to remove email from alarm profile", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgAlarmProfileEmailNotFound))
			return
		}
		api.Log.Debug("Email removed from alarm profile, id: " + rawEmailId)

		c.Status(http.StatusNoContent)
	}
}

// Get alarm profile emails.
// Params:
//   - "limit" Limit of containers returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid params.
//   - 200 If succeeded.
func GetEmailsHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawId := c.Param("profileId")
		id, err := strconv.ParseInt(rawId, 0, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

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

		emails, err := api.PG.GetAlarmProfileEmails(ctx, int32(id), limit, offset)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get emails from alarm profile", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, emails)
	}
}
