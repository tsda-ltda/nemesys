# Teams Config routes

All routes that interact directly with teams configuration and visualization are here.

## Create

Creates a new team.

### Details

- **Role**: Manager
- **Route URL**: `POST` `/config/teams`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "name": "string",
  "ident": "string",
  "descr": "string"
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 400 If ident can be parsed to number.
  - 400 If ident is already in use.
  - 200 If succeeded.

## Update

Updates a team by id or ident.

### Details

- **Role**: Manager
- **Route URL**: `PATCH` `/config/teams/:(ident or id)`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "name": "string",
  "ident": "string",
  "descr": "string"
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 400 If ident in use.
  - 400 If ident can be parsed to number.
  - 404 If team not found.
  - 200 If succeeded.

## Delete

Deletes a team by id or ident.

### Details

- **Role**: Manager
- **Route URL**: `DELETE` `/config/teams/:(ident or id)`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 404 If team not found.
  - 200 If succeeded.

## Get teams

Get a list of teams.

### Details

- **Role**: Manager
- **Route URL**: `GET` `/config/teams`
- **Parameters**:
  - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
  - "offset" Offset for searching. Default is 0, min is 0.
- **Body**: No body.
- **Responses**:
  - 200 If succeeded. With body containing it's data in the format:

```js
{
  "id": "number",
  "name": "string",
  "ident": "string",
  "descr": "string"
}[]
```

## Get team

Get a team by id or ident.

### Details

- **Role**: Manager
- **Route URL**: `GET` `/config/teams/:(ident or id)`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 404 If team not found.
  - 200 If succeeded. With body containing it's data in the format:

```js
{
  "id": "number",
  "name": "string",
  "ident": "number",
  "descr": "string"
}
```

## Add user

Add a user to the team.

### Details

- **Role**: Manager
- **Route URL**: `POST` `/config/teams/:(ident or id)/users`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "users-ids": "number"
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If invalid body fields.
  - 400 If user is already a member.
  - 404 If team not found.
  - 200 If succeeded.

## Remove user

Remove a user from the team.

### Details

- **Role**: Manager
- **Route URL**: `DEL` `/config/teams/:(ident or id)/users/:userId`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If invalid user id.
  - 404 If team not found or user was not a member.
  - 204 If succeeded.

## Get user teams

Return all teams that the user is member.

### Details

- **Role**: Viewer
- **Route URL**: `GET` `/teams`
- **Parameters**:
  - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
  - "offset" Offset for searching. Default is 0, min is 0.
- **Body**: No body.
- **Responses**:
  - 200 If succeeded. With body containing it's data in the format:

```js
{
  "id": "number",
  "name": "string",
  "ident": "number",
  "descr": "string"
}
```
