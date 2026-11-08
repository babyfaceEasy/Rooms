# Admin Frontend Integration Guide

## Quick Start

This guide helps frontend engineers integrate the admin API into their application.

### 1. Authentication Flow

#### Login
```typescript
// Login endpoint
const response = await fetch('/api/v1/admin/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'admin@example.com',
    password: 'securepassword'
  })
});

const { access_token, refresh_token } = await response.json();

// Store tokens securely (HTTP-only cookies recommended)
localStorage.setItem('adminAccessToken', access_token);
localStorage.setItem('adminRefreshToken', refresh_token);
```

#### Making Authenticated Requests
```typescript
function getAuthHeaders() {
  const token = localStorage.getItem('adminAccessToken');
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  };
}

const response = await fetch('/api/v1/admin/customers', {
  method: 'GET',
  headers: getAuthHeaders()
});
```

#### Handling Token Expiration
```typescript
async function fetchWithTokenRefresh(url, options = {}) {
  let response = await fetch(url, {
    ...options,
    headers: getAuthHeaders()
  });

  // If token expired, refresh it
  if (response.status === 401) {
    const refreshToken = localStorage.getItem('adminRefreshToken');
    const refreshResponse = await fetch('/api/v1/admin/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken })
    });

    if (refreshResponse.ok) {
      const { access_token } = await refreshResponse.json();
      localStorage.setItem('adminAccessToken', access_token);
      
      // Retry original request with new token
      response = await fetch(url, {
        ...options,
        headers: getAuthHeaders()
      });
    } else {
      // Refresh failed, redirect to login
      redirectToLogin();
    }
  }

  return response;
}
```

### 2. Admin Setup Flow

New admins are created with a setup token that must be used to set their initial password.

#### Creating an Admin (Super Admin Only)
```typescript
async function createAdmin(email, name, role = 'admin') {
  const response = await fetchWithTokenRefresh('/api/v1/admin/admins', {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ email, name, role })
  });

  const data = await response.json();
  
  if (response.ok) {
    // data contains:
    // - admin: the new admin object
    // - setup_token: send this to the admin via secure channel
    
    console.log('Setup token:', data.setup_token);
    console.log('Send this token to', email, 'via secure channel');
    
    return data;
  } else {
    throw new Error(data.error);
  }
}
```

#### Admin Initial Setup (First Time Login)
When a new admin receives their setup token, they use this endpoint to set their password:

```typescript
async function adminInitialSetup(email, setupToken, password) {
  const response = await fetch('/api/v1/admin/auth/initial-setup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email,
      setup_token: setupToken,
      password
    })
  });

  if (response.ok) {
    const { access_token, refresh_token } = await response.json();
    localStorage.setItem('adminAccessToken', access_token);
    localStorage.setItem('adminRefreshToken', refresh_token);
    return true;
  }
  
  throw new Error('Setup failed');
}
```

### 3. Common Tasks

#### List Customers
```typescript
async function listCustomers(limit = 20, offset = 0) {
  const response = await fetchWithTokenRefresh(
    `/api/v1/admin/customers?limit=${limit}&offset=${offset}`
  );
  
  if (response.ok) {
    return await response.json();
  }
  
  throw new Error('Failed to fetch customers');
}
```

#### Get Customer Details
```typescript
async function getCustomer(userId) {
  const response = await fetchWithTokenRefresh(
    `/api/v1/admin/customers/${userId}`
  );
  
  if (response.ok) {
    return await response.json();
  }
  
  if (response.status === 404) {
    throw new Error('Customer not found');
  }
  
  throw new Error('Failed to fetch customer');
}
```

#### Suspend Customer
```typescript
async function suspendCustomer(userId, reason) {
  const response = await fetchWithTokenRefresh(
    `/api/v1/admin/customers/${userId}/suspend`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({ reason })
    }
  );
  
  if (response.ok) {
    return await response.json();
  }
  
  if (response.status === 403) {
    throw new Error('Insufficient permissions');
  }
  
  throw new Error('Failed to suspend customer');
}
```

#### Unsuspend Customer
```typescript
async function unsuspendCustomer(userId) {
  const response = await fetchWithTokenRefresh(
    `/api/v1/admin/customers/${userId}/suspend`,
    { method: 'DELETE', headers: getAuthHeaders() }
  );
  
  if (response.ok) {
    return await response.json();
  }
  
  throw new Error('Failed to unsuspend customer');
}
```

