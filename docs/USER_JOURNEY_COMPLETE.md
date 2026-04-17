# Complete User Journey: From Landing Page to Active User

**Document Version:** 1.0  
**Last Updated:** April 17, 2026  
**Status:** Production Ready

---

## Table of Contents

1. [Overview](#overview)
2. [Phase 1: Discovery & Landing](#phase-1-discovery--landing)
3. [Phase 2: Authentication & Registration](#phase-2-authentication--registration)
4. [Phase 3: Email Verification](#phase-3-email-verification)
5. [Phase 4: Onboarding](#phase-4-onboarding)
6. [Phase 5: Workspace Setup](#phase-5-workspace-setup)
7. [Phase 6: First Feature: Instagram Connection](#phase-6-first-feature-instagram-connection)
8. [Phase 7: Subscription & Payment](#phase-7-subscription--payment)
9. [Phase 8: Active Usage](#phase-8-active-usage)
10. [Appendix: API Endpoints Reference](#appendix-api-endpoints-reference)

---

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      USER JOURNEY FLOW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Landing Page                                                   │
│      ↓                                                           │
│  Sign Up / Login                                                │
│      ↓                                                           │
│  Email Verification                                             │
│      ↓                                                           │
│  Onboarding (Profile Setup)                                     │
│      ↓                                                           │
│  Create/Select Workspace                                        │
│      ↓                                                           │
│  Connect First Account (Instagram)                              │
│      ↓                                                           │
│  Subscribe to Pro Plan                                          │
│      ↓                                                           │
│  Dashboard & Active Usage                                       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Metrics

- **Target Time for Complete Onboarding:** 5-10 minutes
- **Required API Calls:** ~8-12 requests
- **Number of Screens:** 8-10 views
- **Critical Decision Points:** 3 (Sign up, Connect Account, Subscribe)

---

## Phase 1: Discovery & Landing

### 1.1 Initial Landing Page

**URL:** `https://refyne.app/`

**User Actions:**
- Views marketing landing page
- Sees key features:
  - Instagram integration
  - AI-powered insights
  - Content recommendations
  - Analytics dashboard
- Sees call-to-action buttons:
  - "Get Started" (primary)
  - "Learn More" (secondary)

**UI Elements:**
- Hero section with product demo/screenshot
- Features overview with icons
- Pricing teaser (Pro Plan: $X/month)
- Testimonials/social proof
- FAQ section
- Footer with login link

**What Happens:**
- No API calls required
- Completely client-side
- Session cookie created (for tracking)

**Next Step:** User clicks "Get Started" → Navigate to Sign Up page

---

## Phase 2: Authentication & Registration

### 2.1 Sign Up Page

**URL:** `https://refyne.app/signup`

**User Inputs:**
```
- First Name (required, max 100 chars)
- Last Name (required, max 100 chars)
- Email (required, valid email format)
- Username (required, 3-30 chars, alphanumeric + underscore)
- Password (required, min 8 chars, strong complexity)
- Confirm Password (required, must match password)
- Terms & Conditions (checkbox, required)
```

**API Call #1: Register User**
```bash
POST /api/auth/register
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "username": "johndoe",
  "password": "SecurePass123!",
  "accept_terms": true
}
```

**Success Response (201):**
```json
{
  "status": "ok",
  "message": "Registration successful. Please verify your email.",
  "data": {
    "user_id": "usr_123abc",
    "email": "john@example.com",
    "verification_status": "pending"
  }
}
```

**Error Scenarios:**
- Email already exists → 409 Conflict
- Username already exists → 409 Conflict
- Password too weak → 400 Bad Request
- Terms not accepted → 400 Bad Request

**What Happens After:**
- User account created in database (active = false, until verified)
- Verification email sent to provided email address
- Frontend shows success message
- Auto-redirect to verification page (or show "Check your email" message)

**Next Step:** User checks email → Click verification link

---

## Phase 3: Email Verification

### 3.1 Email Verification Flow

**Email Received:**
```
Subject: Verify Your Refyne Account
From: noreply@refyne.app

Body:
---
Hi John,

Welcome to Refyne! Click the link below to verify your email:

https://refyne.app/verify?token=verification_token_xyz...

This link expires in 24 hours.

If you didn't create this account, please ignore this email.
---
```

### 3.2 Verification Page

**URL:** `https://refyne.app/verify?token=verification_token_xyz`

**API Call #2: Verify Email**
```bash
GET /api/auth/verify?token=verification_token_xyz
```

**Success Response (200):**
```json
{
  "status": "ok",
  "message": "Email verified successfully",
  "data": {
    "user_id": "usr_123abc",
    "email": "john@example.com",
    "verified": true
  }
}
```

**What Happens:**
- User account activated (active = true)
- Verification token marked as used
- Success message displayed
- Auto-redirect to login page (or login modal)

**Alternative: Resend Verification Email**
```bash
POST /api/auth/resend-verification
Content-Type: application/json

{ "email": "john@example.com" }
```

Response: 200 OK with "Verification email sent"

**Next Step:** User logs in with credentials

---

## Phase 4: Authentication - Login

### 4.1 Login Page

**URL:** `https://refyne.app/login`

**User Inputs:**
```
- Email (required)
- Password (required)
```

**API Call #3: Login**
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Success Response (200):**
```json
{
  "status": "ok",
  "message": "Login successful",
  "data": {
    "user_id": "usr_123abc",
    "email": "john@example.com",
    "username": "johndoe",
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "onboarding_completed": false,
    "subscription_status": "onboarding"
  }
}
```

**Token Details:**
- `access_token`: 15 minutes validity
- `refresh_token`: 7 days validity
- Both stored in localStorage/httpOnly cookies

**Error Scenarios:**
- Email not found → 401 Unauthorized
- Password incorrect → 401 Unauthorized (max 5 attempts → 15m lockout)
- Email not verified → 403 Forbidden

**What Happens:**
- Tokens stored in local storage
- User logged in state set
- User info cached in application state
- Auto-redirect to next step based on `onboarding_completed` status

**Next Step:** Check if onboarding completed → If no, proceed to Onboarding

---

## Phase 5: Onboarding

### 5.1 Onboarding Page

**URL:** `https://refyne.app/onboarding`

**Purpose:** Collect additional user information and preferences

**User Inputs:**
```
- Display Name (optional, auto-filled from profile)
- Industry/Niche selection (dropdown):
  - Fitness
  - Fashion
  - Food & Beverage
  - Technology
  - E-commerce
  - Other
- Content Type (multi-select):
  - Photography
  - Video
  - Carousel Posts
  - Reels/Stories
  - Mixed
- Goals (multi-select):
  - Grow followers
  - Increase engagement
  - Drive traffic/sales
  - Build brand awareness
  - Content ideas
- Language preference (select, default: English)
- Timezone (auto-detect, allow override)
- Push notifications (toggle, default: on)
- Email notifications frequency (select):
  - Daily
  - Weekly
  - Monthly
```

**API Call #4: Complete Onboarding**
```bash
POST /api/user/onboarding
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "industry": "fitness",
  "content_types": ["photography", "reels"],
  "goals": ["grow_followers", "increase_engagement"],
  "language": "en",
  "timezone": "America/New_York",
  "notifications": {
    "push_enabled": true,
    "email_frequency": "weekly"
  }
}
```

**Success Response (200):**
```json
{
  "status": "ok",
  "message": "Onboarding completed",
  "data": {
    "user_id": "usr_123abc",
    "onboarding_completed": true,
    "next_step": "workspace_selection"
  }
}
```

**What Happens:**
- User preferences saved to database
- Onboarding status marked as complete
- Profile information updated
- Auto-redirect to workspace selection

**Skip Option:** User can skip → "You can update these later"

**Next Step:** Select or create workspace

---

## Phase 6: Workspace Setup

### 6.1 Workspace Selection Page

**URL:** `https://refyne.app/workspace`

**Purpose:** Allow users to create or join existing workspace

**Option A: First-Time User (No Workspaces)**

Display:
- Welcome message: "Let's set up your workspace"
- Button: "Create Workspace"

**API Call #5: Create First Workspace**
```bash
POST /api/workspace
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "name": "My Brand",
  "description": "Fitness content and community"
}
```

**Success Response (201):**
```json
{
  "status": "ok",
  "message": "Workspace created",
  "data": {
    "workspace_id": "ws_789xyz",
    "name": "My Brand",
    "description": "Fitness content and community",
    "role": "owner",
    "created_at": "2026-04-17T10:30:00Z"
  }
}
```

**Option B: Existing User (Has Workspaces)**

Display list of:
- Workspaces owned by user
- Workspaces user is a member of
- Button: "+ Create New Workspace"

**API Call #5b: List Workspaces**
```bash
GET /api/workspace
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "status": "ok",
  "data": [
    {
      "workspace_id": "ws_789xyz",
      "name": "My Brand",
      "role": "owner",
      "member_count": 1
    },
    {
      "workspace_id": "ws_456def",
      "name": "Team Project",
      "role": "member",
      "member_count": 3
    }
  ]
}
```

### 6.2 User Selects or Creates Workspace

**What Happens:**
- Workspace context set in application
- User considered as workspace owner or member
- Auto-redirect to dashboard or next setup step

**Note for Teams:**
- Owner can invite team members (in Phase 6)
- Members get email invitation with join link
- Members accept and join workspace

**Next Step:** Connect first Instagram account

---

## Phase 7: First Feature - Instagram Connection

### 7.1 Instagram Connection Page

**URL:** `https://refyne.app/workspace/ws_789xyz/settings/instagram` or `https://refyne.app/instagram/connect`

**Purpose:** Connect user's Instagram Business Account

**Display:**
- Explanation of what data will be accessed
- Permissions requested:
  - View Instagram account info
  - Access media and insights
  - Receive real-time webhooks
- Button: "Connect Instagram Account"

### 7.2 OAuth Flow - Instagram Authorization

**API Call #6: Generate OAuth URL**
```bash
POST /api/instagram/auth/connect
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "workspace_id": "ws_789xyz"
}
```

**Response (200):**
```json
{
  "status": "ok",
  "data": {
    "auth_url": "https://api.instagram.com/oauth/authorize?client_id=...",
    "message": "Redirect to this URL to authorize with Instagram"
  }
}
```

**User is Redirected to Instagram:**
```
1. User clicks "Connect Instagram"
2. Redirected to Instagram OAuth page
3. User logs into Instagram (if not already logged in)
4. Instagram shows permissions dialog:
   - Access your profile information
   - View your media and insights
   - Receive webhooks for account updates
5. User clicks "Approve"
```

### 7.3 OAuth Callback & Token Storage

**Instagram Redirects Back to:**
```
https://refyne.app/instagram/auth/callback?code=AUTHORIZATION_CODE&state=STATE
```

**API Call #7: Exchange Authorization Code for Token**
```bash
GET /api/instagram/auth/callback?code=AUTHORIZATION_CODE&state=STATE
Authorization: Bearer {access_token}
```

**Response (200):**
```json
{
  "status": "ok",
  "data": {
    "account_id": "ig_account_001",
    "instagram_user_id": "123456789",
    "username": "@mybrand",
    "connected_at": "2026-04-17T10:45:00Z",
    "sync_status": "syncing",
    "message": "Instagram account connected! We're syncing your media..."
  }
}
```

**What Happens Backend:**
- Authorization code exchanged for long-lived access token
- Token encrypted and stored in database
- Access token: 2-hour validity
- Refresh token: 60-day validity (auto-refreshed before expiry)
- Webhook registered for real-time updates
- Background job queued to sync initial media

### 7.4 First Sync

**What User Sees:**
- Loading screen: "Syncing your Instagram account..."
- Progress indicator showing:
  - ✓ Account verified
  - • Fetching posts...
  - ○ Analyzing content...
  - ○ Setting up insights...

**Behind the Scenes:**
- River job `instagram_sync_media` starts
- Fetches last 50-100 posts from Instagram API
- Stores in database for quick retrieval
- Queues AI analysis job for each post
- Fetches account insights

**Completion:**
- Success message: "Instagram account synced! Found X posts"
- Show preview of latest posts

**Next Step:** Continue or subscribe

---

## Phase 8: Subscription & Payment

### 8.1 Subscription Page

**URL:** `https://refyne.app/workspace/ws_789xyz/subscription` or `https://refyne.app/subscribe`

**Display:**
- Current subscription status: "You're on the Free Plan"
- Plan comparison:
  ```
  ┌─────────────────────────────────────────┐
  │                FREE PLAN                │
  ├─────────────────────────────────────────┤
  │ ✓ 1 Instagram Account                   │
  │ ✓ Basic Media Sync                      │
  │ ✓ Limited Analytics (last 7 days)       │
  │ ✗ AI Content Recommendations            │
  │ ✗ Team Members                          │
  │ ✗ Advanced Scheduling                   │
  │ Price: Free                             │
  │ Action: Currently on this plan          │
  └─────────────────────────────────────────┘
  
  ┌─────────────────────────────────────────┐
  │                PRO PLAN                 │
  ├─────────────────────────────────────────┤
  │ ✓ Unlimited Instagram Accounts          │
  │ ✓ Real-time Media Sync                  │
  │ ✓ Full Analytics History                │
  │ ✓ AI Content Recommendations            │
  │ ✓ Team Members (up to 5)                │
  │ ✓ Advanced Scheduling                   │
  │ ✓ Priority Support                      │
  │ Price: $29/month (or $290/year)         │
  │ Action: [UPGRADE TO PRO] Button         │
  └─────────────────────────────────────────┘
  ```

### 8.2 Upgrade to Pro Plan

**API Call #8: Create Checkout Session**
```bash
POST /api/subscription/checkout
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "workspace_id": "ws_789xyz",
  "plan": "pro",
  "billing_cycle": "monthly"  # or "annual"
}
```

**Response (200):**
```json
{
  "status": "ok",
  "data": {
    "checkout_url": "https://checkout.paddle.com/checkout/...",
    "session_id": "session_123",
    "amount": 29.00,
    "currency": "USD",
    "billing_cycle": "monthly"
  }
}
```

### 8.3 Paddle Checkout Page

**User Redirected to:**
```
https://checkout.paddle.com/checkout/...
```

**Checkout Flow:**
1. Enter email (auto-filled from account)
2. Enter billing address
3. Select payment method:
   - Credit/Debit Card
   - PayPal
   - Apple Pay
   - Google Pay
4. Review order:
   - Product: Refyne Pro Plan
   - Amount: $29.00 USD
   - Billing: Monthly (auto-renew)
5. Click "Complete Purchase"
6. Payment processed

### 8.4 Payment Confirmation

**After Successful Payment:**

**Paddle Redirects to:**
```
https://refyne.app/subscription/success?transaction_id=...
```

**Show Success Message:**
```
✓ Payment successful!
  Amount: $29.00 USD
  Plan: Refyne Pro (monthly)
  Next billing: May 17, 2026
  
  [Return to Dashboard]
```

**Backend Updates:**
- Subscription record created
- Workspace subscription status: "active"
- User features unlocked
- Welcome email sent
- API call #9: Webhook from Paddle received and processed

---

## Phase 9: Active Usage

### 9.1 Dashboard Access

**URL:** `https://refyne.app/dashboard` or `https://refyne.app/workspace/ws_789xyz`

**Post-Subscribe Access Granted to:**

#### 1. Instagram Management
```
POST /api/instagram/media/sync
GET /api/instagram/accounts
GET /api/instagram/media
GET /api/instagram/analytics
POST /api/instagram/ai/caption-suggest
POST /api/instagram/ai/hashtag-suggest
```

#### 2. AI Assistant (Otto)
```
POST /api/otto/conversations  # Start AI chat
POST /api/otto/conversations/{id}/messages  # Send message
GET /api/otto/conversations/{id}/messages  # Get history
```

#### 3. Analytics & Insights
```
GET /api/instagram/analytics
GET /api/instagram/analytics/media
GET /api/instagram/analytics/trends
```

#### 4. Workspace Management
```
GET /api/workspace  # View workspace details
POST /api/workspace/:id/members  # Invite team members
GET /api/workspace/:id/members  # List members
```

### 9.2 Default Dashboard View

**Main Dashboard:**
```
┌─────────────────────────────────────────────────┐
│ Welcome back, John!                             │
├─────────────────────────────────────────────────┤
│                                                 │
│ 📊 Account Overview                             │
│ ├─ Total Posts: 142                             │
│ ├─ Followers: 2,341 (↑ 8% this week)            │
│ ├─ Total Engagement: 4,521 (↑ 12% this week)    │
│ └─ Avg. Engagement Rate: 3.2%                   │
│                                                 │
│ 📈 This Week's Performance                      │
│ ├─ Best post: Photo of new product (542 likes) │
│ ├─ Most engaged audience: 25-34 years old       │
│ └─ Best posting time: Tuesday 2-3 PM            │
│                                                 │
│ 🤖 AI Recommendations                           │
│ ├─ Caption idea for new product: "..."         │
│ ├─ Suggested posting time: Tomorrow at 2 PM    │
│ └─ Hashtag recommendations: #fitness #goals    │
│                                                 │
│ 💬 Start Conversation                           │
│ [Chat with AI about content strategy]           │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 9.3 Key User Actions

**User Can Now:**

1. **View & Analyze Media**
   - Browse all Instagram posts
   - View detailed analytics per post
   - See engagement trends over time

2. **Connect Multiple Accounts** (Pro feature)
   - Add second, third, etc. Instagram accounts
   - Each account synced in real-time

3. **Use AI Assistant (Otto)**
   - Ask questions about content strategy
   - Get caption suggestions
   - Discuss performance metrics
   - Get posting recommendations

4. **Invite Team Members** (Pro feature)
   - Invite up to 5 team members
   - Assign roles: Owner or Member
   - Members get dashboard access

5. **Manage Settings**
   - Update profile
   - Configure notifications
   - Manage connected accounts
   - View subscription details

### 9.4 Real-Time Updates

**Webhook Events Processed:**
- New post published → Synced automatically
- New comments/likes → Metrics updated
- Follower changes → Analytics updated
- Direct messages → Can be viewed in platform

---

## Complete Flow Diagram

```
START (Landing Page)
     ↓
   [Sign Up]
     ↓
[Email Verification] ← (Resend option)
     ↓
   [Login]
     ↓
[Onboarding] ← (Skip option)
     ↓
[Workspace Selection/Creation]
     ↓
[Connect Instagram Account]
     ↓
[Sync Instagram Media]
     ↓
   ┌─[Choose Subscription]─┐
   ↓                        ↓
[Pro Plan]            [Free Plan]
   ↓                        ↓
[Paddle Checkout]    [Limited Access]
   ↓                        ↓
[Payment]             [Dashboard]
   ↓                        ↓
[Success]             [Upgrade Prompt]
   ↓                        ↓
[Full Dashboard] ←──────────┴────→ [Full Features]
     ↓
[Active Usage]
   ├─ Analyze Posts
   ├─ Connect More Accounts
   ├─ Chat with AI (Otto)
   ├─ Invite Team Members
   └─ Manage Settings
```

---

## Appendix: API Endpoints Reference

### Authentication Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|----------------|
| POST | `/api/auth/register` | Create new user account | No |
| GET | `/api/auth/verify` | Verify email address | No |
| POST | `/api/auth/resend-verification` | Resend verification email | No |
| POST | `/api/auth/login` | Authenticate user | No |
| POST | `/api/auth/logout` | Logout user | Yes |
| POST | `/api/auth/refresh-token` | Refresh access token | Yes |
| POST | `/api/auth/forgot-password` | Request password reset | No |
| POST | `/api/auth/reset-password` | Complete password reset | No |

### User Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|----------------|
| GET | `/api/user/profile` | Get user profile | Yes |
| PUT | `/api/user/profile` | Update user profile | Yes |
| GET | `/api/user/settings` | Get user settings | Yes |
| PUT | `/api/user/settings` | Update user settings | Yes |
| POST | `/api/user/onboarding` | Complete onboarding | Yes |
| DELETE | `/api/user/account` | Delete user account | Yes |

### Workspace Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|----------------|
| POST | `/api/workspace` | Create workspace | Yes |
| GET | `/api/workspace` | List user's workspaces | Yes |
| GET | `/api/workspace/:id` | Get workspace details | Yes |
| PUT | `/api/workspace/:id` | Update workspace | Yes |
| DELETE | `/api/workspace/:id` | Delete workspace | Yes |
| GET | `/api/workspace/:id/members` | List workspace members | Yes |
| POST | `/api/workspace/:id/members` | Invite member | Yes |
| DELETE | `/api/workspace/:id/members/:user_id` | Remove member | Yes |

### Instagram Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|----------------|
| POST | `/api/instagram/auth/connect` | Initiate OAuth flow | Yes |
| GET | `/api/instagram/auth/callback` | OAuth callback | Yes |
| POST | `/api/instagram/auth/disconnect` | Revoke access | Yes |
| GET | `/api/instagram/accounts` | List connected accounts | Yes |
| GET | `/api/instagram/accounts/:id` | Get account details | Yes |
| GET | `/api/instagram/media` | List media | Yes |
| GET | `/api/instagram/media/:id` | Get media details | Yes |
| GET | `/api/instagram/media/:id/ai` | Get AI recommendations | Yes |
| GET | `/api/instagram/analytics` | Get account analytics | Yes |
| GET | `/api/instagram/analytics/media` | Get media analytics | Yes |
| POST | `/api/instagram/ai/caption-suggest` | Generate captions | Yes |
| POST | `/api/instagram/ai/hashtag-suggest` | Generate hashtags | Yes |

### Subscription Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|----------------|
| POST | `/api/subscription/checkout` | Create checkout session | Yes |
| GET | `/api/subscription/status` | Get subscription status | Yes |
| GET | `/api/subscription/portal` | Access billing portal | Yes |
| POST | `/api/subscription/webhooks` | Paddle webhook receiver | No |

### Otto AI Endpoints

| Method | Endpoint | Purpose | Auth Required |
|--------|----------|---------|----------------|
| POST | `/api/otto/conversations` | Create conversation | Yes |
| GET | `/api/otto/conversations` | List conversations | Yes |
| GET | `/api/otto/conversations/:id` | Get conversation | Yes |
| PUT | `/api/otto/conversations/:id` | Update conversation | Yes |
| POST | `/api/otto/conversations/:id/archive` | Archive conversation | Yes |
| POST | `/api/otto/conversations/:id/bookmark` | Bookmark conversation | Yes |
| DELETE | `/api/otto/conversations/:id` | Delete conversation | Yes |
| POST | `/api/otto/conversations/:id/messages` | Send message | Yes |
| GET | `/api/otto/conversations/:id/messages` | Get messages | Yes |
| POST | `/api/otto/messages/:id/feedback` | Add feedback | Yes |

---

## Key Takeaways

### For Frontend Developers:
- Phase 1-4 (Auth & Onboarding): Pure client-side + auth APIs
- Phase 5-6 (Workspace): Manage context and state
- Phase 7 (Instagram): OAuth flow handling
- Phase 8 (Subscription): Redirect to Paddle, handle callbacks
- Phase 9 (Active Usage): Real-time data fetching and updates

### For Product Managers:
- Total user journey: ~30 minutes from landing to active usage
- 4 critical drop-off points: Sign up, Email verification, Instagram connect, Payment
- Free plan keeps users engaged with limited features
- Pro plan removes all restrictions

### For Designers:
- Keep onboarding to 2-3 screens maximum
- Clear messaging about data being requested (Instagram OAuth)
- Success confirmations after each major action
- Loading states during sync operations
- Progressive feature unlock based on subscription

---

**Document End**

For questions or clarifications, refer to the full API documentation in `FRONTEND_API_INTEGRATION_UPDATED.md`
