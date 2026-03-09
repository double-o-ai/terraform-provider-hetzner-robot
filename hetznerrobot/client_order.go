package hetznerrobot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// --- Server Product Types ---

type HetznerRobotServerProduct struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description []string `json:"description"`
	Traffic     string   `json:"traffic"`
	Dist        []string `json:"dist"`
	Arch        []int    `json:"arch"`
	Lang        []string `json:"lang"`
	Location    []string `json:"location"`
	Prices      []struct {
		Location string `json:"location"`
		Price    struct {
			Net   string `json:"net"`
			Gross string `json:"gross"`
		} `json:"price"`
		PriceSetup struct {
			Net   string `json:"net"`
			Gross string `json:"gross"`
		} `json:"price_setup"`
	} `json:"price"`
	OrderableAddons []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"orderable_addons"`
}

type HetznerRobotServerProductResponse struct {
	Product HetznerRobotServerProduct `json:"product"`
}

// --- Server Market Product Types ---

type HetznerRobotMarketProduct struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Description   []string `json:"description"`
	Traffic       string   `json:"traffic"`
	Dist          []string `json:"dist"`
	Arch          []int    `json:"arch"`
	Lang          []string `json:"lang"`
	CPU           string   `json:"cpu"`
	CPUBenchmark  int      `json:"cpu_benchmark"`
	Memory        int      `json:"memory_size"`
	HddSize       int      `json:"hdd_size"`
	HddText       string   `json:"hdd_text"`
	HddCount      int      `json:"hdd_count"`
	Datacenter    string   `json:"datacenter"`
	NetworkSpeed  string   `json:"network_speed"`
	Price         string   `json:"price"`
	PriceSetup    string   `json:"price_setup"`
	FixedPrice    bool     `json:"fixed_price"`
	NextReduce    int      `json:"next_reduce"`
	NextReduceDate string  `json:"next_reduce_date"`
}

type HetznerRobotMarketProductResponse struct {
	Product HetznerRobotMarketProduct `json:"product"`
}

// --- Order Transaction Types ---

type HetznerRobotOrderTransaction struct {
	ID            string   `json:"id"`
	Date          string   `json:"date"`
	Status        string   `json:"status"`
	ServerNumber  string   `json:"server_number"`
	ServerIP      string   `json:"server_ip"`
	AuthorizedKey []string `json:"authorized_key"`
	HostKey       []string `json:"host_key"`
}

type HetznerRobotOrderTransactionResponse struct {
	Transaction HetznerRobotOrderTransaction `json:"transaction"`
}

// --- Cancellation Types ---

type HetznerRobotCancellation struct {
	ServerNumber       int    `json:"server_number"`
	CancellationDate   string `json:"cancellation_date"`
	EarliestCancellationDate string `json:"earliest_cancellation_date"`
	ServerIP           string `json:"server_ip"`
}

type HetznerRobotCancellationResponse struct {
	Cancellation HetznerRobotCancellation `json:"cancellation"`
}

// --- API Methods ---

func (c *HetznerRobotClient) getServerProducts(ctx context.Context) ([]HetznerRobotServerProduct, error) {
	res, err := c.makeAPICall(ctx, "GET", fmt.Sprintf("%s/order/server/product", c.url), nil, []int{http.StatusOK})
	if err != nil {
		return nil, err
	}

	var productResponses []HetznerRobotServerProductResponse
	if err = json.Unmarshal(res, &productResponses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal server products: %w", err)
	}

	products := make([]HetznerRobotServerProduct, len(productResponses))
	for i, pr := range productResponses {
		products[i] = pr.Product
	}
	return products, nil
}

func (c *HetznerRobotClient) getMarketProducts(ctx context.Context) ([]HetznerRobotMarketProduct, error) {
	res, err := c.makeAPICall(ctx, "GET", fmt.Sprintf("%s/order/server_market/product", c.url), nil, []int{http.StatusOK})
	if err != nil {
		return nil, err
	}

	var productResponses []HetznerRobotMarketProductResponse
	if err = json.Unmarshal(res, &productResponses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal market products: %w", err)
	}

	products := make([]HetznerRobotMarketProduct, len(productResponses))
	for i, pr := range productResponses {
		products[i] = pr.Product
	}
	return products, nil
}

func (c *HetznerRobotClient) orderServer(ctx context.Context, productID string, authorizedKeys []string, location string, test bool) (*HetznerRobotOrderTransaction, error) {
	data := url.Values{}
	data.Set("product_id", productID)
	for _, key := range authorizedKeys {
		data.Add("authorized_key[]", key)
	}
	if location != "" {
		data.Set("location", location)
	}
	if test {
		data.Set("test", "true")
	}

	res, err := c.makeAPICall(ctx, "POST", fmt.Sprintf("%s/order/server/transaction", c.url), data, []int{http.StatusOK, http.StatusCreated})
	if err != nil {
		return nil, err
	}

	var txnResponse HetznerRobotOrderTransactionResponse
	if err = json.Unmarshal(res, &txnResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order transaction: %w", err)
	}
	return &txnResponse.Transaction, nil
}

func (c *HetznerRobotClient) orderMarketServer(ctx context.Context, productID int, authorizedKeys []string, test bool) (*HetznerRobotOrderTransaction, error) {
	data := url.Values{}
	data.Set("product_id", fmt.Sprintf("%d", productID))
	for _, key := range authorizedKeys {
		data.Add("authorized_key[]", key)
	}
	if test {
		data.Set("test", "true")
	}

	res, err := c.makeAPICall(ctx, "POST", fmt.Sprintf("%s/order/server_market/transaction", c.url), data, []int{http.StatusOK, http.StatusCreated})
	if err != nil {
		return nil, err
	}

	var txnResponse HetznerRobotOrderTransactionResponse
	if err = json.Unmarshal(res, &txnResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal market order transaction: %w", err)
	}
	return &txnResponse.Transaction, nil
}

func (c *HetznerRobotClient) cancelServer(ctx context.Context, serverNumber int, cancellationDate string) (*HetznerRobotCancellation, error) {
	data := url.Values{}
	if cancellationDate == "" {
		cancellationDate = "now"
	}
	data.Set("cancellation_date", cancellationDate)

	res, err := c.makeAPICall(ctx, "POST", fmt.Sprintf("%s/server/%d/cancellation", c.url, serverNumber), data, []int{http.StatusOK, http.StatusCreated})
	if err != nil {
		return nil, err
	}

	var cancelResponse HetznerRobotCancellationResponse
	if err = json.Unmarshal(res, &cancelResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cancellation response: %w", err)
	}
	return &cancelResponse.Cancellation, nil
}

func (c *HetznerRobotClient) withdrawCancellation(ctx context.Context, serverNumber int) error {
	_, err := c.makeAPICall(ctx, "DELETE", fmt.Sprintf("%s/server/%d/cancellation", c.url, serverNumber), nil, []int{http.StatusOK})
	return err
}

func (c *HetznerRobotClient) waitForServerReady(ctx context.Context, serverNumber int, timeout time.Duration) (*HetznerRobotServer, error) {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for server %d to become ready", serverNumber)
		}

		server, err := c.getServer(ctx, serverNumber)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				time.Sleep(30 * time.Second)
				continue
			}
			return nil, fmt.Errorf("error checking server %d status: %w", serverNumber, err)
		}

		if server.Status == "ready" {
			return server, nil
		}

		time.Sleep(30 * time.Second)
	}
}
