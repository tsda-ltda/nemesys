# Containers Config routes

All routes that interact directly with containers configuration are under `/config/containers`.

## Get SNMP containers

Get all SNMP containers.

### Details

- **Role**: Admin
- **Route URL**: `GET` `/config/containers/snmp`
- **Parameters**:
  - "limit" Limit of containers returned. Default is 30, max is 30, min is 0.
  - "offset" Offset for searching. Default is 0, min is 0.
- **Body**: No body.
- **Responses**:

  - 200 If succeeded. With body containing it's data in the format:

  ```js
  {
    "id": "number",
    "name": "string",
    "ident": "string",
    "descr": "string",
    "type": 0
  }[]
  ```

## Get SNMP container

Get a SNMP container.

### Details

- **Role**: Admin
- **Route URL**: `GET` `/config/containers/snmp/:id`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:

  - 404 If containers not found.
  - 200 If succeeded. With body containing it's data in the format:

  ```js
  {
    "base": {
        "id": "number",
        "name": "number",
        "ident": "string",
        "descr": "string",
        "type": 0
    },
    "protocol": {
        "container-id": "number",
        "cache-duration": "number", // miliseconds
        "target": "string",
        "port": "number",
        "transport": "string",
        "community": "string",
        "timeout": "number", // miliseconds
        "retries": "number",
        "msg-flags": "number",
        "version": "number",
        "max-oids": "number"
    }
  }
  ```

## Create SNMP container

Creates a SNMP container.

### Details

- **Role**: Admin
- **Route URL**: `POST` `/config/containers/snmp`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "base": {
      "name": "string",
      "ident": "string",
      "descr": "string"
  },
  "protocol": {
      "version": "number", // 0 = version1, 1 = version2c, 3 = version3
      "cache-duration": "number", // miliseconds
      "timeout": "number", // miliseconds
      "target": "string",
      "port": "number",
      "transport": "string",
      "community": "string",
      "retries": "string",
      "msg-flags": "number",
      "max-oids": "number"
  }
}
```

- **Responses**:

  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 200 If succeeded.

## Update SNMP container

Updates a SNMP container.

### Details

- **Role**: Admin
- **Route URL**: `PATCH` `/config/containers/snmp/:id`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "base": {
      "name": "string",
      "ident": "string",
      "descr": "string"
  },
  "protocol": {
      "version": "number", // 0 = version1, 1 = version2c, 3 = version3
      "cache-duration": "number", // miliseconds
      "timeout": "number", // miliseconds
      "target": "string",
      "port": "number",
      "transport": "string",
      "community": "string",
      "retries": "string",
      "msg-flags": "number",
      "max-oids": "number"
  }
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 404 If container not found.
  - 200 If succeeded.

## Delete SNMP container

Deletes a SNMP container.

### Details

- **Role**: Admin
- **Route URL**: `DELETE` `/config/containers/snmp/:id`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 404 If container not found.
  - 204 If succeeded.
