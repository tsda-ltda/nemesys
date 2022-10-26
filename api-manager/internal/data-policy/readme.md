# Data policies Config routes

All routes that interact directly with data policies configuration are under `/config/data-policies`.

## Get all

Get all data policies.

### Details

- **Role**: Master
- **Route URL**: `GET` `/config/data-policies`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:

  - 200 If succeeded. With body containing it's data in the format:

  ```js
  {
    "id": "number",
    "descr": "string",
    "use-aggregation": "boolean",
    "retention": "number",
    "aggregation-retention": "number",
    "aggregation-interval": "number"
  }[]
  ```

## Create

Creates a data policy.

### Details

- **Role**: Master
- **Route URL**: `POST` `/config/data-policies`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "descr": "string",
  "use-aggregation": "boolean",
  "retention": "number",
  "aggregation-retention": "number",
  "aggregation-interval": "number"
}
```

- **Responses**:

  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 200 If succeeded.

## Update

Updates a data policy by id.

### Details

- **Role**: Master
- **Route URL**: `PATCH` `/config/data-policies/:id`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "descr": "string",
  "use-aggregation": "boolean",
  "retention": "number",
  "aggregation-retention": "number",
  "aggregation-interval": "number"
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 404 If data policy not found.
  - 200 If succeeded.

## Delete

Deletes a data policy by id.

### Details

- **Role**: Master
- **Route URL**: `DELETE` `/config/data-policy/:id`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 404 If data policy not found.
  - 204 If succeeded.
