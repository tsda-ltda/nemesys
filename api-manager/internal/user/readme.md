# Users routes

All routes that interact directly with users configuration are grouped together under `/config/users`.

## Create

Creates a new user.

### Details

- **Role**: Admin
- **Route URL**: `POST` `/config/users`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "role": "number",
  "name": "string",
  "password": "string",
  "email": "string",
  "username": "string"
}
```

- **Responses**:
  - 400 If invalid id.
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 400 If username is already in use.
  - 400 If email is already in use.
  - 200 If succeeded.

## Update

Updates a user by id.

### Details

- **Role**: Admin
- **Route URL**: `PATCH` `/config/users/:id`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "role": "number",
  "name": "string",
  "password": "string",
  "email": "string",
  "username": "string"
}
```

- **Responses**:
  - 400 If invalid id.
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 400 If username is already in use.
  - 400 If email is already in use.
  - 404 If user not found.
  - 200 If succeeded.

## Delete

Deletes a user by id.

### Details

- **Role**: Admin
- **Route URL**: `DELETE` `/config/users/:id`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If invalid id.
  - 404 If user not found.
  - 200 If succeeded.

## Get users

Get a list of users.

### Details

- **Role**: Manager
- **Route URL**: `GET` `/config/users`
- **Parameters**:
  - "limit" Limit of users returned. Default is 30, max is 30, min is 0.
  - "offset" Offset for searching. Default is 0, min is 0.
- **Body**: No body.
- **Responses**:
  - 200 If succeeded. With body containing it's data in the format:

```js
{
  "id": "number",
  "username": "string",
  "name": "string"
}[]
```

## Get user

Get a user by id.

### Details

- **Role**: Admin or Onwer
- **Route URL**: `GET` `/config/users/:id`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If invalid id.
  - 404 If user not found.
  - 200 If succeeded. With body containing it's data in the format:

```js
{
  "id": "number",
  "username": "string",
  "name": "string",
  "role": "number",
  "email": "string"
}
```
