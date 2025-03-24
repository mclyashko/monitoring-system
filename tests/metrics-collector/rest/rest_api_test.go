package metrics_collector_rest_api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const address = "http://localhost:8081"

var client = http.Client{
	Timeout: 5 * time.Minute,
}

type CreateMetricResponse struct {
	Time       time.Time `json:"time"`
	ServiceURL string    `json:"service_url"`
	MetricName string    `json:"metric_name"`
	PodName    string    `json:"pod_name"`
}

func TestCreateInvalidMetric(t *testing.T) {
	code, _ := createMetric(t, "test-service-go/metrics", "", "test-pod", 0.123456)

	require.Equal(t, http.StatusBadRequest, code, "unexpected status code for invalid metric")
}

type GetMetricResponse struct {
	Time        time.Time `json:"time"`
	ServiceURL  string    `json:"service_url"`
	MetricName  string    `json:"metric_name"`
	PodName     string    `json:"pod_name"`
	MetricValue float64   `json:"metric_value"`
}

func TestCreateAndGetMetricByMetricIdentity(t *testing.T) {
	metricToCreate := struct {
		ServiceURL  string
		MetricName  string
		PodName     string
		MetricValue float64
	}{"test-service-go/metrics", "system_cpu_usage", "test-pod", 0.123456}

	code, respIdentity := createMetric(t, metricToCreate.ServiceURL, metricToCreate.MetricName, metricToCreate.PodName, metricToCreate.MetricValue)
	require.Equal(t, http.StatusCreated, code, "unexpected status code when creating metric")

	code, respMetric := getMetricByMetricIdentity(t, respIdentity.Time, respIdentity.ServiceURL, respIdentity.MetricName, respIdentity.PodName)
	require.Equal(t, http.StatusOK, code, "unexpected status code when creating metric")

	require.Equal(t, respIdentity.Time, respMetric.Time, "unexpected time change")
	require.Equal(t, metricToCreate.ServiceURL, respMetric.ServiceURL, "unexpected service url change")
	require.Equal(t, metricToCreate.MetricName, respMetric.MetricName, "unexpected metric name change")
	require.Equal(t, metricToCreate.PodName, respMetric.PodName, "unexpected pod name value change")
	require.Equal(t, metricToCreate.MetricValue, respMetric.MetricValue, "unexpected metric value change")
}

func TestGetMetricInvalidMetricIdentity(t *testing.T) {
	code, _ := getMetricByMetricIdentity(t, time.Now(), "", "system_cpu_usage", "test-pod")

	require.Equal(t, code, http.StatusBadRequest, "unexpected status code when getting metric by empty service url")
}

func TestGetMetricNotFound(t *testing.T) {
	code, _ := getMetricByMetricIdentity(t, time.Now(), "no-service/metrics", "some_metric", "no-pod")

	require.Equal(t, code, http.StatusNotFound, "unexpected status code when getting metric by unexisting identity")
}

func createMetric(t *testing.T, serviceURL, metricName, podName string, metricValue float64) (code int, response CreateMetricResponse) {
	metric := map[string]interface{}{
		"service_url":  serviceURL,
		"metric_name":  metricName,
		"pod_name":     podName,
		"metric_value": metricValue,
	}
	metricJSON, err := json.Marshal(metric)
	require.NoError(t, err, "failed to serialize metric")

	resp, err := client.Post(address+"/metric", "application/json", bytes.NewReader(metricJSON))
	require.NoError(t, err, "failed to send request to create metric")
	defer resp.Body.Close()

	code = resp.StatusCode
	_ = json.NewDecoder(resp.Body).Decode(&response)

	return code, response
}

func getMetricByMetricIdentity(t *testing.T, timeT time.Time, serviceURL, metricName, podName string) (code int, response GetMetricResponse) {
	timeParam := url.QueryEscape(timeT.Format(time.RFC3339Nano))
	serviceURLParam := url.QueryEscape(serviceURL)
	metricNameParam := url.QueryEscape(metricName)
	podNameParam := url.QueryEscape(podName)

	url := fmt.Sprintf("%s/metric?time=%s&service_url=%s&metric_name=%s&pod_name=%s",
		address,
		timeParam,
		serviceURLParam,
		metricNameParam,
		podNameParam,
	)

	resp, err := client.Get(url)
	require.NoError(t, err, "failed to send request to get metric")
	defer resp.Body.Close()

	code = resp.StatusCode
	_ = json.NewDecoder(resp.Body).Decode(&response)

	return code, response
}
