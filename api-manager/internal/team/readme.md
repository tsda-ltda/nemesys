# Teams routes

All routes that interact directly with teams are grouped together under `\teams`.

## Create

Creates a new team.

### Details

- **Route URL**: `POST` `/teams`
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

- **Route URL**: `PATCH` `/teams/:(ident or id)`
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

- **Route URL**: `DELETE` `/teams/:(ident or id)`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 404 If team not found.
  - 200 If succeeded.

## Get teams

Get a list of teams.

### Details

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
  "ident": "string",
  "descr": "string"
}[]
```

## Get team

Get a team by id or ident.

### Details

- **Route URL**: `GET` `/teams/:(ident or id)`
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
  "users-ids": "number[]"
}
```

## Update team's users

Update users that are part of the team.

### Details

- **Route URL**: `PATCH` `/teams/:(ident or id)/users`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "users-ids": "number[]"
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If have duplicated users.
  - 404 If team not found.
  - 200 If succeeded.
