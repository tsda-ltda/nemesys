# Authentication

All routes that interact directly with user authentication are here.

## Login

Login into an user account.

### Details

- **Role**: None.
- **Route URL**: `POST` `/login`
- **Parameters**: No parameters.
- **Body**:

```js
{
  "username": "string" // min: 2, max: 50
  "password": "string" // min: 5, max: 50
}
```

- **Responses**:
  - 400 If invalid body.
  - 400 If json fields are invalid.
  - 401 If email or password is wrong.
  - 200 If succeeded.

## Logout

Logout of an user account.

### Details

- **Role**: Viewer.
- **Route URL**: `POST` `/logout`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 200 If succeeded.

## Force logout

Force user's session logout.

### Details

- **Role**: Admin.
- **Route URL**: `POST` `/users/:userId/logout`
- **Parameters**: No parameters.
- **Body**: No body.
- **Responses**:
  - 400 If no session exists.
  - 403 If target's role is superior then the user who resquested.
  - 200 If succeeded.
