package cloudwalker

import (
	"context"
	"fmt"

	"github.com/clawsec/clawsec/pkg/products"
)

type CloudWalker struct {
	products.BaseProduct
}

func New() *CloudWalker {
	return &CloudWalker{BaseProduct: products.BaseProduct{Name_: "cloudwalker", Headers: make(map[string]string)}}
}

func init() {
	products.Register("cloudwalker", New())
}

func (c *CloudWalker) Connect(config products.Config) error {
	return fmt.Errorf("cloudwalker adapter not yet implemented")
}

func (c *CloudWalker) Query(ctx context.Context, queryType string, params map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("cloudwalker adapter not yet implemented")
}

func (c *CloudWalker) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("cloudwalker adapter not yet implemented")
}
