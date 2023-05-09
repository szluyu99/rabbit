package rabbit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestDoGet Quick Test CheckResponse
func checkResponse(t *testing.T, w *httptest.ResponseRecorder) (response map[string]any) {
	assert.Equal(t, http.StatusOK, w.Code)
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	return response
}

func initTestClient(t *testing.T) (*gorm.DB, *gin.Engine, *TestClient) {
	gin.DisableConsoleColor()
	db := InitDatabase("", "", nil)
	r := gin.Default()

	InitRabbit(db, r)

	client := NewTestClient(r)
	return db, r, client
}

func TestRabbitInit(t *testing.T) {
	gin.DisableConsoleColor()
	db := InitDatabase("", "", nil)

	r := gin.Default()
	InitRabbit(db, r)

	r.GET("/mock_test", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{}) })
	client := NewTestClient(r)
	w := client.Get("/mock_test")
	checkResponse(t, w)
	assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), CORS_ALLOW_ALL)
}

func TestSession(t *testing.T) {
	r := gin.Default()
	r.Use(WithCookieSession("hello"))
	r.GET("/mock", func(ctx *gin.Context) {
		s := sessions.Default(ctx)
		s.Set(UserField, "test")
		s.Save()
	})
	client := NewTestClient(r)
	w := client.Get("/mock")
	assert.Contains(t, w.Header(), "Set-Cookie")
	assert.Contains(t, w.Header().Get("Set-Cookie"), SessionField+"=")
}

func TestAuthHandler(t *testing.T) {
	db, _, client := initTestClient(t)

	{
		form := RegisterUserForm{}
		err := client.CallPost("/auth/register", form, nil)
		assert.Contains(t, err.Error(), "'Email' failed on the 'required'")
	}
	{
		form := LoginForm{}
		err := client.CallPost("/auth/login", form, nil)
		assert.Contains(t, err.Error(), "email is required")
	}
	{
		form := RegisterUserForm{
			Email:    "bob@example.org",
			Password: "hello12345",
		}
		var user User
		err := client.CallPost("/auth/register", form, &user)
		assert.Nil(t, err)
		assert.Equal(t, user.Email, form.Email)

		err = client.CallPost("/auth/register", form, &user)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "email has exists")
	}
	{
		form := LoginForm{
			Email:    "bob@example.org",
			Password: "hello12345",
		}
		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Nil(t, err)
		assert.Equal(t, user.Email, form.Email)
		assert.Empty(t, user.Password)
		// FIXME:
		// assert.Equal(t, user.LastLoginIP, "")
	}
	{
		w := client.Get("/auth/info")
		vals := checkResponse(t, w)
		assert.Contains(t, vals, "email")
		assert.Equal(t, vals["email"], "bob@example.org")
	}
	{
		w := client.Get("/auth/logout")
		checkResponse(t, w)
	}
	{
		w := client.Get("/auth/info")
		assert.Equal(t, http.StatusForbidden, w.Code)
	}
	{
		form := LoginForm{
			Email:    "bob@hello.org",
			Password: "-",
		}
		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Contains(t, err.Error(), "user not exists")
	}
	{
		form := LoginForm{
			Email:    "bob@example.org",
			Password: "-",
		}
		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Contains(t, err.Error(), "unauthorized")
	}
	{
		form := LoginForm{
			Email:    "bob@example.org",
			Password: "hello12345",
		}
		SetValue(db, KEY_USER_ACTIVATED, "true")
		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Contains(t, err.Error(), "waiting for activation")
	}
	{
		u, _ := GetUserByEmail(db, "bob@example.org")
		err := UpdateUserFields(db, u, map[string]any{
			"Enabled": false,
		})
		assert.Nil(t, err)

		form := LoginForm{
			Email:    "bob@example.org",
			Password: "hello12345",
		}
		var user User
		err = client.CallPost("/auth/login", form, &user)
		assert.Contains(t, err.Error(), "user not allow login")
	}
}

func TestAuthToken(t *testing.T) {
	db, _, client := initTestClient(t)

	defer func() {
		db.Where("email", "bob@example.org").Delete(&User{})
	}()
	SetValue(db, KEY_USER_ACTIVATED, "no")
	CreateUser(db, "bob@example.org", "123456")

	form := LoginForm{
		Email:    "bob@example.org",
		Password: "123456",
		Remember: true,
	}
	var user User
	err := client.CallPost("/auth/login", form, &user)
	assert.Nil(t, err)
	assert.NotEmpty(t, user.AuthToken)
	{
		form := LoginForm{
			Email:     "bob@example.org",
			AuthToken: user.AuthToken,
		}
		var user User
		err = client.CallPost("/auth/login", form, &user)
		assert.Nil(t, err)
		assert.Empty(t, user.AuthToken)
	}
}

func TestAuthPassword(t *testing.T) {
	db, _, client := initTestClient(t)
	defer func() {
		db.Where("email", "bob@example.org").Delete(&User{})
	}()

	SetValue(db, KEY_USER_ACTIVATED, "no")
	CreateUser(db, "bob@example.org", "123456")

	form := LoginForm{
		Email:    "bob@example.org",
		Password: "123456",
	}
	var user User
	err := client.CallPost("/auth/login", form, &user)
	assert.Nil(t, err)
	{
		form := ChangePasswordForm{
			Password: "123456",
		}
		var r bool
		err = client.CallPost("/auth/change_password", form, &r)
		assert.Nil(t, err)
		assert.True(t, r)
	}
}
