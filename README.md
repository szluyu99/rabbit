# rabbit

Rabbit is a golang library for simplifying backend develop.

# Features

## Dynamic Handlers & Dynamic Gorm Functions

```go
func HandleGet[T any](c *gin.Context, db *gorm.DB, onRender onRenderFunc[T])
func HandleDelete[T any](c *gin.Context, db *gorm.DB, onDelete onDeleteFunc[T]) 
func HandleCreate[T any](c *gin.Context, db *gorm.DB, onCreate onCreateFunc[T])
func HandleEdit[T any](c *gin.Context, db *gorm.DB, editables []string, onUpdate onUpdateFunc[T])
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

### With authentication module

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

### With authorization module

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

### With middleware module

```go
ar := r.Group("/api").Use(
  rabbit.WithAuthentication(), 
  rabbit.WithAuthorization("/api"),
)

rabbit.RegisterAuthorizationHandlers(db, ar)
```
