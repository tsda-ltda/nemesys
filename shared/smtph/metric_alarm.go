package smtph

import (
	"net/smtp"

	"github.com/fernandotsda/nemesys/shared/env"
)

// SendAlarmMetric send a metric alarm message to emails, using
// the enviroment metric alarm sender configuration.
func SendAlarmMetric(to []string, message []byte) (err error) {
	from := env.MetricAlarmEmailSender
	password := env.MetricAlarmEmailSenderPassword
	host := env.MetricAlarmEmailSenderHost
	port := env.MetricAlarmEmailSenderHostPort
	auth := smtp.PlainAuth("", from, password, host)
	return smtp.SendMail(host+":"+port, auth, from, to, message)
}
