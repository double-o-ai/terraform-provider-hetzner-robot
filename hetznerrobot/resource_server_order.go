package hetznerrobot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServerOrder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerOrderCreate,
		ReadContext:   resourceServerOrderRead,
		DeleteContext: resourceServerOrderDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Required input
			"product_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Product ID to order (e.g. 'AX102' for standard, or market product ID as string)",
			},

			// Optional input
			"market": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Whether to order from the server market/auction (default: false)",
			},
			"authorized_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "List of SSH key fingerprints for initial server access",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"location": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Preferred datacenter location (e.g. 'FSN1', 'NBG1', 'HEL1'). Only for standard orders.",
			},
			"test": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Execute a test order (no actual server will be provisioned)",
			},

			// Computed outputs
			"transaction_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Order transaction ID",
			},
			"server_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Assigned server number",
			},
			"server_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server primary IPv4 address",
			},
			"server_ipv6": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server IPv6 network",
			},
			"server_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server name",
			},
			"product": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server product name",
			},
			"datacenter": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Datacenter location",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server status",
			},
			"is_cancelled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Server cancellation status",
			},
		},
	}
}

func resourceServerOrderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(HetznerRobotClient)

	productID := d.Get("product_id").(string)
	isMarket := d.Get("market").(bool)
	isTest := d.Get("test").(bool)
	location := d.Get("location").(string)

	var authorizedKeys []string
	if v, ok := d.GetOk("authorized_keys"); ok {
		for _, key := range v.([]interface{}) {
			authorizedKeys = append(authorizedKeys, key.(string))
		}
	}

	var txn *HetznerRobotOrderTransaction
	var err error

	if isMarket {
		marketProductID, convErr := strconv.Atoi(productID)
		if convErr != nil {
			return diag.Errorf("market product_id must be a numeric ID: %s", convErr)
		}
		txn, err = c.orderMarketServer(ctx, marketProductID, authorizedKeys, isTest)
	} else {
		txn, err = c.orderServer(ctx, productID, authorizedKeys, location, isTest)
	}

	if err != nil {
		return diag.Errorf("Failed to order server: %s", err)
	}

	d.Set("transaction_id", txn.ID)

	serverNumber, err := strconv.Atoi(txn.ServerNumber)
	if err != nil {
		return diag.Errorf("Invalid server number in order response: %s", txn.ServerNumber)
	}

	d.Set("server_number", serverNumber)
	d.Set("server_ip", txn.ServerIP)
	d.SetId(strconv.Itoa(serverNumber))

	// Wait for server to become ready
	timeout := d.Timeout(schema.TimeoutCreate)
	server, err := c.waitForServerReady(ctx, serverNumber, timeout)
	if err != nil {
		return diag.Errorf("Server ordered (number: %d) but timed out waiting for ready status: %s", serverNumber, err)
	}

	d.Set("server_ip", server.ServerIP)
	d.Set("server_ipv6", server.ServerIPv6)
	d.Set("server_name", server.ServerName)
	d.Set("product", server.Product)
	d.Set("datacenter", server.DataCenter)
	d.Set("status", server.Status)
	d.Set("is_cancelled", server.Cancelled)

	return diag.Diagnostics{}
}

func resourceServerOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(HetznerRobotClient)

	serverNumber, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid server number ID: %s", d.Id())
	}

	server, err := c.getServer(ctx, serverNumber)
	if err != nil {
		// If the server no longer exists, remove from state
		d.SetId("")
		return diag.Diagnostics{}
	}

	d.Set("server_number", server.ServerNumber)
	d.Set("server_ip", server.ServerIP)
	d.Set("server_ipv6", server.ServerIPv6)
	d.Set("server_name", server.ServerName)
	d.Set("product", server.Product)
	d.Set("datacenter", server.DataCenter)
	d.Set("status", server.Status)
	d.Set("is_cancelled", server.Cancelled)

	return diag.Diagnostics{}
}

func resourceServerOrderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(HetznerRobotClient)

	serverNumber, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid server number ID: %s", d.Id())
	}

	_, err = c.cancelServer(ctx, serverNumber, "now")
	if err != nil {
		return diag.FromErr(fmt.Errorf("Failed to cancel server %d: %w", serverNumber, err))
	}

	return diag.Diagnostics{}
}
