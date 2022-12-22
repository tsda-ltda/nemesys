# Users

All routes that interact directly with users are here.

## Create

Creates a new user.

### Details

- **Role**: Admin
- **Route URL**: `POST` `/users`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "role": "number", // min: 1, max: 4
  "first-name": "string", // min: 1, max: 50
  "last-name": "string", // min: 1, max: 50
  "password": "string", // min: 5, max: 50
  "email": "string", // min: 1, max: 255
  "username": "string" // min: 2, max: 50
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 400 If username or email is already in use.
  - 200 If succeeded.

## Update

Updates a user by id.

### Details

- **Role**: Admin
- **Route URL**: `PATCH` `/users/:id`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "role": "number", // min: 1, max: 4
  "first-name": "string", // min: 1, max: 50
  "last-name": "string", // min: 1, max: 50
  "password": "string", // min: 5, max: 50
  "email": "string", // min: 1, max: 255
  "username": "string" // min: 2, max: 50
}
```

- **Responses**:
  - 400 If invalid id.
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 400 If username or email is already in use.
  - 404 If user not found.
  - 200 If succeeded.

## Delete

Deletes a user by id.

### Details

- **Role**: Admin
- **Route URL**: `DELETE` `/users/:id`
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
- **Route URL**: `GET` `/users`
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
  "first-name": "string",
  "last-name": "string",
  "role": "number",
  "email": "string"
}[]
```

## Get user

Get a user by id.

### Details

- **Role**: Admin or Viwer (needs to be logged as the user)
- **Route URL**: `GET` `/users/:id`
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
  "first-name": "string",
  "last-name": "string",
  "role": "number",
  "email": "string"
}
```

## Get session

Get the user of the current session.

### Details

- **Role**: Viewer
- **Route URL**: `GET` `/session`
- **Body**: No body.
- **Responses**:
  - 200 If succeeded. With body containing it's data in the format:

```js
{
  "id": "number",
  "username": "string",
  "first-name": "string",
  "last-name": "string",
  "role": "number",
  "email": "string"
}
```
