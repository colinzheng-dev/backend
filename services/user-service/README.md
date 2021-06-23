# User service

## External API routes relating to users

```
GET /users?q={search-term}&page={n}&per_page={n}

GET /me
PUT /me
DELETE /me
POST /me/api-key
DELETE /me/api-key

GET /user/{id}
PUT /user/{id}
DELETE /user/{id}
POST /user/{id}/api-key
DELETE /user/{id}/api-key
```

## Inter-service API routes relating to users

```
POST /login  {"email": "user@example.com"}
```

## Headers from API gateway

```
X-Auth-User-Id: user_id
X-Auth-Method: none/session/api-key
X-Auth-Is-Admin: true/false
```

These are extracted into the request context using a middleware in the
chassis called `AuthCtx`, which should be reusable everywhere (don't
know how it will work for gRPC, since that uses persistent HTTP/2
connections and I don't know what happens to the HTTP headers then,
but something similar can probably be done).
