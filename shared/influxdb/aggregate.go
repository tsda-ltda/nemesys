package influxdb

import (
	"context"
	"fmt"

	"github.com/fernandotsda/nemesys/shared/models"
	iapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// getTaskName returns a task name for a data policy id.
func getTaskName(dataPolicyId int16) string {
	return fmt.Sprintf("%d-aggr-task", dataPolicyId)
}

// createAggrTask creates an task of data aggregation.
func (c *Client) createAggrTask(ctx context.Context, dp models.DataPolicy) (err error) {
	api := c.TasksAPI()

	every := fmt.Sprintf("%ds", dp.AggregationInterval)
	flux := fmt.Sprintf(`
			from(bucket: "%s")
				|> range(start: -task.every)
				|> aggregateWindow(every: %ds, fn: mean)
				|> to(bucket: "%s")
		`,
		getBucketName(dp.Id, false),
		dp.AggregationInterval,
		getBucketName(dp.Id, true),
	)

	_, err = api.CreateTask(ctx, &domain.Task{
		Flux:  flux,
		Every: &every,
		Name:  getTaskName(dp.Id),
		OrgID: *c.DefaultOrg.Id,
	})
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
	every := fmt.Sprintf("%ds", dp.AggregationInterval)

	// update params
	task.Every = &every
	task.Flux = fmt.Sprintf(`
			option task = {name: "%s", every: %s}

			from(bucket: "%s")
				|> range(start: -task.every)
				|> aggregateWindow(every: %ds, fn: mean)
				|> to(bucket: "%s")
		`,
		getTaskName(dp.Id),
		every,
		getBucketName(dp.Id, false),
		dp.AggregationInterval,
		getBucketName(dp.Id, true),
	)

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
