package test_service_go_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const address = "http://localhost:8080"

var client = http.Client{
	Timeout: 5 * time.Minute,
}

type CreateOrderResponse struct {
	ID int `json:"id"`
}

func TestCreateOrder(t *testing.T) {
	code, resp := createOrder(t, 1, 10, 123)

	require.Equal(t, http.StatusCreated, code, "unexpected status code when creating order")
	require.Greater(t, resp.ID, 0, "order ID should be greater than 0")
}

func TestCreateInvalidOrder(t *testing.T) {
	code, _ := createOrder(t, 0, 0, 0)

	require.Equal(t, http.StatusBadRequest, code, "unexpected status code for invalid order")
}

type GetOrderResponse struct {
	ID        int `json:"id"`
	ProductId int `json:"product_id"`
	Quantity  int `json:"quantity"`
	UserId    int `json:"user_id"`
}

func TestCreateAndGetOrderById(t *testing.T) {
	orderToCreate := struct{ ProductID, Quantity, UserID int }{10, 100, 321}

	code, respId := createOrder(t, orderToCreate.ProductID, orderToCreate.Quantity, orderToCreate.UserID)
	require.Equal(t, http.StatusCreated, code, "unexpected status code when creating order")

	code, respOrder := getOrderById(t, respId.ID)
	require.Equal(t, http.StatusOK, code, "unexpected status code when creating order")

	require.Equal(t, respId.ID, respOrder.ID, "unexpected id change")
	require.Equal(t, orderToCreate.ProductID, respOrder.ProductId, "unexpected product id change")
	require.Equal(t, orderToCreate.Quantity, respOrder.Quantity, "unexpected quantity change")
	require.Equal(t, orderToCreate.UserID, respOrder.UserId, "unexpected user id change")
}

func TestGetOrderInvalidOrderId(t *testing.T) {
	code, _ := getOrderById(t, -123)

	require.Equal(t, code, http.StatusBadRequest, "unexpected status code when getting order by negative id")
}

func TestGetOrderNotFound(t *testing.T) {
	code, _ := getOrderById(t, 99999)

	require.Equal(t, code, http.StatusNotFound, "unexpected status code when getting order by unexisting id")
}

func createOrder(t *testing.T, productID, quantity, userID int) (code int, response CreateOrderResponse) {
	order := map[string]interface{}{
		"product_id": productID,
		"quantity":   quantity,
		"user_id":    userID,
	}
	orderJSON, err := json.Marshal(order)
	require.NoError(t, err, "failed to serialize order")

	resp, err := client.Post(address+"/order", "application/json", bytes.NewReader(orderJSON))
	require.NoError(t, err, "failed to send request to create order")
	defer resp.Body.Close()

	code = resp.StatusCode
	_ = json.NewDecoder(resp.Body).Decode(&response)

	return code, response
}

func getOrderById(t *testing.T, orderId int) (code int, response GetOrderResponse) {
	url := address + "/order/" + strconv.Itoa(orderId)

	resp, err := client.Get(url)
	require.NoError(t, err, "failed to send request to get order")
	defer resp.Body.Close()

	code = resp.StatusCode
	_ = json.NewDecoder(resp.Body).Decode(&response)

	return code, response
}
