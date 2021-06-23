# Questions

 - Having the `chassis` stuff in a separate module is convenient, but
   having to use a `replace` in the `go.mod` to pick up local changes
   isn't the best, since it loses versionability of the chassis for
   each service. Need to think about this a bit.

# Todos

## Checks

 - Make sure the order of the authentication and CORS middleware ends
   up in the right order. CORS preflight request must not require
   authentication, i.e. either an API token or a session token, and
   should be handled before any authentication checking.
 - Think about whether and when `Access-Control-Allow-Credentials`
   header is needed in CORS preflight responses.
 - The blob service needs permissions checking, since it currently
   doesn't have any at all, so anyone can upload files to our blob
   store.
 - Make sure there's no middleware in front of healthcheck routes
   everywhere.
 - Come up with a better way to organise routes and middleware
   everywhere.

# Future features

## Users

 - Free text (but limited length) user status.
 - User tags.
 - User points.
 - User links to items: discovered, wishlist, owned, etc.
