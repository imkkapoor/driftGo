# API Contract

## Base URL
All endpoints are relative to your base URL (e.g., `http://localhost:8080`)

## Authentication
Most endpoints require authentication via session tokens. The session token should be included in requests (implementation depends on your middleware - likely in cookies or Authorization header).

## Error Response Format
All error responses follow this structure:
```json
{
  "status_code": 400,
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

### Error Codes
- `INVALID_REQUEST` - The request was invalid or malformed
- `UNAUTHORIZED` - Not authorized to perform this action
- `FORBIDDEN` - No permission to access this resource
- `NOT_FOUND` - Resource not found
- `INTERNAL_ERROR` - Unexpected server error
- `VALIDATION_ERROR` - Request validation failed
- `AUTHENTICATION_ERROR` - Authentication failed
- `INVALID_FORMAT` - Invalid request format

---

## Webhooks

### POST /webhook/stytch
**Purpose:** Receives Stytch webhook events for user lifecycle management (CREATE, UPDATE, DELETE)

**Authentication:** None (webhook signature verification required)

**Headers:**
- Stytch webhook signature headers (for verification)

**Request Body:**
```json
{
  "action": "CREATE|UPDATE|DELETE",
  "event_id": "string",
  "id": "string (stytch_user_id)",
  "source": "string",
  "timestamp": "string",
  "user": {
    "created_at": "string",
    "emails": [
      {
        "email": "string",
        "email_id": "string",
        "verified": boolean
      }
    ],
    "name": {
      "first_name": "string",
      "last_name": "string",
      "middle_name": "string"
    },
    "phone_numbers": [],
    "providers": [],
    "status": "string"
  }
}
```

**Success Response:** `200 OK`

**Error Responses:**
- `400` - Invalid request format
- `401` - Invalid webhook signature

---

## Auth Endpoints
**Note:** All auth endpoints require authentication (protected routes)

### POST /auth/create
**Purpose:** Send a magic link email to create a new account

**Authentication:** Required

**Request Body:**
```json
{
  "email": "user@example.com",
  "code_challenge": "string"
}
```

**Success Response:** `200 OK`
```json
{
  // Response from Stytch API for magic link creation
}
```

**Error Responses:**
- `400` - Invalid request format or validation error
- `500` - Internal server error

---

### POST /auth/authenticate/OAuth
**Purpose:** Authenticate user via OAuth token

**Authentication:** Required

**Request Body:**
```json
{
  "token": "oauth_token_string"
}
```

**Success Response:** `200 OK`
```json
{
  // Authentication response with session details
}
```

**Error Responses:**
- `400` - Invalid request format or validation error
- `401` - OAuth authentication failed
- `500` - Internal server error

---

### POST /auth/authenticate/magiclink
**Purpose:** Authenticate user via magic link token

**Authentication:** Required

**Request Body:**
```json
{
  "token": "magic_link_token",
  "code_verifier": "string"
}
```

**Success Response:** `200 OK`
```json
{
  // Authentication response with session details
}
```

**Error Responses:**
- `400` - Invalid request format or validation error
- `401` - Magic link authentication failed
- `500` - Internal server error

---

### POST /auth/setPassword
**Purpose:** Set password for authenticated user's account

**Authentication:** Required

**Request Body:**
```json
{
  "password": "string",
  "session_duration_minutes": 60
}
```

**Success Response:** `200 OK`
```json
{
  // Updated session response
}
```

**Error Responses:**
- `400` - Invalid request format, validation error, or failed to set password
- `500` - Internal server error

---

### POST /auth/login
**Purpose:** Log in user with email and password

**Authentication:** Required

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "string"
}
```

**Success Response:** `200 OK`
```json
{
  // Session response with authentication details
}
```

**Error Responses:**
- `400` - Invalid request format or validation error
- `401` - Invalid email or password
- `500` - Internal server error

---

### POST /auth/logout
**Purpose:** Log out current user session

