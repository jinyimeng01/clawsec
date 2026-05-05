package ddr

import (
	"context"
	"fmt"

	"github.com/clawsec/clawsec/pkg/products"
)

type DDR struct {
	products.BaseProduct
}

func New() *DDR {
	return &DDR{BaseProduct: products.BaseProduct{Name_: "ddr", Headers: make(map[string]string)}}
}

func init() {
	products.Register("ddr", New())
}

func (d *DDR) Connect(config products.Config) error {
	return fmt.Errorf("ddr adapter not yet implemented")
}

func (d *DDR) Query(ctx context.Context, queryType string, params map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("ddr adapter not yet implemented")
}

func (d *DDR) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ddr adapter not yet implemented")
}
