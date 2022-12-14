package middleware

import (
	"net/http"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/api-manager/internal/tools"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/gin-gonic/gin"
)

// TeamGuard allow authenticated users with a role bigger or equal to the accessLevel
// and members of team pass. If user role is bigger ot equal to the freepassLevel will
// let pass even if not member.
// Responses:
//   - 400 If invalid id
//   - 403 If invalid role
//   - 403 If not member
func TeamGuard(api *api.API, accessLevel roles.Role, freepassLevel roles.Role) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		teamId, err := strconv.ParseInt(c.Param("teamId"), 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, tools.MsgRes(tools.MsgInvalidParams))
			return
		}

		meta, err := tools.GetSessionMeta(c)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if meta.Role >= freepassLevel {
			c.Next()
			return
		} else if meta.Role < accessLevel {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		exists, err := api.PG.TeamMemberExists(ctx, int32(teamId), meta.UserId)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.AbortWithStatus(http.StatusInternalServerError)
			api.Log.Error("Fail to check if user is a member of a team", logger.ErrField(err))
			return
		}

		if !exists {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
