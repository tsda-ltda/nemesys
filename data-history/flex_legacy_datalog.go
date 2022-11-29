package dhs

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fernandotsda/nemesys/shared/env"
	"github.com/fernandotsda/nemesys/shared/evaluator"
	"github.com/fernandotsda/nemesys/shared/logger"
	"github.com/fernandotsda/nemesys/shared/models"
	"github.com/fernandotsda/nemesys/shared/types"
)

const (
	DownloadControlPath     = "downloads/log/DownloadControl.txt"
	DownloadDatalogBasePath = "downloads/log/"
)

type flexLegacyDatalogRequest struct {
	// Type is the flex legacy datalog port type.
	Type types.FlexLegacyPortType
	// FileName is the file name.
	FileName string
	// Timestamp is the datalog timestamp.
	Timestamp time.Time
}

type flexLegacyDataResponse struct {
	// Err is the non fatal error happened during the process due to
	// metric wrong configuration.
	Err error
	// Alarmed is the alarmed state of the flex legacy port.
	Alarmed bool
	// MetricId is the metric id.
	MetricId int64
	// MetricType is the metric type.
	MetricType types.MetricType
	// DataPolicyId is the data policy id.
	DataPolicyId int16
	// Timestamp is the data timestamp.
	Timestamp time.Time
	// Value is the data value;
	Value any
}

type flexLegacyDatalogWorker struct {
	// workerNumber is the worker number.
	workerNumber int64
	// requests is the chan of datalog request.
	requests <-chan int32
	// httpClient is the http client.
	httpClient *http.Client
	// dhs is the Data History service pointer.
	dhs *DHS
	// closed is thhe closed chan.
	closed chan any
}

func (d *DHS) getFlexLegacyDatalog(id int32) {
	d.getFlexLegacyDatalogCh <- id
}

func (d *DHS) createFlexLegacyWorkers() {
	n, err := strconv.ParseInt(env.DHSFlexLegacyDatalogWorkers, 0, 10)
	if err != nil {
		d.log.Fatal("Fail to parse env.DHSFlexLegacyDatalogWorkers to int, received: " + env.DHSFlexLegacyDatalogWorkers)
		return
	}

	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	var _n int64
	for _n <= n {
		worker := &flexLegacyDatalogWorker{
			workerNumber: _n,
			requests:     d.getFlexLegacyDatalogCh,
			httpClient:   &httpClient,
			dhs:          d,
		}
		go worker.Run()
		d.flexLegacyDatalogWorkers = append(d.flexLegacyDatalogWorkers, worker)
		_n++
	}
	d.log.Info(fmt.Sprintf("%d flex legacy datalog workers created", _n))
}

func (w *flexLegacyDatalogWorker) Close() {
	w.closed <- nil
}

