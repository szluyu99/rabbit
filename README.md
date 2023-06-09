# rabbit

Rabbit is a golang library for simplifying backend develop.

⭐ Open source project using Rabbit, which can be a reference for an example project:
- [https://github.com/szluyu99/rabbit-admin](https://github.com/szluyu99/rabbit-admin)

# Features

> All features have corresponding unit tests, which is convenient for developers to learn and use.

## Dynamic Handlers & Dynamic Gorm Functions

> Reference: [./handler_common_test.go](https://github.com/szluyu99/rabbit/blob/main/handler_common_test.go)

```go

```go
func HandleGet[T any](c *gin.Context, db *gorm.DB, onRender onRenderFunc[T])
func HandleDelete[T any](c *gin.Context, db *gorm.DB, onDelete onDeleteFunc[T]) 
func HandleCreate[T any](c *gin.Context, db *gorm.DB, onCreate onCreateFunc[T])
func HandleEdit[T any](c *gin.Context, db *gorm.DB, editables []string, onUpdate onUpdateFunc[T])
func HandleQuery[T any](c *gin.Context, db *gorm.DB, ctx *QueryOption)
```

```go
func ExecuteGet[T any, V Key](db *gorm.DB, key V) (*T, error)
func ExecuteEdit[T any, V Key](db *gorm.DB, key V, vals map[string]any) (*T, error)
func ExecuteQuery[T any](db *gorm.DB, form QueryForm) (items []T, count int, err error)
```

About how to use: Please refer to the corresponding unit tests.

## Integration Web Objects - Generate RESTful API

Reference [https://github.com/restsend/gormpher](https://github.com/restsend/gormpher)

## Env Config

### Load environment variables

Functions:
  
```go
func GetEnv(key string) string
func LookupEnv(key string) (string, bool)
```

Examples: 

```bash
# .env
xx
EXIST_ENV=100	
```

```bash
# run with env 
EXIST_ENV=100 go run .
```

```go
rabbit.GetEnv("EXIST_ENV") // 100
rabbit.LookupEnv("EXIST_ENV") // 100, true
```

### Load config from DB

Functions:

```go
func CheckValue(db *gorm.DB, key, default_value string)
func SetValue(db *gorm.DB, key, value string)
func GetValue(db *gorm.DB, key string) string
func GetIntValue(db *gorm.DB, key string, default_value int) int
func GetBoolValue(db *gorm.DB, key string) bool
```

Examples:

```go
db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
db.AutoMigrate(&rabbit.Config{})

rabbit.SetValue(db, "test_key", "test_value")
value := rabbit.GetValue(db, "test_key") // test_value

rabbit.CheckValue(db, "check_key", "default_value")
value = rabbit.GetValue(db, "check_key") // default_value

rabbit.SetValue(db, "int_key", "42")
intValue := rabbit.GetIntValue(db, "int_key", -1) // 42

rabbit.SetValue(db, "bool_key", "true")
boolValue := rabbit.GetBoolValue(db, "bool_key") // true
```

## Built-in Handlers

### Permission models

```go
User <-UserRole-> Role
Role <-RolePermission-> Permission
User <-GroupMember-> Group

User
- ID
- Email
- Password
- ...

// for association
UserRole
- UserID
- RoleID

Role
- Name
- Label

// for association 
RolePermission
- RoleID
- PermissionID

Permission
- Name
- Uri
- Method
- Anonymous
- ParentID  // for tree struct
- Children  // for tree struct

Group
- Name

// for association
GroupMember
- UserID
- GroupID
```

### Authentication handlers

```go
RegisterAuthenticationHandlers("/auth", db, r)
```
 
```
GET    /auth/info
POST   /auth/login
POST   /auth/register
GET    /auth/logout
POST   /auth/change_password
```

### Authorization handlers

```go
rabbit.RegisterAuthorizationHandlers(db, r.Group("/api"))
```

```
PUT    /api/role
PATCH  /api/role/:key
DELETE /api/role/:key
PUT    /api/permission
PATCH  /api/permission/:key
DELETE /api/permission/:key
```

### Middleware

```go
ar := r.Group("/api").Use(
  rabbit.WithAuthentication(), 
  rabbit.WithAuthorization("/api"),
)

rabbit.RegisterAuthorizationHandlers(db, ar)
```

## Unit Tests Utils

> Reference: [tests_test.go](https://github.com/szluyu99/rabbit/blob/main/tests_test.go)

Example:

```go
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
```

Example Test:

```go
c := rabbit.NewTestClient(r)
```

```go
// CallGet
{
  var result bool
  err := c.CallGet("/ping", nil, &result)
  assert.Nil(t, err)
  assert.True(t, result)
}
// Get
{
  w := c.Get("/ping")
  assert.Equal(t, http.StatusOK, w.Code)
  assert.Equal(t, "true", w.Body.String())
}
// Native
{
  req := httptest.NewRequest(http.MethodGet, "/ping", nil)
  w := httptest.NewRecorder()
  r.ServeHTTP(w, req)

  assert.Equal(t, http.StatusOK, w.Code)
  assert.Equal(t, "true", w.Body.String())
}
```

```go
// CallPost
{
  var result map[string]any
  err := c.CallPost("/pong", user{Name: "test", Age: 11}, &result)

  assert.Nil(t, err)
  assert.Equal(t, "test", result["name"])
  assert.Equal(t, float64(11), result["age"])
}
// Post
{
  b, _ := json.Marshal(user{Name: "test", Age: 11})
  w := c.Post("/pong", b)

  assert.Equal(t, http.StatusOK, w.Code)
  assert.Equal(t, `{"age":11,"name":"test"}`, w.Body.String())
}
// Native
{
  b, _ := json.Marshal(user{Name: "test", Age: 11})
  req := httptest.NewRequest(http.MethodPost, "/pong", bytes.NewReader(b))
  w := httptest.NewRecorder()
  r.ServeHTTP(w, req)

  assert.Equal(t, http.StatusOK, w.Code)
  assert.Equal(t, `{"age":11,"name":"test"}`, w.Body.String())
}
```

# Acknowledgement Project:
- [https://github.com/restsend/carrot](https://github.com/restsend/carrot)
- [https://github.com/restsend/gormpher](https://github.com/restsend/gormpher)
