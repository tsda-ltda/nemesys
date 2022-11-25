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

		teamId, err := strconv.ParseInt(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		var userId models.AddMemberReq
		err = c.ShouldBind(&userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidJSONFields))
			return
		}

		r, err := api.PG.ExistsRelUserTeam(ctx, userId.UserId, int32(teamId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get realation, user and team existence", logger.ErrField(err))
			return
		}
		if r.RelationExist {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgRelationExists))
			return
		}
		if !r.UserExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgUserNotFound))
			return
		}
		if !r.TeamExists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgTeamNotFound))
			return
		}

		err = api.PG.AddTeamMember(ctx, userId.UserId, int32(teamId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
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

		teamId, err := strconv.ParseInt(rawTeamId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		userId, err := strconv.ParseInt(rawUserId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.JSONMSG(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.RemoveTeamMember(ctx, int32(userId), int32(teamId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to remove member", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.JSONMSG(tools.MsgMemberNotFound))
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

		id, err := strconv.ParseInt(c.Param("id"), 10, 32)
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

		m, err := api.PG.GetTeamMembers(ctx, int32(id), limit, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("fail to get team members", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, m)
	}
}
