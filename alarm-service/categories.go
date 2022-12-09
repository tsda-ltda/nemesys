package alarm

import (
	"context"

	"github.com/fernandotsda/nemesys/shared/models"
)

func (a *Alarm) getCategoriesSimplified(ctx context.Context, ids []int32) (categories []models.AlarmCategorySimplified, err error) {
	categories, err = a.cache.GetAlarmCategoriesSimplified(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(categories) == len(ids) {
		return categories, nil
	}

	remainingIds := make([]int32, len(ids)-len(categories))
	for _, id := range ids {
		var found bool
		for _, c := range categories {
			if c.Id == id {
				found = true
				break
			}
		}
		if found {
			continue
		}
		remainingIds = append(remainingIds, id)
	}

	remainingCategories, err := a.pg.GetAlarmCategoriesSimplifiedByIds(ctx, remainingIds)
	if err != nil {
		return nil, err
	}
	return append(categories, remainingCategories...), a.cache.SetAlarmCategoriesSimplified(ctx, remainingCategories)
}
