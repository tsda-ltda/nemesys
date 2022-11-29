package dhs

import (
	"testing"
	"time"

	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

func TestProcessDatalog(t *testing.T) {
	c := `BD_URFLEX_FX.sqlite3;Tue Nov 29 12:10:15 2022
metering_log_00.txt.gz;Tue Nov 29 11:55:56 2022
status_log_00.txt.gz;Mon Nov 28 15:57:13 2022
command_log_00.txt.gz;Mon Nov 28 15:57:13 2022
virtual_log_00.txt.gz;Mon Nov 28 14:39:09 2022
metering_log_01.txt.gz;Tue Nov 29 12:10:15 2022
status_log_01.txt.gz;Mon Nov 28 16:12:41 2022
command_log_01.txt.gz;Mon Nov 28 19:51:31 2022
virtual_log_01.txt.gz;Mon Nov 28 14:53:50 2022
status_log_02.txt.gz;Mon Nov 28 16:17:49 2022
command_log_02.txt.gz;Mon Nov 28 20:07:08 2022
virtual_log_02.txt.gz;Mon Nov 28 14:54:40 2022
metering_log_03.txt.gz;Mon Nov 28 14:07:53 2022
status_log_03.txt.gz;Mon Nov 28 16:28:20 2022
command_log_03.txt.gz;Mon Nov 28 20:22:39 2022
virtual_log_03.txt.gz;Mon Nov 28 15:10:18 2022
metering_log_04.txt.gz;Mon Nov 28 14:23:26 2022
status_log_04.txt.gz;Mon Nov 28 16:43:59 2022`

	timeMetering, _ := time.Parse(time.ANSIC, "Tue Nov 29 12:10:15 2022")
	timeStatus, _ := time.Parse(time.ANSIC, "Mon Nov 28 15:57:13 2022")
	timeCommand, _ := time.Parse(time.ANSIC, "Mon Nov 28 15:57:13 2022")
	timeVirutal, _ := time.Parse(time.ANSIC, "Mon Nov 28 14:39:09 2022")

	logs, err := processDownloadControlFile(c, models.FlexLegacyDatalogDownloadRegistry{
		Metering: timeMetering.Unix(),
		Status:   timeStatus.Unix(),
		Command:  timeCommand.Unix(),
		Virtual:  timeVirutal.Unix(),
	}, []models.FlexLegacyDatalogMetricRequest{
		{
			PortType: types.FLPTMetering,
		},
		{
			PortType: types.FLPTStatus,
		},
		{
			PortType: types.FLPTCommand,
		},
		{
			PortType: types.FLPTVirtual,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(logs) != 10 {
		t.Errorf("Expected 10 logs, got: %d", len(logs))
		return
	}
}
