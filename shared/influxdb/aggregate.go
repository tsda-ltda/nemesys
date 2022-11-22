package influxdb

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/models"
	iapi "github.com/influxdata/influxdb-client-go/v2/api"
)

// getTaskName returns a task name for a data policy id.
func getTaskName(dataPolicyId int16) string {
	return fmt.Sprintf("%d-aggr-task", dataPolicyId)
}

// getTaskFlux returns the task flux.
func getTaskFlux(dp models.DataPolicy) string {
	return fmt.Sprintf(`
		option task = {name: "%s", every: %s, offset: %s}
		from(bucket: "%s")
			|> range(start: -%s, stop: -%s)
			|> filter(fn: (r) => r._measurement == "metrics")
			|> aggregateWindow(every: %ds, fn: mean, createEmpty: false)
			|> to(bucket: "%s")
			`,
		getTaskName(dp.Id),
		fmt.Sprintf("%ds", dp.AggregationInterval),
		fmt.Sprintf("%ds", dp.Retention*3600),
		GetBucketName(dp.Id, false),
		fmt.Sprintf("%ds", dp.Retention*3600),
		fmt.Sprintf("%ds", dp.Retention*3600-dp.AggregationInterval),
		dp.AggregationInterval,
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