**Authentication:** Required

**Request Body:** None

**Success Response:** `200 OK`
```json
{
  // Logout confirmation response
}
```

**Error Responses:**
- `500` - Internal server error

---

### POST /auth/attachOAuth
**Purpose:** Attach OAuth provider to user account

**Authentication:** Required

**Request Body:**
```json
{
  "user_id": "string",
  "provider": "google|github|etc"
}
```

**Success Response:** `200 OK`
```json
{
  // OAuth attachment response
}
```

**Error Responses:**
- `400` - Invalid request format or validation error
- `500` - Internal server error

---

### POST /auth/extendSession
**Purpose:** Extend current user session duration

**Authentication:** Required

**Request Body:**
```json
{
  "session_duration_minutes": 120
}
```

**Success Response:** `200 OK`
```json
{
  // Updated session response with new expiration
}
```

**Error Responses:**
- `400` - Invalid request format or validation error
- `500` - Internal server error

---

## Link Endpoints (Plaid Integration)
**Note:** All link endpoints require authentication (protected routes)

### POST /link/create
**Purpose:** Create a Plaid Link token to initialize the Plaid Link flow

**Authentication:** Required

**Request Body:** None

**Success Response:** `200 OK`
```json
{
  "link_token": "string",
  "expiration": "2026-01-28T12:00:00Z",
  "request_id": "string"
}
```

**Error Responses:**
- `500` - Failed to create link token

---

### POST /link/exchange
**Purpose:** Exchange Plaid public token for access token and save to database

**Authentication:** Required

**Request Body:**
```json
{
  "public_token": "string"
}
```

**Success Response:** `200 OK` (no response body)

**Error Responses:**
- `400` - Invalid request format or validation error (missing public_token)
- `500` - Failed to exchange token or save to database

---

### POST /link/createStripeProcessorToken
**Purpose:** Create Stripe processor token for linked bank account

**Authentication:** Required

**Request Body:**
```json
{
  "account_id": "string"
}
```

**Success Response:** `200 OK` (no response body)

**Error Responses:**
- `400` - Invalid request format or validation error (missing account_id)
- `500` - Failed to get access token or create Stripe processor token

---

## Transfer Endpoints (Plaid Transfer)
**Note:** All transfer endpoints require authentication (protected routes). Transfers are US-only and use Plaid Ledger for two-leg fund movement.

### POST /transfer/create
**Purpose:** Initiate a transfer from the authenticated user's linked bank account to another user's linked bank account

**Authentication:** Required

**Request Body:**
```json
{
  "sender_account_id": "string (Plaid account ID of sender's bank account)",
  "receiver_user_id": 123,
  "receiver_account_id": "string (Plaid account ID of receiver's bank account)",
  "amount": "10.00",
  "description": "string (optional)"
}
```

**Success Response:** `201 Created`
```json
{
  "id": 1,
  "sender_user_id": 1,
  "receiver_user_id": 2,
  "sender_account_id": "string",
  "receiver_account_id": "string",
  "amount": "10.00",
  "currency": "USD",
  "status": "debit_pending",
  "description": "string",
  "created_at": "2026-02-21T12:00:00Z",
  "updated_at": "2026-02-21T12:00:00Z"
}
```

**Error Responses:**
- `400` - Invalid request format, validation error, self-transfer attempt, or invalid amount
- `500` - Internal server error (Plaid authorization or transfer creation failed)

---

### GET /transfer/{id}
**Purpose:** Get details of a specific transfer. Only accessible to the sender or receiver.

**Authentication:** Required

**Path Parameters:**
- `id` - Transfer ID (integer)

**Success Response:** `200 OK`
```json
{
  "id": 1,
  "sender_user_id": 1,
  "receiver_user_id": 2,
  "sender_account_id": "string",
  "receiver_account_id": "string",
  "amount": "10.00",
  "currency": "USD",
  "status": "pending|debit_authorized|debit_pending|debit_settled|credit_authorized|credit_pending|completed|failed|returned|cancelled",
  "description": "string",
  "created_at": "2026-02-21T12:00:00Z",
  "updated_at": "2026-02-21T12:00:00Z",
  "completed_at": "2026-02-21T12:00:00Z"
}
```

