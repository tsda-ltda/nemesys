# Metrics Config routes

All routes that interact directly with metrics configuration are under `/config/metrics`.

## Get SNMP metrics

Get all SNMP metrics.

### Details

- **Role**: Admin
- **Route URL**: `GET` `/config/metrics/snmp`
- **Parameters**:
  - "limit" Limit of metrics returned. Default is 30, max is 30, min is 0.
  - "offset" Offset for searching. Default is 0, min is 0.
- **Body**: No body.
- **Responses**:

  - 200 If succeeded. With body containing it's data in the format:

  ```js
  {
    "id": "number",
    "container-id": "number",
    "container-type": 0,
    "name": "string",
    "ident": "string",
    "descr": "string",
  }[]
  ```

## Get SNMP metrics

Get a SNMP metrics.

### Details

- **Role**: Admin
- **Route URL**: `GET` `/config/metrics/snmp/:id`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:

  - 404 If metric not found.
  - 200 If succeeded. With body containing it's data in the format:

  ```js
  {
    "base": {
        "id": "number",
        "container-id": "number",
        "container-type": 0,
        "name": "string",
        "ident": "string",
        "descr": "string",
        "data-policy-id": "number",
        "rts-pulling-interval": "number", // miliseconds
        "rts-pulling-times": "number",
        "rts-cache-duration": "number" // miliseconds
    },
    "protocol": {
        "oid": "string"
    }
  }
  ```

## Create SNMP metric

Creates a SNMP metric.

### Details

- **Role**: Admin
- **Route URL**: `POST` `/config/metrics/snmp`
- **Parameters**: No parameters.
- **Body**:

```js
{
   "base": {
        "container-id": "number",
        "name": "string",
        "ident": "string",
        "descr": "string",
        "data-policy-id": "number",
        "rts-pulling-interval": "number", // miliseconds
        "rts-pulling-times": "number",
        "rts-cache-duration": "number" // miliseconds
    },
    "protocol": {
        "oid": "string"
    }
}
```

- **Responses**:

  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 404 If container or data policy not found.
  - 200 If succeeded.

## Update SNMP metric

Updates a SNMP metric.

### Details

- **Role**: Admin
- **Route URL**: `PATCH` `/config/metrics/snmp/:id`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "base": {
        "container-id": "number",
        "name": "string",
        "ident": "string",
        "descr": "string",
        "data-policy-id": "number",
        "rts-pulling-interval": "number", // miliseconds
        "rts-pulling-times": "number",
        "rts-cache-duration": "number" // miliseconds
    },
    "protocol": {
        "oid": "string"
    }
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 404 If container or data policy or metric not found.
  - 200 If succeeded.

## Delete SNMP metric

Deletes a SNMP metric.

### Details

- **Role**: Admin
- **Route URL**: `DELETE` `/config/metrics/snmp/:id`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 404 If metric not found.
  - 200 If succeeded.
