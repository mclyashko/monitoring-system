package metrics_collector_grpc_api_test

import (
	"context"
	"testing"
	"time"

	metricspb "github.com/mclyashko/monitoring-system/tests/test-service-go/metrics-collector/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const grpcAddress = "localhost:81"

func TestGrpcPing(t *testing.T) {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	c := metricspb.NewMetricsCollectorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = c.Ping(ctx, &emptypb.Empty{})
	require.NoError(t, err)
}

func TestSendMetric(t *testing.T) {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	c := metricspb.NewMetricsCollectorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testTime := time.Now().UTC()
	req := &metricspb.SendMetricRequest{
		ServiceUrl:  "test-service-go/metrics",
		MetricName:  "system_cpu_usage",
		PodName:     "test-pod",
		MetricValue: 0.123456,
	}

	resp, err := c.SendMetric(ctx, req)
	require.NoError(t, err)

	respTime := resp.Time.AsTime()
	require.WithinDuration(t, testTime, respTime, time.Second)
	require.Equal(t, req.ServiceUrl, resp.ServiceUrl)
	require.Equal(t, req.MetricName, resp.MetricName)
	require.Equal(t, req.PodName, resp.PodName)
}

func TestSendInvalidMetric(t *testing.T) {
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	c := metricspb.NewMetricsCollectorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &metricspb.SendMetricRequest{
		ServiceUrl:  "",
		MetricName:  "system_cpu_usage",
		PodName:     "test-pod",
		MetricValue: 0.123456,
	}

	_, err = c.SendMetric(ctx, req)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok, "expected gRPC status error")
	require.Equal(t, codes.InvalidArgument, st.Code())
}
