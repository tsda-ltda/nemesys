package influxdb

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/models"
	iapi "github.com/influxdata/influxdb-client-go/v2/api"
)

var aggrFns = []string{"mean", "median", "max", "min", "sum", "derivative", "nonnegative derivative", "distinct", "count", "increase", "skew", "spread", "stddev", "first", "last", "unique", "sort"}

// ValidateAggrFunction validates the aggregation function.
func ValidateAggrFunction(fn string) (valid bool) {
	for _, v := range aggrFns {
		if v == fn {
			return true
		}
	}
	return false
}

// getTaskName returns a task name for a data policy id.
func getTaskName(dataPolicyId int16) string {
	return fmt.Sprintf("%d-aggr-task", dataPolicyId)
}

// getTaskFlux returns the task flux.
func getTaskFlux(dp models.DataPolicy) string {
	return fmt.Sprintf(`
		option task = {name: "%s", every: %s, offset: %s}
		data = from(bucket: "%s")
			|> range(start: -%s, stop: -%s)
			|> filter(fn: (r) => r._measurement == "metrics")
		data
			|> aggregateWindow(every: %ds, fn: %s, createEmpty: false)
			|> to(bucket: "%s")`,
		getTaskName(dp.Id),
		fmt.Sprintf("%ds", dp.AggrInterval),
		fmt.Sprintf("%ds", dp.Retention*3600),
		GetBucketName(dp.Id, false),
		fmt.Sprintf("%ds", dp.Retention*3600),
		fmt.Sprintf("%ds", dp.Retention*3600-dp.AggrInterval),
		dp.AggrInterval,
		dp.AggrFn,
		GetBucketName(dp.Id, true),
	)
}

// createAggrTask creates an task of data aggregation.
func (c *Client) createAggrTask(ctx context.Context, dp models.DataPolicy) (err error) {
	api := c.TasksAPI()
	_, err = api.CreateTaskByFlux(ctx, getTaskFlux(dp), *c.DefaultOrg.Id)
	return err
}

// updateAggrTask updates a task of data aggregation.
func (c *Client) updateAggrTask(ctx context.Context, dp models.DataPolicy) (err error) {
	api := c.TasksAPI()
	// find task
	filter := iapi.TaskFilter{
		Name:  getTaskName(dp.Id),
		OrgID: *c.DefaultOrg.Id,
		Limit: 1,
	}
	tasks, err := api.FindTasks(ctx, &filter)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return ErrTaskNotFound
	}
	task := tasks[0]
	task.Flux = getTaskFlux(dp)
	_, err = api.UpdateTask(ctx, &task)
	return err
}

// deleteAggrTask deletes a task of data aggregation.
func (c *Client) deleteAggrTask(ctx context.Context, id int16) (err error) {
	api := c.TasksAPI()

	// find task
	filter := iapi.TaskFilter{
		Name:  getTaskName(id),
		OrgID: *c.DefaultOrg.Id,
		Limit: 1,
	}
	tasks, err := api.FindTasks(ctx, &filter)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return ErrTaskNotFound
	}
	task := tasks[0]

	return api.DeleteTask(ctx, &task)
}
