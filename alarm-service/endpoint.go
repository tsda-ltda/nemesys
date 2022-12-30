package alarm

import (
	"bytes"
	"context"
	"net/http"

	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	jsoniter "github.com/json-iterator/go"
)

func (a *Alarm) notifyEndpoints(info models.AlarmNotificationInfo, profiles []models.AlarmProfileSimplified) {
	ids := make([]int32, len(profiles))
	for i, p := range profiles {
		ids[i] = p.Id
	}
	endpoints, err := a.pg.GetAlamProfilesAlarmEndpoints(context.Background(), ids)
	if err != nil {
		a.log.Error("Fail to get alarm profiles notifications enpoints", logger.ErrField(err))
		return
	}

	b, err := jsoniter.Marshal(info)
	if err != nil {
		a.log.Error("Fail to marshal alarm notification info", logger.ErrField(err))
		return
	}

	buffer := bytes.NewBuffer(b)
	for _, endpoint := range endpoints {
		req, err := http.NewRequest(http.MethodPost, endpoint.URL, buffer)
		if err != nil {
			a.log.Warn("Fail to create http request for "+endpoint.URL, logger.ErrField(err))
		}

		if len(endpoint.Headers) > 0 {
			for _, h := range endpoint.Headers {
				req.Header.Set(h.Header, h.Value)
			}
		}

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			a.log.Warn("Fail to do request for "+endpoint.URL, logger.ErrField(err))
			return
		}

		a.log.Debug("Notification sent with success, name: " + endpoint.Name)
	}
}
