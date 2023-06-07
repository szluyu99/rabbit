package rabbit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// check response code and unmarshal response body to map[string]any
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

		// register user
		var user User
		err := client.CallPost("/auth/register", form, &user)
		assert.Nil(t, err)
		assert.Equal(t, user.Email, form.Email)

		// register user with same email, should fail
		err = client.CallPost("/auth/register", form, &user)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "email has exists")
	}
	{
		// login success
		form := LoginForm{
			Email:    "bob@example.org",
			Password: "hello12345",
		}

		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Nil(t, err)
		assert.Equal(t, user.Email, form.Email)
		assert.Empty(t, user.Password)
		assert.Equal(t, "127.0.0.1", user.LastLoginIP)

		// get user info after login
		w := client.Get("/auth/info")
		vals := checkResponse(t, w)
		assert.Contains(t, vals, "email")
		assert.Equal(t, vals["email"], "bob@example.org")

		// logout
		w = client.Get("/auth/logout")
		checkResponse(t, w)

		// get user info after logout, should fail
		w = client.Get("/auth/info")
		assert.Equal(t, http.StatusForbidden, w.Code)
	}
	{
		// not exist user, should fail
		form := LoginForm{
			Email:    "bob@hello.org",
			Password: "-",
		}
		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Contains(t, err.Error(), "user not exists")
	}
	{
		// wrong password, should fail
		form := LoginForm{
			Email:    "bob@example.org",
			Password: "-",
		}
		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Contains(t, err.Error(), "unauthorized")
	}
	{
		// set need activate env
		form := LoginForm{
			Email:    "bob@example.org",
			Password: "hello12345",
		}
		SetValue(db, KEY_USER_NEED_ACTIVATE, "true")
		var user User
		err := client.CallPost("/auth/login", form, &user)
		assert.Contains(t, err.Error(), "waiting for activation")
	}
	{
		// user not enabled, should fail
		u, _ := GetUserByEmail(db, "bob@example.org")
		err := UpdateFields(db, u, map[string]any{
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

	// login with token
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

	SetValue(db, KEY_USER_NEED_ACTIVATE, "no")
	CreateUser(db, "bob@example.org", "123456")

	// login with password
	form := LoginForm{
		Email:    "bob@example.org",
		Password: "123456",
	}
	var user User
	err := client.CallPost("/auth/login", form, &user)
	assert.Nil(t, err)

	// change password
	{
		form := ChangePasswordForm{
			Password: "654321",
		}
		var r bool
		err = client.CallPost("/auth/change_password", form, &r)
		assert.Nil(t, err)
		assert.True(t, r)
	}

	// login with new password
	{
		form := LoginForm{
			Email:    "bob@example.org",
			Password: "654321",
		}

		err := client.CallPost("/auth/login", form, nil)
		assert.Nil(t, err)
	}
}
