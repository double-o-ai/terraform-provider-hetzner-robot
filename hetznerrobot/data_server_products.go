package hetznerrobot

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataServerProducts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceServerProductsRead,
		Schema: map[string]*schema.Schema{
			"products": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"traffic": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"location": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceServerProductsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(HetznerRobotClient)

	products, err := c.getServerProducts(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	productList := make([]map[string]interface{}, len(products))
	for i, p := range products {
		productList[i] = map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"traffic":     p.Traffic,
			"location":    p.Location,
		}
	}

	if err := d.Set("products", productList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("server_products")

	return diag.Diagnostics{}
}

func dataMarketProducts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMarketProductsRead,
		Schema: map[string]*schema.Schema{
			"products": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cpu": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cpu_benchmark": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"hdd_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"hdd_text": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hdd_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"datacenter": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"traffic": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"network_speed": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"price_setup": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_reduce": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Hours until next price reduction",
						},
						"next_reduce_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			// Optional filters
			"min_memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Minimum RAM in GB",
			},
			"min_cpu_benchmark": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Minimum CPU benchmark score",
			},
			"datacenter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by datacenter (e.g. 'FSN1-DC14')",
			},
			"max_price": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Maximum monthly price"),
			},
		},
	}
}

func dataSourceMarketProductsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(HetznerRobotClient)

	products, err := c.getMarketProducts(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Apply filters
	minMemory, _ := d.Get("min_memory").(int)
	minCPU, _ := d.Get("min_cpu_benchmark").(int)
	dcFilter, _ := d.Get("datacenter").(string)
	maxPriceStr, _ := d.Get("max_price").(string)

	var maxPrice float64
	if maxPriceStr != "" {
		maxPrice, _ = strconv.ParseFloat(maxPriceStr, 64)
	}

	var filtered []HetznerRobotMarketProduct
	for _, p := range products {
		if minMemory > 0 && p.Memory < minMemory {
			continue
		}
		if minCPU > 0 && p.CPUBenchmark < minCPU {
			continue
		}
		if dcFilter != "" && p.Datacenter != dcFilter {
			continue
		}
		if maxPrice > 0 {
			price, _ := strconv.ParseFloat(p.Price, 64)
			if price > maxPrice {
				continue
			}
		}
		filtered = append(filtered, p)
	}

	productList := make([]map[string]interface{}, len(filtered))
	for i, p := range filtered {
		productList[i] = map[string]interface{}{
			"id":               p.ID,
			"name":             p.Name,
			"description":      p.Description,
			"cpu":              p.CPU,
			"cpu_benchmark":    p.CPUBenchmark,
			"memory_size":      p.Memory,
			"hdd_size":         p.HddSize,
			"hdd_text":         p.HddText,
			"hdd_count":        p.HddCount,
			"datacenter":       p.Datacenter,
			"traffic":          p.Traffic,
			"network_speed":    p.NetworkSpeed,
			"price":            p.Price,
			"price_setup":      p.PriceSetup,
			"next_reduce":      p.NextReduce,
			"next_reduce_date": p.NextReduceDate,
		}
	}

	if err := d.Set("products", productList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("market_products")

	return diag.Diagnostics{}
}
