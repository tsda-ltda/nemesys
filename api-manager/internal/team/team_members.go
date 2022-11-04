package team

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/gin-gonic/gin"
)

// Add a member.
// Responses:
//   - 400 If invalid body.
//   - 400 If invalid user or team id.
//   - 400 If user is already a member.
//   - 404 If team does not exists.
//   - 200 If succeeded.
func AddMemberHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get team id
		teamId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get user id
		var userId models.AddMemberReq
		err = c.ShouldBind(&userId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// validate userid
		err = api.Validate.Struct(userId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get realation, user and team existence
		re, ue, te, err := api.PgConn.Teams.ExistsRelUserTeam(ctx, userId.UserId, teamId)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get realation, user and team existence", logger.ErrField(err))
			return
		}

		// check if realation already exists
		if re {
			c.Status(http.StatusBadRequest)
			return
		}

		// check if team or user doesn't exists
		if !te || !ue {
			c.Status(http.StatusNotFound)
			return
		}

		// add user to the team
		err = api.PgConn.Teams.AddMember(ctx, userId.UserId, int32(teamId))
		if err != nil {
			api.Log.Error("fail to add member to team", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	}
}

// Remove a member.
// Responses:
//   - 400 If invalid user or team id.
//   - 400 If user is already a member.
//   - 404 If relation does not exists.
//   - 204 If succeeded.
func RemoveMemberHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawTeamId := c.Param("id")
		rawUserId := c.Param("userId")

		// get team teamId
		teamId, err := strconv.Atoi(rawTeamId)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// get user id
		userId, err := strconv.Atoi(rawUserId)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// remove user from team
		e, err := api.PgConn.Teams.RemMember(ctx, int32(userId), int32(teamId))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to remove member", logger.ErrField(err))
			return
		}

		// check if relation exists
		if !e {
			c.Status(http.StatusNotFound)
			return
		}

		api.Log.Debug("user " + rawUserId + " removed from team " + rawTeamId)

		c.Status(http.StatusNoContent)
	}
}

// Remove a member.
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid user or team id.
//   - 204 If succeeded.
func MGetMembersHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// get team id
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// db query params
		limit, err := tools.IntRangeQuery(c, "limit", 30, 30, 1)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		offset, err := tools.IntMinQuery(c, "offset", 0, 0)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		// get members
		m, err := api.PgConn.Teams.MGetMembers(ctx, int32(id), limit, offset)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get team members", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, m)
	}
}