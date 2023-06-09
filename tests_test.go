package rabbit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTests(t *testing.T) {
	type user struct {
		ID   uint   `json:"id" gorm:"primarykey"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	r := gin.Default()

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, true)
	})
	r.POST("/pong", func(ctx *gin.Context) {
		var form user
		ctx.BindJSON(&form)
		ctx.JSON(http.StatusOK, gin.H{
			"name": form.Name,
			"age":  form.Age,
		})
	})

	c := NewTestClient(r)

	// GET
	{
		var result bool
		err := c.CallGet("/ping", nil, &result)
		assert.Nil(t, err)
		assert.True(t, result)
	}
	{
		w := c.Get("/ping")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "true", w.Body.String())
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "true", w.Body.String())
	}

	// POST
	{
		var result map[string]any
		err := c.CallPost("/pong", user{Name: "test", Age: 11}, &result)

		assert.Nil(t, err)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, float64(11), result["age"])
	}
	{
		b, _ := json.Marshal(user{Name: "test", Age: 11})
		w := c.Post("/pong", b)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"age":11,"name":"test"}`, w.Body.String())
	}
	{
		b, _ := json.Marshal(user{Name: "test", Age: 11})
		req := httptest.NewRequest(http.MethodPost, "/pong", bytes.NewReader(b))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"age":11,"name":"test"}`, w.Body.String())
	}
}
