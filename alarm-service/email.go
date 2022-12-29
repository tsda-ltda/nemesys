package alarm

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func getEmailMessage(info models.AlarmNotificationInfo) string {
	return fmt.Sprintf(`METRIC '%s' ALARM!
	
Description: %s
Occurency date:	%s
------------------------------
Alarm Category id: %d
Alarm Category name: %s
Alarm Category Level: %d
------------------------------
Metric id: %d
Metric Name: %s
Metric Value: %v
------------------------------
Container id: %d
Container Name: %s
Container Type: %s`,
		info.AlarmCategory.Name,
		info.Descr,
		time.Unix(info.OccurencyDate, 0).Format(time.RFC3339),
		info.AlarmCategory.Id,
		info.AlarmCategory.Name,
		info.AlarmCategory.Level,
		info.MetricId,
		info.MetricName,
		info.Value,
		info.ContainerId,
		info.ContainerName,
		types.StringfyContainerType(info.ContainerType),
	)
}

func (a *Alarm) notifyEmail(info models.AlarmNotificationInfo, profiles []models.AlarmProfileSimplified) {
	ids := make([]int32, len(profiles))
	for i, p := range profiles {
		ids[i] = p.Id
	}
	emails, err := a.pg.GetAlarmProfilesEmails(context.Background(), ids)
	if err != nil {
		a.log.Error("Fail to get alarm profiles on database", logger.ErrField(err))
		return
	}
	if len(emails) == 0 {
		return
	}

	bytes := []byte(getEmailMessage(info))

	err = smtp.SendMail(
		env.MetricAlarmEmailSenderHost+":"+env.MetricAlarmEmailSenderHostPort,
		a.smtpAuth,
		env.MetricAlarmEmailSender,
		emails,
		bytes,
	)
	if err != nil {
		a.log.Error("Fail to send alarm emails", logger.ErrField(err))
		return
	}
	a.log.Info("Alarm emails sent with success")
}
