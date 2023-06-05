# rabbit

Rabbit is a golang library for simplifying backend develop.

# Features

## Dynamic handlers

## Web Objects

## Built-in modules

### With authentication module
 
```
GET    /auth/info
POST   /auth/login
POST   /auth/register
GET    /auth/logout
POST   /auth/change_password
```

### With authorization module

```
PUT    /auth/role
PATCH  /auth/role/:id
DELETE /auth/role/:id
PUT    /auth/permission
PATCH  /auth/permission/:key
DELETE /auth/permission/:id
```

### With config module

```
PATCH  /auth/config/:key
DELETE /auth/config/:key
POST   /auth/config
```
