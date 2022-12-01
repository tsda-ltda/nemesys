package tools

import (
	"errors"
	"strconv"

	"github.com/fernandotsda/nemesys/api-manager/internal/api"
	"github.com/gin-gonic/gin"
)

var (
	ErrFailToGetCustomQueryOnCache    = errors.New("fail to get custom query on cache")
	ErrFailToSetCustomQueryOnCache    = errors.New("fail to set custom query on cache")
	ErrFailToGetCustomQueryOnDatabase = errors.New("fail to get custom query on database")
	ErrCustomQueryNotFound            = errors.New("custom query does not exists")
)

// GetCustomQueryFlux get the custom query id/ident on gin context query. Try to get
// the flux on cache, if cache is missing goes to database and save on cache after.
func GetCustomQueryFlux(api *api.API, c *gin.Context) (flux string, err error) {
	ctx := c.Request.Context()
	rawCustomQuery := c.Query("custom_query")
	if len(rawCustomQuery) != 0 {
		id, err := strconv.ParseInt(rawCustomQuery, 0, 32)
		if err != nil {
			cacheRes, err := api.Cache.GetCustomQueryByIdent(ctx, rawCustomQuery)
			if err != nil {
				return flux, ErrFailToGetCustomQueryOnCache
			}
			if !cacheRes.Exists {
				dbRes, err := api.PG.GetCustomQueryByIdent(ctx, rawCustomQuery)
				if err != nil {
					return flux, ErrFailToGetCustomQueryOnDatabase
				}
				if !dbRes.Exists {
					return flux, ErrCustomQueryNotFound
				}
				flux = dbRes.CustomQuery.Flux
				err = api.Cache.SetCustomQueryByIdent(ctx, dbRes.CustomQuery.Flux, rawCustomQuery)
				if err != nil {
					return flux, ErrFailToSetCustomQueryOnCache
				}
			} else {
				flux = cacheRes.Flux
			}
		} else {
			cacheRes, err := api.Cache.GetCustomQuery(ctx, int32(id))
			if err != nil {
				return flux, ErrFailToGetCustomQueryOnCache
			}

			if !cacheRes.Exists {
				dbRes, err := api.PG.GetCustomQuery(ctx, int32(id))
				if err != nil {
					return flux, ErrFailToGetCustomQueryOnDatabase
				}
				if !dbRes.Exists {
					return flux, ErrCustomQueryNotFound
				}
				flux = dbRes.CustomQuery.Flux
				err = api.Cache.SetCustomQuery(ctx, dbRes.CustomQuery.Flux, int32(id))
				if err != nil {
					return flux, ErrFailToSetCustomQueryOnCache
				}
			} else {
				flux = cacheRes.Flux
			}
		}
	}
	return flux, nil
}
