package tests_test

import (
	"net/url"
	"testing"

	"github.com/botscommunity/botsgo/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
	query := make(url.Values)

	func(properties ...any) {
		schema.NewSchema(schema.TypeDefs{
			schema.Integer: schema.NewType(schema.ParameterNames{"user_id", "group_id"}),
			schema.String:  schema.NewType(schema.ParameterNames{"message"}),
			schema.Struct:  nil,
		}).ConvertToQuery(query, properties...)
	}(101, 100, "Hello, World!", struct {
		CanRead bool `json:"can_read"`
	}{CanRead: true})

	assert.Equal(t, "101", query.Get("user_id"))
	assert.Equal(t, "100", query.Get("group_id"))
	assert.Equal(t, "Hello, World!", query.Get("message"))
	assert.Equal(t, "true", query.Get("can_read"))
}
