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

Updates a team by id.

### Details

- **Role**: Manager
- **Route URL**: `PATCH` `/config/teams/:id`
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

Deletes a team by id.

### Details

- **Role**: Manager
- **Route URL**: `DELETE` `/config/teams/:id`
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

Get a team by id.

### Details

- **Role**: Manager
- **Route URL**: `GET` `/config/teams/:id`
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

## Add member

Add a member to the team.

### Details

- **Role**: Manager
- **Route URL**: `POST` `/config/teams/:id/members`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "users-id": "number"
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If invalid body fields.
  - 400 If user is already a member.
  - 404 If team not found.
  - 200 If succeeded.

## Remove member

Remove a member from the team.

### Details

- **Role**: Manager
- **Route URL**: `DEL` `/config/teams/:id/members/:userId`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If invalid user id.
  - 404 If team not found or user was not a member.
  - 204 If succeeded.

## Get members

Get all members.

### Details

- **Role**: Manager
- **Route URL**: `GET` `/config/teams/:id/members`
- **Parameters**:
  - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
  - "offset" Offset for searching. Default is 0, min is 0.
- **Body**: No body.
- **Responses**:
  - 400 If invalid team id.
  - 404 If team not found.
  - 200 If succeeded with body containing it's data in the format:

```js
  {
    "id": "number",
    "name": "string",
    "username": "string"
  }[]
```

## Add context

Add a context to the team.

### Details

- **Role**: Manager
- **Route URL**: `POST` `/config/teams/:id/contexts`
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
  - 400 If invalid body fields.
  - 400 If ident already exists.
  - 404 If team not found.
  - 200 If succeeded.

## Remove Context

Remove a context from the team.

### Details

- **Role**: Manager
- **Route URL**: `DEL` `/config/teams/:id/contexts/:contextId`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If invalid team or context id.
  - 404 If team or context not found.
  - 204 If succeeded.

## Get contexts

Get all team's contexts.

### Details

- **Role**: Manager
- **Route URL**: `GET` `/config/teams/:id/contexts`
- **Parameters**:
  - "limit" Limit of teams returned. Default is 30, max is 30, min is 0.
  - "offset" Offset for searching. Default is 0, min is 0.
- **Body**: No body.
- **Responses**:
  - 400 If invalid team id.
  - 404 If team not found.
  - 200 If succeeded with body containing it's data in the format:

```js
  {
    "id": "number",
    "name": "string",
    "ident": "string",
    "descr": "string"
  }[]
```

## Get user's teams

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
