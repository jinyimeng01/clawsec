package tanswer

import (
	"context"
	"fmt"

	"github.com/clawsec/clawsec/pkg/products"
)

type TAnswer struct {
	products.BaseProduct
}

func New() *TAnswer {
	return &TAnswer{BaseProduct: products.BaseProduct{Name_: "tanswer", Headers: make(map[string]string)}}
}

func init() {
	products.Register("tanswer", New())
}

func (t *TAnswer) Connect(config products.Config) error {
	return fmt.Errorf("tanswer adapter not yet implemented")
}

func (t *TAnswer) Query(ctx context.Context, queryType string, params map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("tanswer adapter not yet implemented")
}

func (t *TAnswer) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("tanswer adapter not yet implemented")
}