func (w *flexLegacyDatalogWorker) Run() {
	for {
		select {
		case id := <-w.requests:
			ctx := context.Background()
			idString := strconv.FormatInt(int64(id), 10)
			logField := logger.Int64Field("worker", w.workerNumber)
			w.dhs.log.Debug("Starting flex legacy datalog fetch process, container id: "+idString, logField)

			exists, target, err := w.dhs.pg.GetFlexLegacyContainerTarget(ctx, id)
			if err != nil {
				w.dhs.log.Error("Fail to get Flex legacy target value on database", logger.ErrField(err), logField)
				continue
			}
			if !exists {
				w.dhs.log.Error("Fail to get Flex legacy target value, container does not exists", logField)
				continue
			}

			registryExists, registry, err := w.dhs.pg.GetFlexLegacyDatalogDownloadRegistry(ctx, id)
			if err != nil {
				w.dhs.log.Error("Fail to get flex legacy datalog download registry on database", logger.ErrField(err), logField)
				continue
			}

			// check if is the first time downloading the datalogs
			if !registryExists {
				registry = models.FlexLegacyDatalogDownloadRegistry{
					ContainerId: id,
				}
			}
			newRegistry := registry

			txt, err := w.fetchFlexLegacyDownloadControl(target)
			if err != nil {
				w.dhs.log.Warn("Fail to download the DownloadControl", logger.ErrField(err), logger.StringField("target", target), logField)
				continue
			}

			metricsRequests, err := w.dhs.pg.GetFlexLegacyMetricsRequests(ctx, id)
			if err != nil {
				w.dhs.log.Error("Fail to get flex legacy metrics requests on database", logger.ErrField(err), logField)
				continue
			}

			ids := make([]int64, len(metricsRequests))
			for i, m := range metricsRequests {
				ids[i] = m.Id
			}
			expressionsMap := make(map[int64]string, len(metricsRequests))
			expressions, err := w.dhs.pg.GetMetricsEvaluableExpressions(ctx, ids)
			if err != nil {
				w.dhs.log.Error("Fail to get metrics evaluable expressions", logger.ErrField(err), logField)
				continue
			}
			for _, e := range expressions {
				expressionsMap[e.Id] = e.Expression
			}

			requests, err := processDownloadControlFile(txt, registry, metricsRequests)
			if err != nil {
				w.dhs.log.Error("Fail to process DownloadControl", logger.ErrField(err), logField)
				continue
			}
			for _, req := range requests {
				b, err := w.fetchDatalog(target, req.FileName)
				if err != nil {
					w.dhs.log.Error("Fail to fetch datalog", logger.StringField("filename", req.FileName), logger.StringField("target", target), logField)
					continue
				}
				metricsData, err := processDatalog(b, metricsRequests)
				for _, data := range metricsData {
					if data.Err != nil {
						w.dhs.log.Warn("Flex legacy metric datalog process failed"+strconv.FormatInt(data.MetricId, 10), logger.ErrField(err), logField)
						continue
					}

					v, err := evaluator.DirectEvaluation(data.Value, data.MetricType, expressionsMap[data.MetricId])
					if err != nil {
						w.dhs.log.Warn("Fail to do direct evaluation of datalog point, metric id: "+strconv.FormatInt(data.MetricId, 10), logField)
						continue
					}

					err = w.dhs.influxClient.WritePoint(ctx, models.MetricDataResponse{
						MetricBasicDataReponse: models.MetricBasicDataReponse{
							Id:           data.MetricId,
							Type:         data.MetricType,
							Value:        v,
							DataPolicyId: data.DataPolicyId,
							Failed:       false,
						},
					}, data.Timestamp)
					if err != nil {
						w.dhs.log.Error("Fail to write flex legacy metric datalog data on influxdb", logger.ErrField(err), logField)
						continue
					}

					// do something with alarmed ...
				}

				t := req.Timestamp.Unix()
				switch req.Type {
				case types.FLPTMetering:
					if t > newRegistry.Metering {
						newRegistry.Metering = t
					}
				case types.FLPTCommand:
					if t > newRegistry.Command {
						newRegistry.Command = t
					}
				case types.FLPTStatus:
					if t > newRegistry.Status {
						newRegistry.Status = t
					}
				case types.FLPTVirtual:
					if t > newRegistry.Virtual {
						newRegistry.Virtual = t
					}
				}
			}
			if registryExists {
				_, err = w.dhs.pg.UpdateFlexLegacyDatalogDownloadRegistry(ctx, newRegistry)
				if err != nil {
					w.dhs.log.Error("Fail to update flex legaly datalog download registry", logger.ErrField(err), logField)
					continue
				}
			} else {
				err = w.dhs.pg.CreateFlexLegacyDatalogDownloadRegistry(ctx, newRegistry)
				if err != nil {
					w.dhs.log.Error("Fail to create flex legacy donwload registry", logger.ErrField(err), logField)
					continue
				}
			}
			w.dhs.log.Debug(fmt.Sprintf("Flex legacy datalog process finished with success, %d datalogs downloaded for %d metrics", len(requests), len(metricsRequests)), logField)
		case <-w.closed:
			return
		}
	}
}

func (w *flexLegacyDatalogWorker) fetchDatalog(target string, filename string) (bytes []byte, err error) {
	url := fmt.Sprintf("http://%s/%s/%s", target, DownloadDatalogBasePath, filename)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}

func (w *flexLegacyDatalogWorker) fetchFlexLegacyDownloadControl(target string) (txt string, err error) {
	url := fmt.Sprintf("http://%s/%s", target, DownloadControlPath)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return txt, err
	}

	res, err := w.httpClient.Do(req)
	if err != nil {
		return txt, err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return txt, err
	}

	return string(resBody), nil
}

