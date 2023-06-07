package rabbit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCORSMiddleware(t *testing.T) {
	// Setting up the Gin router and the CORS middleware
	router := gin.Default()
	router.Use(CORSEnabled())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello")
	})

	// Creating an HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	assert.Nil(t, err)

	// Recording the HTTP response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verifying the CORS headers in the response
	assert.Equal(t, CORS_ALLOW_ALL, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, CORS_ALLOW_CREDENTIALS, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, CORS_ALLOW_HEADERS, w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, CORS_ALLOW_METHODS, w.Header().Get("Access-Control-Allow-Methods"))

	// Verifying the response code and body
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello", w.Body.String())

	// Creating an HTTP request with OPTIONS method
	req2, _ := http.NewRequest("OPTIONS", "/", nil)
	req2.Header.Set("Origin", "http://example.com")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNoContent, w2.Code)
}

func TestWithCookieSessionMiddleware(t *testing.T) {
	const SessionSecret = "my_session_secret_key"

	// Setting up the Gin router and the WithCookieSession middleware
	router := gin.Default()
	router.Use(WithCookieSession(SessionSecret))
	router.GET("/set", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("key", "value")
		session.Save()
	})
	router.GET("/get", func(c *gin.Context) {
		session := sessions.Default(c)
		val := session.Get("key")
		if val == nil {
			c.String(http.StatusBadRequest, "")
			return
		}
		c.String(http.StatusOK, val.(string))
	})

	// Set cookie
	req, _ := http.NewRequest("GET", "/set", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Verifying the Set-Cookie header in the response
	setCookie := w.Header().Get("Set-Cookie")
	assert.Contains(t, setCookie, SessionField+"=")

	// Get session, with Cookie header
	req2, _ := http.NewRequest("GET", "/get", nil)
	req2.Header.Set("Cookie", setCookie) // with Cookie header
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "value", w2.Body.String())

	// Get session, but no Cookie header
	req3, _ := http.NewRequest("GET", "/get", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestWithMemSession(t *testing.T) {
	const SessionSecret = "my_mem_session_key"

	// Setting up the Gin router and the WithMemSession middleware
	router := gin.Default()
	router.Use(WithMemSession(SessionSecret))
	router.GET("/set", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("key", "value")
		session.Save()
	})
	router.GET("/get", func(c *gin.Context) {
		session := sessions.Default(c)
		val := session.Get("key")
		if val == nil {
			c.String(http.StatusBadRequest, "")
			return
		}
		c.String(http.StatusOK, val.(string))
	})

	// Set session
	req, _ := http.NewRequest("GET", "/set", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Verifying the Set-Cookie header in the response
	setCookie := w.Header().Get("Set-Cookie")
	assert.Contains(t, setCookie, SessionField+"=")

	// Get session, with Cookie header
	req2, _ := http.NewRequest("GET", "/get", nil)
	req2.Header.Set("Cookie", setCookie) // with Cookie header
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "value", w2.Body.String())

	// Get session, but no Cookie header
	req3, _ := http.NewRequest("GET", "/get", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestWithGormDB(t *testing.T) {
	router := gin.New()
	db := &gorm.DB{} // create a dummy gorm.DB instance
	router.Use(WithGormDB(db))

	router.GET("/test", func(c *gin.Context) {
		db := c.MustGet(DbField).(*gorm.DB)
		c.String(http.StatusOK, fmt.Sprintf("%v", db))
	})

	// Create a new HTTP request.
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	// Perform the HTTP request and check the response.
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, fmt.Sprintf("%v", db), resp.Body.String())
}
