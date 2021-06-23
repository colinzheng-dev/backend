#Connector

##Setup

### Creating the database

Execute the following script on the database. Remember that the password credentials must be set on deployment (Kubernetes).

```sql
CREATE DATABASE vb_shipping;

CREATE USER shipping_service PASSWORD 'abc123';
GRANT ALL ON DATABASE vb_shipping TO shipping_service;
CREATE ROLE vb_shipping;
GRANT vb_shipping TO postgres;
GRANT vb_shipping TO shipping_service;
```