func processDownloadControlFile(txt string, resgistry models.FlexLegacyDatalogDownloadRegistry, metrics []models.FlexLegacyDatalogMetricRequest) (logs []flexLegacyDatalogRequest, err error) {
	skipMetering := true
	skipStatus := true
	skipCommand := true
	skipVirtual := true
	for _, m := range metrics {
		switch m.PortType {
		case types.FLPTCommand:
			skipCommand = false
		case types.FLPTMetering:
			skipMetering = false
		case types.FLPTVirtual:
			skipVirtual = false
		case types.FLPTStatus:
			skipStatus = false
		}
	}

	lines := strings.Split(txt, "\n")
	logs = make([]flexLegacyDatalogRequest, 0, len(lines)-1)

	// i = 1 because the txt first line is the database name.
	for i := 1; i < len(lines); i++ {
		// Line example: "virtual_log_00.txt.gz;Fri Nov 25 18:26:37 2022"
		var log flexLegacyDatalogRequest

		splited := strings.SplitN(lines[i], ";", 2)
		if len(splited) != 2 {
			continue
		}

		log.FileName = splited[0]
		timestamp, err := time.Parse(time.ANSIC, splited[1])
		if err != nil {
			return logs, errors.New("Fail to parse line date using ANSIC format, err: " + err.Error())
		}
		log.Timestamp = timestamp

		splited = strings.SplitN(splited[0], "_", 2)
		if len(splited) != 2 {
			continue
		}

		datalogType := splited[0]
		log.Type, err = types.ParseFlexPortType(datalogType)
		if err != nil {
			return nil, err
		}

		// Check if the datalog have metric request of this type and is newer then the last downloaded one.
		// If true append it to the response.
		switch log.Type {
		case types.FLPTMetering:
			if timestamp.Unix() > resgistry.Metering && !skipMetering {
				logs = append(logs, log)
			}
		case types.FLPTStatus:
			if timestamp.Unix() > resgistry.Status && !skipStatus {
				logs = append(logs, log)
			}
		case types.FLPTCommand:
			if timestamp.Unix() > resgistry.Command && !skipCommand {
				logs = append(logs, log)
			}
		case types.FLPTVirtual:
			if timestamp.Unix() > resgistry.Virtual && !skipVirtual {
				logs = append(logs, log)
			}
		default:
			return logs, errors.New("Unsupported datalog type, type: " + datalogType)
		}
	}
	return logs, nil
}

func processDatalog(b []byte, metrics []models.FlexLegacyDatalogMetricRequest) (r []flexLegacyDataResponse, err error) {
	reader := bytes.NewReader(b)
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		return r, err
	}
	defer gzreader.Close()

	output, err := io.ReadAll(gzreader)
	if err != nil {
		return r, err
	}
	lines := strings.Split(string(output), "\n")
	r = make([]flexLegacyDataResponse, 0, len(lines)-1)
	// i = 1 because the first line is a header.
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if len(line) != 97 {
			continue
		}

		t := line[6:16]
		portType, err := types.ParseFlexPortType(t[:8])
		if err != nil {
			return nil, err
		}
		port, err := strconv.ParseInt(t[8:], 0, 8)
		if err != nil {
			return r, err
		}

		var metric *models.FlexLegacyDatalogMetricRequest
		for _, m := range metrics {
			if m.Port == int16(port) && m.PortType == portType {
				metric = &m
				break
			}
		}
		if metric == nil {
			continue
		}

		value, err := types.ParseValue(strings.TrimSpace(line[55:63]), metric.Type)
		if err != nil {
			r = append(r, flexLegacyDataResponse{
				MetricId: metric.Id,
				Err:      err,
			})
			continue
		}

		alarmed := !(line[64:65] == "0")
		timestamp, err := strconv.ParseInt(line[87:], 0, 64)
		if err != nil {
			return r, err
		}

		r = append(r, flexLegacyDataResponse{
			Err:          nil,
			Alarmed:      alarmed,
			MetricId:     metric.Id,
			Timestamp:    time.Unix(timestamp, 0),
			MetricType:   metric.Type,
			DataPolicyId: metric.DataPolicyId,
			Value:        value,
		})
	}
	return r, nil
}
