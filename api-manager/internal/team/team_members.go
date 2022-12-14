package team

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/pg"
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

		rawTeamId := c.Param("teamId")
		teamId, err := strconv.ParseInt(rawTeamId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		var id models.Id32
		err = c.ShouldBind(&id)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidBody))
			return
		}

		err = api.Validate.Struct(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidJSONFields))
			return
		}

		r, err := api.PG.ExistsRelUserTeam(ctx, id.Id, int32(teamId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get realation, user and team existence", logger.ErrField(err))
			return
		}
		if r.RelationExist {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgRelationExists))
			return
		}
		if !r.UserExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgUserNotFound))
			return
		}
		if !r.TeamExists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgTeamNotFound))
			return
		}

		err = api.PG.AddTeamMember(ctx, id.Id, int32(teamId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			api.Log.Error("Fail to add member to team", logger.ErrField(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		api.Log.Info(fmt.Sprintf("User added to team, team id: %s, user id: %d", rawTeamId, id.Id))

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

// Remove a member.
// Responses:
//   - 400 If invalid user or team id.
//   - 400 If user is already a member.
//   - 404 If relation does not exists.
//   - 200 If succeeded.
func RemoveMemberHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		rawTeamId := c.Param("teamId")
		rawUserId := c.Param("userId")

		teamId, err := strconv.ParseInt(rawTeamId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		userId, err := strconv.ParseInt(rawUserId, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		exists, err := api.PG.RemoveTeamMember(ctx, int32(userId), int32(teamId))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to remove member", logger.ErrField(err))
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, tools.MsgRes(tools.MsgMemberNotFound))
			return
		}
		api.Log.Debug("User " + rawUserId + " removed from team " + rawTeamId)

		c.JSON(http.StatusOK, tools.EmptyRes())
	}
}

// Get team's members.
// Params:
//   - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
//   - "offset" Offset for searching. Default is 0, min is 0.
//
// Responses:
//   - 400 If invalid user or team id.
//   - 200 If succeeded.
func MGetMembersHandler(api *api.API) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		id, err := strconv.ParseInt(c.Param("teamId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}
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

		role, _ := strconv.ParseInt(c.Query("role"), 0, 16)
		m, err := api.PG.GetTeamMembers(ctx, pg.MemberQueryFilters{
			Limit:     limit,
			Offset:    offset,
			TeamId:    int32(id),
			Role:      int16(role),
			FirstName: c.Query("first-name"),
			LastName:  c.Query("last-name"),
			Username:  c.Query("username"),
			Email:     c.Query("email"),
			OrderBy:   c.Query("order-by"),
			OrderByFn: c.Query("order-by-fn"),
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.Status(http.StatusInternalServerError)
			api.Log.Error("Fail to get team members", logger.ErrField(err))
			return
		}

		c.JSON(http.StatusOK, tools.DataRes(m))
	}
}
