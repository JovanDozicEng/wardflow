# WardFlow Frontend Navigation Guide

## 🎯 Feature Access Overview

All Developer C features are now accessible via the updated sidebar navigation.

---

## 📍 Navigation Structure

### **Governance & Safety** (Always Visible)
| Link | URL | Access |
|------|-----|--------|
| 🏥 Consults | `/consults` | All users (View Only for non-providers) |
| ⚠️ Exceptions | `/exceptions` | All authenticated users |
| 📝 Report Incident | `/incidents/report` | All authenticated users |
| 🔍 Review Incidents | `/incidents/review` | `quality_safety`, `admin` only |

### **Care Coordination** (Coming Soon - Disabled)
- 👥 Encounters
- 📋 Tasks

### **Operations** (Coming Soon - Disabled)
- 🛏️ Bed Management
- 🚑 Transport
- 🏠 Discharge Planning

---

## 👤 Role-Based Access Control

### **Admin** (`admin`)
**Full Access to Everything:**
- ✅ Create/manage consults
- ✅ Create/edit/finalize/correct exceptions
- ✅ Report incidents
- ✅ Access incident review queue
- ✅ Update incident status

### **Provider** (`provider`)
**Clinical Features:**
- ✅ Create/manage consults
- ✅ Create/edit/finalize exceptions (cannot correct)
- ✅ Report incidents
- ❌ No incident review access

### **Quality/Safety** (`quality_safety`)
**Safety & Compliance Focus:**
- ✅ View consults (cannot manage)
- ✅ Create/finalize/correct exceptions
- ✅ Report incidents
- ✅ Access incident review queue
- ✅ Update incident status

### **Consult Service** (`consult`)
**Consultation Management:**
- ✅ Create/manage consults
- ✅ Create/finalize exceptions (cannot correct)
- ✅ Report incidents
- ❌ No incident review access

### **Nurse** (`nurse`)
**Basic Access:**
- ✅ View consults (read-only, see "View Only" label)
- ✅ Create/edit exceptions (cannot finalize)
- ✅ Report incidents
- ❌ No incident review access

### **Charge Nurse** (`charge_nurse`)
**Supervisory Access:**
- ✅ View consults (read-only)
- ✅ Create/finalize exceptions (cannot correct)
- ✅ Report incidents
- ❌ No incident review access

---

## 🎨 Visual Cues

### **Active Route**
- Blue background (`bg-blue-100`)
- Blue text (`text-blue-700`)
- Indicates current page

### **Available Route**
- Gray text (`text-gray-700`)
- Hover: light gray background
- Clickable navigation

### **Disabled Route** (Coming Soon)
- Light gray text (`text-gray-400`)
- Gray background (`bg-gray-50`)
- Non-clickable
- Shows "(Coming Soon)" label

### **View-Only Access**
- Normal navigation link
- Shows "(View Only)" label for nurses on Consults
- Users can view but cannot perform management actions

---

## 🚀 Getting Started

### 1. Start the Development Server
```bash
cd frontend
npm run dev
```

### 2. Login with Test User
Navigate to `http://localhost:5173/login` and authenticate.

### 3. Access Features via Sidebar
The sidebar is visible on all protected pages (after login). Simply click any active link.

### 4. Test Different Roles
Create test users with different roles to see how navigation adapts:

**Admin User Example:**
```json
{
  "email": "admin@wardflow.com",
  "role": "admin",
  "name": "Admin User"
}
```

**Quality/Safety User Example:**
```json
{
  "email": "safety@wardflow.com",
  "role": "quality_safety",
  "name": "Safety Officer"
}
```

**Nurse User Example:**
```json
{
  "email": "nurse@wardflow.com",
  "role": "nurse",
  "name": "Nurse Staff"
}
```

---

## 📋 Feature Walkthrough

### **Consults Workflow**
1. Click "🏥 Consults" in sidebar
2. View consult inbox with status filters
3. Click "+ New Consult" (if you have provider/consult/admin role)
4. Fill form: Encounter ID, Target Service, Urgency, Details
5. Submit to create consult request
6. Use action buttons: Accept, Decline, Redirect, Complete

### **Exceptions Workflow**
1. Click "⚠️ Exceptions" in sidebar
2. View exception list with status/type filters
3. Click "+ New Exception" to create draft
4. Fill form: Encounter ID, Type, JSON data
5. Draft exceptions can be edited
6. Click "Finalize" to make immutable (provider+ roles)
7. Click "Correct" on finalized exceptions (quality_safety/admin only)

### **Incidents Workflow**
1. Click "📝 Report Incident" to report new incident
2. Fill form: Type, Severity, Event Time, Harm Indicators (JSON)
3. Submit incident report
4. **Quality/Safety users:** Click "🔍 Review Incidents"
5. View submitted incidents awaiting review
6. Click "Review Incident" to update status
7. Change status: submitted → under_review → closed

---

## 🔧 Customization

### Add More Navigation Items
Edit `/frontend/src/shared/components/layout/Sidebar.tsx`:

```typescript
// Add new section
<div className="pt-4 border-t border-gray-200 mt-4">
  <p className="px-4 text-xs font-semibold text-gray-500 uppercase mb-2">
    Your Section
  </p>
  
  <NavLink to="/your-route" isActive={isActive('/your-route')}>
    🎯 Your Feature
  </NavLink>
</div>
```

### Change Role Requirements
Modify permission checks in Sidebar component:

```typescript
const canAccessYourFeature = hasAnyRole(['role1', 'role2', 'admin']);

{canAccessYourFeature && (
  <NavLink to="/your-route" isActive={isActive('/your-route')}>
    Your Feature
  </NavLink>
)}
```

---

## 🐛 Troubleshooting

### "I can't see a navigation link"
**Cause:** Your user role doesn't have permission.  
**Solution:** Check the RBAC table above and login with appropriate role.

### "Link shows 'View Only'"
**Cause:** You have read access but not management permission.  
**Solution:** This is expected. You can view data but cannot create/edit. Login with provider/consult/admin role for full access.

### "All links are disabled"
**Cause:** Not logged in or session expired.  
**Solution:** Navigate to `/login` and authenticate.

### "Incident Review link is missing"
**Cause:** Only `quality_safety` and `admin` roles can see this link.  
**Solution:** Login with quality_safety or admin role.

---

## ✅ Summary

The sidebar navigation is now:
- ✅ **Role-aware** - Shows/hides links based on permissions
- ✅ **Visual feedback** - Active routes highlighted in blue
- ✅ **User-friendly** - Disabled state for unimplemented features
- ✅ **Type-safe** - Uses centralized ROUTES constants
- ✅ **Accessible** - Proper semantic HTML with Link components

All three Developer C features (Consults, Exceptions, Incidents) are now fully navigable!