**Error Responses:**
- `400` - Invalid transfer ID
- `404` - Transfer not found (or not authorized to view)
- `500` - Internal server error

---

### GET /transfer/list
**Purpose:** List all transfers for the authenticated user (both sent and received)

**Authentication:** Required

**Success Response:** `200 OK`
```json
{
  "transfers": [
    {
      "id": 1,
      "sender_user_id": 1,
      "receiver_user_id": 2,
      "sender_account_id": "string",
      "receiver_account_id": "string",
      "amount": "10.00",
      "currency": "USD",
      "status": "completed",
      "description": "string",
      "created_at": "2026-02-21T12:00:00Z",
      "updated_at": "2026-02-21T12:00:00Z",
      "completed_at": "2026-02-21T12:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `500` - Internal server error

---

## Webhooks

### POST /webhook/plaid
**Purpose:** Receives Plaid webhook events for transfer status updates

**Authentication:** None (Plaid webhook verification)

**Request Body:**
```json
{
  "webhook_type": "TRANSFER",
  "webhook_code": "TRANSFER_EVENTS_UPDATE",
  "environment": "sandbox|production"
}
```

**Success Response:** `200 OK`

**Error Responses:**
- `400` - Invalid request format
- `500` - Failed to process transfer events

---

## Common Data Models

### Transfer Status Flow
```
pending -> debit_authorized -> debit_pending -> debit_settled -> credit_authorized -> credit_pending -> completed
                                    |                                     |
                                    v                                     v
                                  failed                                failed
                                  returned                              returned
                                  cancelled                             cancelled
```

### Session Object
```json
{
  "user_id": "string",
  "token": "string",
  "refresh_token": "string",
  "expires_at": "2026-01-28T12:00:00Z"
}
```

### User Object
```json
{
  "id": "string",
  "email": "string",
  "created_at": "2026-01-28T12:00:00Z",
  "updated_at": "2026-01-28T12:00:00Z"
}
```

---

## Notes for Frontend Implementation

1. **Request Timeout:** All requests have a 60-second timeout
2. **Content-Type:** All requests should use `application/json` content type
3. **Session Management:** Session tokens are managed via middleware - check your implementation for how tokens are sent (cookies vs headers)
4. **Validation:** All required fields must be provided; missing fields will return validation errors
5. **Error Handling:** Always handle error responses by checking the `status_code` and `code` fields
6. **Plaid Flow:** 
   - Call `/link/create` to get link token
   - Use link token to initialize Plaid Link UI
   - After user completes Plaid Link, get public token
   - Call `/link/exchange` with public token
   - Optionally call `/link/createStripeProcessorToken` for Stripe integration
7. **OAuth Flow:** Use `/auth/authenticate/OAuth` for OAuth login or `/auth/attachOAuth` to attach OAuth to existing account
8. **Magic Link Flow:** 
   - Call `/auth/create` to send magic link email
   - User clicks link and gets token
   - Call `/auth/authenticate/magiclink` with token to complete authentication
9. **Transfer Flow:**
   - Both sender and receiver must have linked bank accounts (via Plaid Link with transfer product enabled)
   - Sender calls `POST /transfer/create` with receiver's user ID and account IDs
   - The debit leg (sender -> Plaid Ledger) is initiated immediately
   - The credit leg (Plaid Ledger -> receiver) is triggered automatically when the debit settles (via webhook)
   - Use `GET /transfer/{id}` or `GET /transfer/list` to check transfer status
   - Transfers are US-only, currency is always USD
   - Transfer lifecycle: `pending` -> `debit_pending` -> `debit_settled` -> `credit_pending` -> `completed`
