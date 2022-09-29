# Users Auth routes

All routes that interact directly with users authentication are grouped are here.

## Login

Login into an user account.

### Details

- **Role**: None.
- **Route URL**: `POST` `/login`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "username": "string"
  "password": "string"
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 401 If email or password are wrong.
  - 200 If succeeded.

## Logout

Login into an user account.

### Details

- **Role**: None.
- **Route URL**: `POST` `/logout`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If no session was running.
  - 200 If succeeded.

## Force logout

Force a user to logout.

### Details

- **Role**: Admin.
- **Route URL**: `POST` `/users/:id/logout`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If no session was running.
  - 403 If user role is superior to request session role.
  - 200 If succeeded.