#### Get User Analytics
```typescript
async function getUserAnalytics() {
  const response = await fetchWithTokenRefresh(
    '/api/v1/admin/analytics/users'
  );
  
  if (response.ok) {
    const data = await response.json();
    return data.data; // Contains: total_users, deleted_users, suspended_users, etc.
  }
  
  throw new Error('Failed to fetch analytics');
}
```

#### Get Admin Logs
```typescript
async function getAdminLogs(limit = 50, offset = 0) {
  const response = await fetchWithTokenRefresh(
    `/api/v1/admin/logs?limit=${limit}&offset=${offset}`
  );
  
  if (response.ok) {
    return await response.json();
  }
  
  throw new Error('Failed to fetch logs');
}
```

### 4. Error Handling

All admin endpoints return standardized error responses:

```typescript
async function handleAdminError(response) {
  const data = await response.json();
  
  switch (response.status) {
    case 400:
      console.error('Invalid input:', data.error);
      return { type: 'validation', message: data.error };
      
    case 401:
      console.error('Unauthorized:', data.error);
      // Redirect to login
      redirectToLogin();
      return { type: 'auth', message: data.error };
      
    case 403:
      console.error('Forbidden:', data.error);
      return { type: 'permission', message: 'Insufficient permissions' };
      
    case 404:
      console.error('Not found:', data.error);
      return { type: 'notfound', message: data.error };
      
    case 409:
      console.error('Conflict:', data.error);
      return { type: 'conflict', message: data.error };
      
    case 500:
      console.error('Server error');
      return { type: 'server', message: 'An unexpected error occurred' };
      
    default:
      return { type: 'unknown', message: 'An error occurred' };
  }
}
```

### 5. React Hooks Example

```typescript
// useAdminAuth.ts
import { useState, useCallback } from 'react';

export function useAdminAuth() {
  const [isLoggedIn, setIsLoggedIn] = useState(() => {
    return !!localStorage.getItem('adminAccessToken');
  });
  const [adminEmail, setAdminEmail] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const login = useCallback(async (email: string, password: string) => {
    setLoading(true);
    setError('');
    try {
      const response = await fetch('/api/v1/admin/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Login failed');
      }

      const { access_token, refresh_token } = await response.json();
      localStorage.setItem('adminAccessToken', access_token);
      localStorage.setItem('adminRefreshToken', refresh_token);
      setAdminEmail(email);
      setIsLoggedIn(true);
      return true;
    } catch (err) {
      setError(err.message);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('adminAccessToken');
    localStorage.removeItem('adminRefreshToken');
    setIsLoggedIn(false);
    setAdminEmail('');
  }, []);

  return {
    isLoggedIn,
    adminEmail,
    error,
    loading,
    login,
    logout
  };
}

// useAdminAPI.ts
import { useState, useCallback } from 'react';

export function useAdminAPI() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchWithAuth = useCallback(async (url: string, options: RequestInit = {}) => {
    setLoading(true);
    setError('');
    
    try {
      const token = localStorage.getItem('adminAccessToken');
      
      let response = await fetch(url, {
        ...options,
        headers: {
          ...options.headers,
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      // Handle token expiration
      if (response.status === 401) {
        const refreshToken = localStorage.getItem('adminRefreshToken');
        const refreshResponse = await fetch('/api/v1/admin/auth/refresh', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken })
        });

        if (refreshResponse.ok) {
          const { access_token } = await refreshResponse.json();
          localStorage.setItem('adminAccessToken', access_token);

          // Retry with new token
          response = await fetch(url, {
            ...options,
            headers: {
              ...options.headers,
              'Authorization': `Bearer ${access_token}`,
              'Content-Type': 'application/json'
            }
          });
        } else {
          throw new Error('Session expired. Please log in again.');
        }
      }

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'API request failed');
      }

      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  return { fetchWithAuth, loading, error };
}

// Usage in a component
import React, { useState, useEffect } from 'react';
import { useAdminAuth } from './hooks/useAdminAuth';
import { useAdminAPI } from './hooks/useAdminAPI';

export function CustomerList() {
  const { isLoggedIn } = useAdminAuth();
  const { fetchWithAuth, loading } = useAdminAPI();
  const [customers, setCustomers] = useState([]);

  useEffect(() => {
    if (!isLoggedIn) return;

    fetchWithAuth('/api/v1/admin/customers?limit=20&offset=0')
      .then(data => setCustomers(data.data))
      .catch(err => console.error('Failed to load customers:', err));
  }, [isLoggedIn, fetchWithAuth]);

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      <h1>Customers</h1>
      <table>
        <thead>
          <tr>
            <th>Email</th>
            <th>Name</th>
            <th>Verified</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {customers.map(customer => (
            <tr key={customer.id}>
              <td>{customer.email}</td>
              <td>{customer.name}</td>
              <td>{customer.is_email_verified ? 'Yes' : 'No'}</td>
              <td>{customer.is_suspended ? 'Suspended' : 'Active'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### 6. Permission Checking

The API automatically checks permissions before executing actions. However, you can check permissions on the client side for UX purposes by:

1. Storing the admin role and permissions after login
2. Showing/hiding UI elements based on permissions

```typescript
// After login, store this:
const claims = decodeJWT(access_token); // If using JWT (careful with XSS)
// Or fetch from a profile endpoint

