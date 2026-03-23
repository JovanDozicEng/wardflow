# WardFlow Frontend - Testing Guide

## ✅ Authentication Implementation Complete!

The complete authentication flow has been implemented and is ready to test.

### 🚀 Quick Start

1. **Start Backend** (if not running):
```bash
cd backend
podman-compose up -d
```

2. **Start Frontend**:
```bash
cd frontend
npm run dev
```

Frontend runs on: **http://localhost:5176/**

---

## 📝 Test Scenarios

### 1. Registration Flow

**Steps:**
1. Open http://localhost:5176/
2. Click "Create Account" button
3. Fill in the registration form:
   - **Name:** John Doe
   - **Email:** john@example.com
   - **Role:** Nurse (select from dropdown)
   - **Password:** Test1234
   - **Confirm Password:** Test1234
4. Click "Create Account"

**Expected Result:**
- ✅ Form validates password (8+ chars, uppercase, lowercase, number)
- ✅ Passwords must match
- ✅ On success: Redirects to Dashboard
- ✅ Header shows "John Doe" and "Nurse"
- ✅ Dashboard displays user information

**Backend Request:**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"Test1234","name":"John Doe","role":"nurse"}'
```

---

### 2. Login Flow

**Steps:**
1. Click "Logout" in header (if logged in)
2. Go to http://localhost:5176/login
3. Enter credentials:
   - **Email:** john@example.com
   - **Password:** Test1234
4. Click "Sign In"

**Expected Result:**
- ✅ On success: Redirects to Dashboard
- ✅ Token stored in localStorage
- ✅ User info displayed in header
- ✅ On error: Shows error message

**Backend Request:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"Test1234"}'
```

**Response:**
```json
{
  "token": "eyJhbGci...",
  "expiresAt": 1774366541,
  "user": {
    "id": "uuid",
    "email": "john@example.com",
    "name": "John Doe",
    "role": "nurse",
    "isActive": true
  }
}
```

---

### 3. Protected Route Access

**Test While Logged Out:**
1. Make sure you're logged out
2. Try to access http://localhost:5176/dashboard directly

**Expected Result:**
- ✅ Automatically redirects to /login
- ✅ After login, returns to dashboard

**Test While Logged In:**
1. Login first
2. Access http://localhost:5176/dashboard

**Expected Result:**
- ✅ Shows dashboard with user info
- ✅ Layout with Header and Sidebar visible

---

### 4. Logout Flow

**Steps:**
1. While logged in, click "Logout" button in header

**Expected Result:**
- ✅ Token removed from localStorage
- ✅ User state cleared
- ✅ Redirects to home page
- ✅ Accessing /dashboard now redirects to /login

---

### 5. Form Validation

**Test Invalid Email:**
1. On registration/login form
2. Enter: "notanemail"
3. Click submit

**Expected Result:**
- ✅ Shows "Invalid email address" error

**Test Weak Password:**
1. On registration form
2. Enter password: "weak"
3. Click submit

**Expected Result:**
- ✅ Shows validation errors:
  - "Password must be at least 8 characters"
  - "Password must contain at least one uppercase letter"
  - "Password must contain at least one number"

**Test Password Mismatch:**
1. On registration form
2. Password: Test1234
3. Confirm: Different123
4. Click submit

**Expected Result:**
- ✅ Shows "Passwords don't match" error

---

### 6. API Error Handling

**Test Invalid Credentials:**
1. Go to login page
2. Enter wrong password
3. Click "Sign In"

**Expected Result:**
- ✅ Shows error message: "invalid email or password"

**Test Duplicate Email:**
1. Register with an existing email
2. Click "Create Account"

**Expected Result:**
- ✅ Shows error from backend

---

## 🔍 Developer Tools Testing

### Check localStorage:
```javascript
// Open browser DevTools → Console
localStorage.getItem('auth_token')
// Should show JWT token when logged in
```

### Check Network Requests:
1. Open DevTools → Network tab
2. Filter by "Fetch/XHR"
3. Login/Register
4. Check requests to `/auth/login` or `/auth/register`

**Headers should include:**
```
Authorization: Bearer eyJhbGci...
```

### Check Zustand Store:
```javascript
// In DevTools Console
// After using React DevTools
// Check component state for auth data
```

---

## 🧪 Backend Integration Tests

### Test with curl:

**Register:**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@wardflow.com",
    "password": "Secure123",
    "name": "Test User",
    "role": "provider"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@wardflow.com",
    "password": "Secure123"
  }'
```

**Access Protected Endpoint:**
```bash
TOKEN="your-token-here"
curl -X GET http://localhost:8080/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

---

## ✨ Features Implemented

### ✅ User Interface
- Home page with landing design
- Login page with form validation
- Registration page with role selection
- Dashboard with user stats
- Header with user info and logout
- Responsive layout (mobile/desktop)

### ✅ Form Validation
- React Hook Form integration
- Zod schema validation
- Real-time error messages
- Password strength requirements
- Email format validation
- Password confirmation matching

### ✅ Authentication Logic
- JWT token storage
- Automatic token injection (axios interceptor)
- Token persistence (localStorage + Zustand)
- Protected route guards
- Auto-redirect on auth state change
- Logout functionality

### ✅ Backend Integration
- Connected to Go API
- Correct request/response formats
- Error handling with user-friendly messages
- CORS-ready
- TypeScript types match backend DTOs

---

## 🐛 Troubleshooting

### Backend not responding?
```bash
# Check if backend is running
podman ps | grep wardflow

# Check backend logs
podman logs wardflow-backend

# Restart backend
cd backend && podman-compose restart
```

### CORS errors?
- Backend should already have CORS enabled
- Check backend logs for CORS configuration
- Verify API_BASE_URL in .env

### Token not persisting?
- Check browser DevTools → Application → Local Storage
- Look for "auth-storage" key
- Clear storage and try again

### Build errors?
```bash
cd frontend
rm -rf node_modules
npm install
npm run build
```

---

## 📊 Expected Results Summary

| Action | URL | Expected Outcome |
|--------|-----|------------------|
| Visit home | / | Shows landing page with CTAs |
| Click "Sign In" | /login | Shows login form |
| Click "Create Account" | /register | Shows registration form |
| Submit login | /login → /dashboard | Redirects to dashboard |
| Submit register | /register → /dashboard | Creates user, redirects |
| Access dashboard (logged out) | /dashboard → /login | Redirects to login |
| Access dashboard (logged in) | /dashboard | Shows user dashboard |
| Click logout | (any page) → / | Clears session, redirects home |

---

## 🎯 Next Steps

Now that authentication is working, you can:

1. **Test with real backend data**
2. **Add more features:**
   - Password reset
   - Email verification
   - Session timeout
   - Remember me
3. **Build other modules:**
   - Care Team Management
   - Task Board
   - Consults
   - Patient Flow
4. **Enhance UI:**
   - Better animations
   - Loading skeletons
   - Toast notifications (already configured!)

---

**Status: ✅ READY FOR TESTING**

All authentication flows are implemented and connected to your backend!
