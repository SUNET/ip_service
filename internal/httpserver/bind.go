package httpserver

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

func (s *Service) bindRequest(ctx context.Context, c *gin.Context, v interface{}) error {
	ctx, span := s.TP.Start(ctx, "httpserver:bindRequest")
	defer span.End()

	if c.ContentType() == gin.MIMEJSON {
		_ = c.ShouldBindJSON(v)
	}
	_ = s.bindRequestQuery(ctx, c, v)
	_ = c.ShouldBindQuery(v)
	return c.ShouldBindUri(v)
}

func (s *Service) bindRequestQuery(ctx context.Context, c *gin.Context, v interface{}) error {
	_, span := s.TP.Start(ctx, "httpserver:bindRequestQuery")
	defer span.End()

	refV := reflect.ValueOf(v).Elem()
	refT := reflect.ValueOf(v).Elem().Type()
	for i := 0; i < refT.NumField(); i++ {
		field := refT.Field(i)
		fieldType := field.Type
		fieldKey := field.Tag.Get("form")
		if fieldKey == "" {
			fieldKey = field.Name
		}
		switch fieldType.String() {
		case "map[string]string":
			v := c.QueryMap(fieldKey)
			if len(v) == 0 {
				continue
			}
			refV.FieldByName(field.Name).Set(reflect.ValueOf(v))
		case "*map[string]string":
			v := c.QueryMap(fieldKey)
			if len(v) == 0 {
				continue
			}
			refV.FieldByName(field.Name).Set(reflect.ValueOf(&v))
		case "map[string][]string":
			v := make(map[string][]string)
			for key := range c.QueryMap(fieldKey) {
				v[key] = c.QueryArray(fmt.Sprintf("%s[%s]", fieldKey, key))
			}
			if len(v) == 0 {
				continue
			}
			refV.FieldByName(field.Name).Set(reflect.ValueOf(v))
		case "*map[string][]string":
			v := make(map[string][]string)
			for key := range c.QueryMap(fieldKey) {
				v[key] = c.QueryArray(fmt.Sprintf("%s[%s]", fieldKey, key))
			}
			if len(v) == 0 {
				continue
			}
			refV.FieldByName(field.Name).Set(reflect.ValueOf(&v))
		}
	}
	return nil
}