const hasPermission = (permission: string) => {
  const permissions = localStorage.getItem('adminPermissions')?.split(',') || [];
  const role = localStorage.getItem('adminRole');
  return role === 'super_admin' || permissions.includes(permission);
};

// Usage:
{hasPermission('manage_users') && <DeleteUserButton />}
{hasPermission('view_analytics') && <AnalyticsLink />}
```

### 7. Pagination

All list endpoints support pagination with `limit` and `offset`:

```typescript
// Get 25 items starting from item 50
async function getPaginatedData(resource: string, page: number, pageSize: number = 20) {
  const offset = (page - 1) * pageSize;
  const response = await fetchWithTokenRefresh(
    `/api/v1/admin/${resource}?limit=${pageSize}&offset=${offset}`
  );
  
  if (response.ok) {
    const data = await response.json();
    return {
      items: data.data,
      limit: data.limit,
      offset: data.offset,
      totalPages: Math.ceil((offset + data.data.length) / pageSize)
    };
  }
  
  throw new Error('Failed to fetch data');
}
```

### 8. Validation & Input

Always validate input on the client side before sending:

```typescript
function validatePassword(password: string): string | null {
  if (!password) return 'Password is required';
  if (password.length < 8) return 'Password must be at least 8 characters';
  return null;
}

function validateEmail(email: string): string | null {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) return 'Invalid email format';
  return null;
}

async function submitSuspension(userId: string, reason: string) {
  if (!reason.trim()) {
    alert('Please provide a reason for suspension');
    return;
  }

  try {
    await suspendCustomer(userId, reason);
    alert('Customer suspended successfully');
  } catch (error) {
    alert('Failed to suspend customer: ' + error.message);
  }
}
```

### 9. Testing

Use curl or Postman to test endpoints:

```bash
# Login
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "securepassword"
  }'

# List customers (with token)
curl -X GET http://localhost:8080/api/v1/admin/customers \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Suspend customer
curl -X POST http://localhost:8080/api/v1/admin/customers/USER_ID/suspend \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Violated community guidelines"}'
```

### 10. Troubleshooting

**"Insufficient permissions" error**
- Check your admin role and required permissions in the error response
- Super admins have all permissions; regular admins have limited permissions
- Contact a super admin to grant additional permissions

**"Admin not found" after login**
- The admin account may have been deleted
- Create a new admin account or contact support

**Token expiration issues**
- Implement automatic token refresh before making requests
- The refresh endpoint handles expired access tokens
- If refresh token is also expired, redirect to login

**CORS errors**
- These should not occur for same-origin requests
- For cross-origin requests, ensure the backend CORS is configured

---

## Summary

Key points for frontend integration:
1. Always use `Authorization: Bearer <token>` header for protected endpoints
2. Implement automatic token refresh on 401 responses
3. Store tokens securely (HTTP-only cookies recommended)
4. Validate user input before sending to API
5. Handle all error codes appropriately
6. Implement pagination for list endpoints
7. Display appropriate UI based on user permissions
8. Log out users if session expires or they lose connection

For more details, see [ADMIN_API.md](./ADMIN_API.md).
