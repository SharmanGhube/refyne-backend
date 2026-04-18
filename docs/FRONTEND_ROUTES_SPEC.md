# Refyne Frontend Routes & Pages Specification

This document outlines all frontend pages, routes, and components required to integrate with the Refyne backend API. Each page is mapped to the corresponding backend endpoints.

**Last Updated:** 2026-04-18  
**Target Framework:** React/Next.js or equivalent  
**API Base URL:** `https://api.refyne.io` (or Railway dev URL)

---

## Table of Contents
1. [Authentication & Onboarding](#authentication--onboarding)
2. [Dashboard & Home](#dashboard--home)
3. [User Profile & Settings](#user-profile--settings)
4. [Workspaces & Team Management](#workspaces--team-management)
5. [Instagram Integration](#instagram-integration)
6. [AI Assistant (Otto)](#ai-assistant-otto)
7. [Subscription & Billing](#subscription--billing)
8. [Shared Components](#shared-components)

---

## Authentication & Onboarding

### 1. Login Page
**Route:** `/login` or `/auth/login`  
**Purpose:** User authentication via password or OTP  
**Backend Endpoints Used:**
- `POST /api/auth/login` - Password login
- `POST /api/auth/otp/send` - Send OTP
- `POST /api/auth/otp/verify` - Verify OTP and login

**Page Components:**
```
LoginPage
├── LoginForm
│   ├── Email input
│   ├── Password input (for password login)
│   ├── "Login" button
│   ├── "Forgot Password?" link
│   └── "Sign Up" link
├── OR Divider
├── OTPForm (toggle-able)
│   ├── Email input
│   ├── "Send OTP" button
│   ├── OTP input (appears after OTP sent)
│   ├── "Verify & Login" button
│   └── Countdown timer for OTP resend
└── Social Login (future: Google, GitHub)
```

**Data Flow:**
1. User enters email + password → `POST /api/auth/login`
2. On success → Save JWT tokens (access + refresh) → Redirect to `/dashboard`
3. On error → Show error message, allow retry
4. OTP flow: Email → `POST /api/auth/otp/send` → User receives OTP email → Enter OTP → `POST /api/auth/otp/verify`

**State Management:**
- Store access token in secure cookie (HttpOnly)
- Store user info in context/state
- Auto-redirect if already logged in

---

### 2. Registration Page
**Route:** `/register` or `/auth/register`  
**Purpose:** Create new user account  
**Backend Endpoints Used:**
- `POST /api/auth/register` - User registration

**Page Components:**
```
RegisterPage
├── RegistrationForm
│   ├── First Name input
│   ├── Last Name input
│   ├── Username input
│   ├── Email input
│   ├── Password input
│   ├── Confirm Password input
│   ├── Terms & Privacy checkbox
│   ├── "Create Account" button
│   └── "Already have an account? Login" link
├── Progress indicator (optional)
└── Email verification notice
```

**Data Flow:**
1. User fills form → Validate locally (email format, password strength, username available)
2. Submit → `POST /api/auth/register` with `{ first_name, last_name, username, email, password }`
3. On success:
   - Show "Verification email sent" message
   - Optionally auto-navigate to `/verify-email` or show instructions
   - User receives verification email with token link
4. On error:
   - Username already taken → Show inline error
   - Email already registered → Show suggestion to login
   - Password too weak → Show requirements
   - Server error → Show retry option

**Validation:**
- Email format (RFC 5322)
- Password strength (min 8 chars, uppercase, lowercase, number, special char)
- Username format (alphanumeric, underscores, 3-20 chars)
- Terms acceptance required

---

### 3. Email Verification Page
**Route:** `/verify-email` or `/verify-email/:token`  
**Purpose:** Verify user email during registration  
**Backend Endpoints Used:**
- `GET /api/auth/verify/email/resend` - Resend verification email

**Page Components:**
```
EmailVerificationPage
├── Message: "Check your email to verify your account"
├── Verification link auto-detector (if token in URL)
├── OR Manual token input
│   ├── Paste verification link or token
│   ├── "Verify" button
├── "Didn't receive email?" section
│   ├── "Resend verification email" button
│   └── Countdown timer (resend available in 60s)
└── Auto-redirect on success to login or onboarding
```

**Data Flow:**
1. User receives email with verification link: `https://app.refyne.io/verify-email?token=xyz`
2. Frontend detects token in URL → Automatically verify (no user action needed)
3. On successful verification → Show success message → Auto-redirect to `/login` or `/onboarding`
4. If resend needed → `POST /api/auth/verify/email/resend` → Show confirmation message

---

### 4. Password Reset Request Page
**Route:** `/forgot-password` or `/auth/password-reset`  
**Purpose:** Initiate password reset flow  
**Backend Endpoints Used:**
- `POST /api/auth/password/reset/request` - Request password reset

**Page Components:**
```
PasswordResetPage
├── Step 1: Email Entry
│   ├── Email input
│   ├── "Send Reset Link" button
│   ├── "Back to Login" link
│   └── Error/success messages
├── Step 2: Confirmation Message (after submit)
│   ├── "Reset link sent to email"
│   ├── "Check your email" message
│   └── "Resend" button (after cooldown)
```

**Data Flow:**
1. User enters email → `POST /api/auth/password/reset/request` with `{ email }`
2. Backend sends email with reset token link
3. Show confirmation message (don't reveal if email exists - security)
4. User clicks email link → Navigate to `/reset-password?token=xyz`

---

### 5. Password Reset Confirmation Page
**Route:** `/reset-password` or `/reset-password/:token`  
**Purpose:** Set new password  
**Backend Endpoints Used:**
- `POST /api/auth/password/reset/confirm` - Confirm password reset

**Page Components:**
```
PasswordResetConfirmPage
├── Token validation indicator
├── New Password input
├── Confirm Password input
├── Password strength indicator
├── "Reset Password" button
├── "Back to Login" link
└── Error/success messages
```

**Data Flow:**
1. User receives email with reset link: `https://app.refyne.io/reset-password?token=xyz`
2. Frontend extracts token from URL
3. User enters new password twice
4. Submit → `POST /api/auth/password/reset/confirm` with `{ token, new_password }`
5. On success → Show "Password reset successfully" → Redirect to `/login`
6. On error (token expired, invalid) → Show error + link to request new reset

---

### 6. Onboarding Page
**Route:** `/onboarding`  
**Purpose:** Complete user onboarding after registration/first login  
**Backend Endpoints Used:**
- `GET /api/user/profile` - Get user profile
- `PUT /api/user/profile` - Update profile
- `POST /api/user/onboarding/complete` - Mark onboarding as complete
- `POST /api/user/settings` - Set user preferences

**Page Components:**
```
OnboardingPage (Multi-step form)
├── Step 1: Welcome
│   ├── Greeting message
│   ├── App overview
│   └── "Get Started" button
├── Step 2: Profile Completion
│   ├── Avatar upload
│   ├── First Name (prefilled from registration)
│   ├── Last Name (prefilled from registration)
│   ├── Username (prefilled from registration)
│   ├── Bio/About (optional)
│   └── "Continue" button
├── Step 3: Preferences
│   ├── Language selector
│   ├── Timezone selector
│   ├── Email notification preferences
│   └── "Continue" button
├── Step 4: Connect Instagram (optional)
│   ├── "Connect Instagram Account" button
│   ├── Instructions
│   └── "Skip for now" link
├── Step 5: Plan Selection (simplified to single Pro tier)
│   ├── Pro tier details
│   ├── Pricing display
│   ├── "Start 14-day free trial" button (if applicable)
│   └── "I have a coupon" link
└── Completion: "Setup Complete" → Redirect to `/dashboard`
```

**Data Flow:**
1. User enters onboarding → Fetch `GET /api/user/profile` (prefill data)
2. Step 1: Just informational
3. Step 2: Update profile → `PUT /api/user/profile` with `{ first_name, last_name, username, ... }`
4. Step 3: Save preferences → `POST /api/user/settings` with `{ language, timezone, email_notifications }`
5. Step 4: Optional Instagram connect → Redirect to OAuth (see Instagram section)
6. Step 5: Show subscription plan → `POST /api/subscription/checkout` for Pro tier
7. Final: `POST /api/user/onboarding/complete` → Mark onboarding done
8. Redirect to `/dashboard`

**State:**
- Track current step
- Auto-save progress (resume if user closes)
- Don't require Instagram setup to proceed

---

## Dashboard & Home

### 7. Main Dashboard
**Route:** `/dashboard` or `/`  
**Purpose:** Main hub after login - shows workspaces, recent activity, quick stats  
**Backend Endpoints Used:**
- `GET /api/workspaces` - List user's workspaces
- `GET /api/health/detailed` - Check service health
- `GET /api/user/profile` - Get current user info

**Page Components:**
```
DashboardPage
├── Header
│   ├── Logo
│   ├── Search bar (future: search posts, contacts)
│   ├── Notifications bell icon
│   ├── User profile dropdown
│   └── Settings icon
├── Sidebar Navigation
│   ├── Home
│   ├── Workspaces (expandable list)
│   ├── Instagram (if connected)
│   ├── AI Assistant (Otto)
│   ├── Subscription
│   └── Settings
├── Main Content
│   ├── Welcome message (hi {name})
│   ├── Quick stats cards
│   │   ├── Connected accounts
│   │   ├── Posts this month
│   │   ├── Engagement rate
│   │   └── Messages in inbox
│   ├── Recent activity
│   │   ├── Last 5 posts
│   │   ├── Team invitations pending
│   │   └── Recent AI insights
│   ├── Quick actions
│   │   ├── "Connect Instagram"
│   │   ├── "Start AI Analysis"
│   │   ├── "Invite Team Member"
│   │   └── "View Subscription"
│   └── Workspaces section
│       ├── List of workspaces (cards)
│       ├── "+ Create Workspace" button
│       └── Quick access to workspace features
```

**Data Flow:**
1. Page loads → `GET /api/user/profile` (get user name)
2. Fetch → `GET /api/workspaces` (list all workspaces)
3. Display workspaces in grid/list
4. Show quick stats (aggregated from workspace data)
5. Show recent activity (from last login)

**Conditional Rendering:**
- If no workspaces → Show "Create your first workspace" prompt
- If Instagram not connected → Show "Connect Instagram" CTA
- If onboarding not complete → Redirect to `/onboarding`

---

### 8. Workspace Dashboard
**Route:** `/workspaces/:id` or `/workspace/:id/home`  
**Purpose:** Workspace-specific dashboard and activity hub  
**Backend Endpoints Used:**
- `GET /api/workspaces/:id` - Get workspace details
- `GET /api/workspaces/:id/members` - List workspace members
- `GET /api/instagram/analytics` - Get Instagram analytics (if connected)

**Page Components:**
```
WorkspaceDashboardPage
├── Header
│   ├── Workspace name
│   ├── Workspace icon/avatar
│   └── Workspace settings icon
├── Navigation Tabs/Sidebar
│   ├── Overview (current)
│   ├── Instagram Feed
│   ├── Team
│   ├── AI Assistant
│   ├── Analytics
│   └── Settings
├── Main Content
│   ├── Workspace overview
│   │   ├── Member count
│   │   ├── Instagram accounts linked
│   │   ├── Subscription status
│   │   └── Storage used
│   ├── Team members section
│   │   ├── List of members with roles
│   │   ├── "+ Invite Member" button
│   │   └── Member management options
│   ├── Instagram feed preview
│   │   ├── Recent posts (3-5)
│   │   ├── View analytics link
│   │   └── "+ Post" button (future)
│   ├── AI insights preview
│   │   ├── Last AI conversation
│   │   └── "Chat with AI" link
│   └── Quick actions
│       ├── Edit workspace
│       ├── Manage team
│       ├── View subscription
│       └── Leave workspace (if member)
```

**Data Flow:**
1. User clicks workspace → Fetch `GET /api/workspaces/:id`
2. Fetch `GET /api/workspaces/:id/members` (display team)
3. Fetch Instagram data if connected
4. Display workspace-specific stats and activity
5. Show role-based options (Owner sees settings, Members see limited options)

---

## User Profile & Settings

### 9. User Profile Page
**Route:** `/settings/profile` or `/user/profile`  
**Purpose:** View and edit user profile information  
**Backend Endpoints Used:**
- `GET /api/user/profile` - Get user profile
- `PUT /api/user/profile` - Update profile

**Page Components:**
```
ProfilePage
├── Profile Header
│   ├── Avatar (with upload button)
│   ├── User name
│   ├── Username (@username)
│   ├── Email
│   └── Member since date
├── Profile Form (Editable sections)
│   ├── First Name input
│   ├── Last Name input
│   ├── Username input (with availability check)
│   ├── Bio/About textarea
│   ├── Website URL input
│   ├── Profile visibility toggle
│   ├── "Save Changes" button
│   └── "Cancel" button
└── Profile Preview
    └── How profile appears to others
```

**Data Flow:**
1. Page loads → `GET /api/user/profile` (prefill form)
2. User edits fields
3. Real-time validation (username availability check)
4. Submit → `PUT /api/user/profile` with updated data
5. Show success message
6. Update local state and header display

---

### 10. User Settings Page
**Route:** `/settings`  
**Purpose:** Manage user preferences and account settings  
**Backend Endpoints Used:**
- `GET /api/user/settings` - Get user settings
- `PUT /api/user/settings` - Update settings

**Page Components:**
```
SettingsPage
├── Settings Navigation (Sidebar)
│   ├── General
│   ├── Preferences
│   ├── Security
│   ├── Notifications
│   ├── Privacy
│   └── Account
├── Main Content Area
│   ├── [General Section]
│   │   ├── Language selector (en, es, fr, de, etc.)
│   │   ├── Timezone selector
│   │   ├── Date format selector
│   │   ├── Theme toggle (light/dark)
│   │   └── "Save" button
│   ├── [Preferences Section]
│   │   ├── Email notification toggles
│   │   │   ├── Newsletter
│   │   │   ├── Team invitations
│   │   │   ├── New messages
│   │   │   └── Product updates
│   │   ├── Frequency selector (immediate, daily, weekly)
│   │   └── "Save" button
│   ├── [Security Section]
│   │   ├── Change password
│   │   │   ├── Current password input
│   │   ├── New password input
│   │   ├── Confirm password input
│   │   └── "Update Password" button
│   │   ├── Two-factor authentication (future)
│   │   ├── Active sessions list
│   │   └── "Logout all devices" button
│   ├── [Privacy Section]
│   │   ├── Profile visibility
│   │   ├── Show email to team members
│   │   ├── Allow contact from others
│   │   └── "Save" button
│   └── [Account Section]
│       ├── Email address (with change option)
│       ├── Account created date
│       ├── Storage used / quota
│       ├── "Download my data" button (GDPR)
│       ├── "Deactivate account" button
│       └── "Delete account" button (with confirmation)
```

**Data Flow:**
1. Page loads → `GET /api/user/settings` (prefill all sections)
2. User edits section → Live updates
3. Each save → `PUT /api/user/settings` with updated section
4. Show success notification
5. Account deletion → Confirmation dialog → `DELETE /api/user/account` → Logout

**Conditional Rendering:**
- Show theme toggle only if frontend supports dark mode
- Security section shows only after login verification
- Delete account requires password confirmation

---

## Workspaces & Team Management

### 11. Workspaces List Page
**Route:** `/workspaces` or `/workspaces/list`  
**Purpose:** View all user workspaces and create new ones  
**Backend Endpoints Used:**
- `GET /api/workspaces` - List all workspaces
- `POST /api/workspaces` - Create new workspace

**Page Components:**
```
WorkspacesListPage
├── Header
│   ├── "My Workspaces" title
│   ├── "+ Create Workspace" button
│   └── Search/filter workspaces
├── Workspaces Grid (or List)
│   └── For each workspace:
│       ├── Workspace avatar/icon
│       ├── Workspace name
│       ├── Member count
│       ├── Your role (Owner/Member)
│       ├── Subscription status
│       ├── Last active date
│       ├── "Open" link
│       ├── "Settings" icon (if Owner)
│       ├── "Leave" button (if Member)
│       └── "..." menu (more options)
├── Empty state (if no workspaces)
│   ├── Illustration
│   ├── "Create your first workspace" message
│   └── "+ Create Workspace" button
└── Workspace creation modal (see below)
```

**Data Flow:**
1. Page loads → `GET /api/workspaces` (list all)
2. Display workspaces in cards/list
3. User clicks workspace → Navigate to `/workspaces/:id`
4. Click "Create Workspace" → Show modal

---

### 12. Create/Edit Workspace Modal
**Route:** Modal on `/workspaces` page  
**Purpose:** Create or edit workspace  
**Backend Endpoints Used:**
- `POST /api/workspaces` - Create workspace
- `PUT /api/workspaces/:id` - Update workspace
- `DELETE /api/workspaces/:id` - Delete workspace (in edit mode)

**Modal Components:**
```
WorkspaceModal
├── Modal Title ("Create Workspace" or "Edit Workspace")
├── Workspace Name input (required)
├── Workspace Description textarea
├── Workspace Icon/Avatar upload
├── Privacy setting
│   ├── Private (only invited members)
│   ├── Public (discoverable)
│   └── Internal (team only)
├── Action Buttons
│   ├── "Create"/"Save" button
│   ├── "Cancel" button
│   └── "Delete Workspace" button (edit mode only, Owner only)
└── Error/success messages
```

**Data Flow:**
1. Create mode: User fills form → `POST /api/workspaces` with `{ name, description, ... }`
2. Edit mode: Prefill from workspace data → User edits → `PUT /api/workspaces/:id`
3. Delete: Show confirmation → `DELETE /api/workspaces/:id` → Refresh list
4. On success → Close modal, refresh workspaces list

---

### 13. Team Management Page
**Route:** `/workspaces/:id/team` or `/workspace/:id/settings/team`  
**Purpose:** Manage workspace team members and permissions  
**Backend Endpoints Used:**
- `GET /api/workspaces/:id/members` - List members
- `POST /api/workspaces/:id/members` - Invite member
- `DELETE /api/workspaces/:id/members/:user_id` - Remove member

**Page Components:**
```
TeamManagementPage
├── Header
│   ├── "Team Management" title
│   ├── Member count
│   └── "+ Invite Member" button
├── Members List
│   └── For each member:
│       ├── Avatar
│       ├── Name
│       ├── Email
│       ├── Role (Owner/Member) - with role selector if Owner
│       ├── Status (Active/Pending/Removed)
│       ├── Joined date
│       ├── Last active
│       └── Remove button (Owner only)
├── Pending Invitations
│   └── For each pending invitation:
│       ├── Email
│       ├── Invited by
│       ├── Invited date
│       ├── Expires date (if applicable)
│       ├── "Resend" button
│       └── "Revoke" button
├── Team Activity Log (future)
│   ├── Member joined
│   ├── Member removed
│   ├── Role changed
│   └── Invitation sent
└── Team Settings
    ├── Allow member invitations (toggle)
    ├── Require email verification
    └── Default role for new members
```

**Data Flow:**
1. Page loads → `GET /api/workspaces/:id/members` (list all members)
2. Owner sees "Invite Member" button
3. Click "Invite Member" → Show modal:
   ```
   ├── Email input (multiple emails with comma/new line)
   ├── Role selector (Member/Admin)
   ├── Message (optional)
   ├── "Send Invitations" button
   └── "Cancel" button
   ```
4. Submit → `POST /api/workspaces/:id/members` with invited emails and role
5. Show success message with invitation status
6. Member emails get invite link: `https://app.refyne.io/workspaces/join?token=xyz`
7. Remove member → Confirmation dialog → `DELETE /api/workspaces/:id/members/:user_id`

---

### 14. Workspace Join Page
**Route:** `/workspaces/join?token=xyz`  
**Purpose:** Accept workspace membership invitation  
**Backend Endpoints Used:**
- Email token verification (backend handles)
- Workspace join endpoint (backend auto-accepts via token)

**Page Components:**
```
WorkspaceJoinPage
├── Verification in progress indicator
├── Message: "Joining {workspace_name}..."
├── Auto-redirect to workspace dashboard on success
└── Error message with retry option
```

**Data Flow:**
1. User clicks email invite link → Navigate to `/workspaces/join?token=xyz`
2. Frontend extracts token
3. Frontend calls backend verification endpoint
4. Auto-accept invitation
5. On success → Redirect to `/workspaces/:id`
6. On error → Show error message and retry button

---

## Instagram Integration

### 15. Instagram Connection Page
**Route:** `/instagram` or `/settings/instagram`  
**Purpose:** Connect/disconnect Instagram account  
**Backend Endpoints Used:**
- `GET /api/instagram/auth/url` - Get OAuth login URL
- `POST /api/instagram/auth/callback` - Handle OAuth callback
- `GET /api/instagram/media` - List connected media

**Page Components:**
```
InstagramConnectionPage
├── Header
│   ├── "Instagram Integration" title
│   └── Help icon with instructions
├── Connection Status
│   ├── If NOT connected:
│   │   ├── "Connect Your Instagram Account" heading
│   │   ├── Instagram logo
│   │   ├── "Connect with Instagram" button
│   │   ├── Feature list (what you can do):
│   │   │   ├── View all your posts
│   │   │   ├── Get AI insights
│   │   │   ├── Analyze engagement
│   │   │   └── Monitor comments
│   │   └── Privacy notice
│   ├── If connected:
│   │   ├── "Connected Account: @username" display
│   │   ├── Profile picture
│   │   ├── Account type (Personal/Business)
│   │   ├── Follower count
│   │   ├── Bio
│   │   ├── Last sync time
│   │   ├── Sync button
│   │   ├── "Disconnect" button
│   │   └── "Reconnect with different account" option
├── Connected Accounts List (if multiple)
│   ├── Add new account
│   └── Manage each account
└── Sync Settings
    ├── Auto-sync toggle
    ├── Sync frequency (every 1h, 6h, 24h)
    └── Data retention period
```

**Data Flow:**
1. Page loads → Check if Instagram connected
2. If not connected:
   - User clicks "Connect" → Fetch `GET /api/instagram/auth/url`
   - Redirect to Instagram OAuth login
   - Instagram redirects back to `/instagram/callback?code=xyz`
3. Handle callback → `POST /api/instagram/auth/callback` with auth code
4. Backend stores OAuth token
5. Frontend redirects to `/instagram` with success message
6. If connected:
   - Show account details
   - Offer sync button → `GET /api/instagram/media` (refresh media list)
   - Show disconnect option

---

### 16. Instagram Feed Page
**Route:** `/instagram/feed` or `/workspaces/:id/instagram`  
**Purpose:** View and manage Instagram media  
**Backend Endpoints Used:**
- `GET /api/instagram/media` - List all media
- `GET /api/instagram/media/:id` - Get media details
- `POST /api/instagram/media/sync` - Trigger media sync

**Page Components:**
```
InstagramFeedPage
├── Header
│   ├── "Instagram Feed" title
│   ├── Account selector (if multiple accounts)
│   ├── Sync button (with last sync time)
│   ├── Sort/filter options
│   │   ├── Sort by (newest, oldest, most liked)
│   │   ├── Filter by type (photo, video, carousel, story)
│   │   ├── Date range picker
│   │   └── Search media by caption/hashtag
│   └── View toggle (grid, list, timeline)
├── Media Grid/List
│   └── For each media item:
│       ├── Media thumbnail/preview
│       ├── Media type icon (photo, video)
│       ├── Caption (truncated)
│       ├── Engagement stats
│       │   ├── Likes count
│       │   ├── Comments count
│       │   ├── Shares count
│       │   └── Saves count
│       ├── Posted date
│       ├── Click to expand (see details)
│       └── AI analysis button (see AI Assistant page)
├── Empty State (if no media)
│   ├── Illustration
│   ├── "No posts found" message
│   └── "Connect Instagram to see your posts" link
├── Pagination/Infinite scroll
│   └── Load more button or auto-load on scroll
└── Bulk Actions (future)
    ├── Select multiple posts
    ├── Archive selected
    ├── Analyze selected
    └── Export selected
```

**Data Flow:**
1. Page loads → `GET /api/instagram/media` (list all posts, paginated)
2. Display media in grid
3. User filters/sorts → Update query → Fetch new data
4. User clicks media → Show modal with details (see below)
5. Click sync → `POST /api/instagram/media/sync` → Show progress → Refresh feed

**Media Details Modal:**
```
MediaDetailsModal
├── Full-size media preview
├── Caption text
├── Post URL link
├── Engagement statistics
│   ├── Likes, comments, shares, saves
│   ├── Engagement rate calculation
│   └── Comparison to average
├── Comments section (top comments)
│   ├── Comment list
│   ├── Total comments count
│   └── "View all" link
├── Posted date and time
├── Media location (if tagged)
├── Hashtags (clickable)
├── Tagged users (clickable)
├── "Analyze with AI" button (links to Otto)
├── "Share post" button (copy link)
└── Close button
```

---

### 17. Instagram Analytics Page
**Route:** `/instagram/analytics` or `/workspaces/:id/analytics`  
**Purpose:** View Instagram account and post analytics  
**Backend Endpoints Used:**
- `GET /api/instagram/analytics` - Get analytics data

**Page Components:**
```
InstagramAnalyticsPage
├── Header
│   ├── "Analytics" title
│   ├── Account selector (if multiple)
│   ├── Date range picker (last 7d, 30d, 90d, custom)
│   └── Export report button
├── Account-Level Stats
│   ├── Follower growth chart (line graph over time)
│   ├── Engagement rate (overall metric)
│   ├── Average post performance
│   ├── Most engaged post
│   └── Follower demographics (age, location, gender)
├── Post Performance
│   ├── Top performing posts (table or cards)
│   │   ├── Post thumbnail
│   │   ├── Engagement metrics
│   │   ├── Reach and impressions
│   │   └── Performance rank
│   └── Post type breakdown (pie chart)
│       ├── Photos, Videos, Carousel
│       ├── Reels performance
│       └── Stories performance
├── Audience Insights
│   ├── Most active times (heatmap)
│   ├── Top hashtags used
│   ├── Top mentioned accounts
│   └── Content themes breakdown
├── Engagement Metrics
│   ├── Likes trend
│   ├── Comments trend
│   ├── Saves trend
│   └── Shares trend
└── AI Recommendations (from Otto AI)
    ├── "Best times to post"
    ├── "Top performing content types"
    ├── "Suggested hashtags"
    └── "Engagement opportunities"
```

**Data Flow:**
1. Page loads → `GET /api/instagram/analytics` with date range
2. Display charts and stats
3. User changes date range → Refetch data
4. User clicks "Export" → Download report (CSV or PDF)

---

## AI Assistant (Otto)

### 18. AI Assistant Chat Page
**Route:** `/otto` or `/ai-assistant` or `/workspaces/:id/ai`  
**Purpose:** Chat with AI assistant for insights and analysis  
**Backend Endpoints Used:**
- `POST /api/otto/conversations` - Create conversation
- `GET /api/otto/conversations` - List conversations
- `GET /api/otto/conversations/:id` - Get conversation
- `POST /api/otto/conversations/:id/messages` - Send message
- `GET /api/otto/conversations/:id/messages` - Get messages
- `POST /api/otto/conversations/:id/feedback` - Provide feedback

**Page Components:**
```
OttoAIPage
├── Sidebar
│   ├── "Conversations" heading
│   ├── Search conversations
│   ├── "+ New Conversation" button
│   ├── Conversation List
│   │   └── For each conversation:
│   │       ├── Title (auto-generated from first message)
│   │       ├── Last message preview
│   │       ├── Date/time
│   │       ├── Pin icon (pin important conversations)
│   │       └── Delete icon (with confirmation)
│   └── Conversation Filters
│       ├── All
│       ├── Pinned
│       ├── Archived
│       └── Starred
├── Main Chat Area
│   ├── Header
│   │   ├── Conversation title
│   │   ├── Last updated time
│   │   ├── Options menu (rename, archive, delete)
│   │   └── Info icon (show context)
│   ├── Message Thread
│   │   ├── For each message:
│   │   │   ├── Avatar (user/AI)
│   │   │   ├── Sender name
│   │   │   ├── Timestamp
│   │   │   ├── Message content (markdown supported)
│   │   │   ├── Reaction buttons (👍, 👎)
│   │   │   ├── Feedback button (if AI message)
│   │   │   └── Copy/Share/Delete options
│   │   └── Auto-scroll to latest message
│   ├── Input Area
│   │   ├── Text input with placeholder: "Ask Otto anything..."
│   │   ├── File attachment button (images, docs for context)
│   │   ├── Emoji picker
│   │   ├── Send button
│   │   ├── Voice input button (future)
│   │   └── Suggested prompts
│   │       ├── "Analyze my engagement"
│   │       ├── "Best times to post"
│   │       ├── "Content ideas"
│   │       └── "Audience insights"
│   └── Typing Indicator (when AI is responding)
│       └── Shows "Otto is thinking..."
└── Empty State (when no conversations)
    ├── Otto greeting
    ├── Suggested questions
    └── "Start a conversation" button
```

**Data Flow:**
1. Page loads → `GET /api/otto/conversations` (list all conversations)
2. Display conversation list in sidebar
3. User clicks conversation → `GET /api/otto/conversations/:id` → Load messages
4. User types message + sends → `POST /api/otto/conversations/:id/messages` with `{ content, context_ids }`
5. Show message immediately (optimistic update)
6. Backend processes → AI responds → Real-time update (WebSocket or polling)
7. Display AI response with timestamp
8. User clicks feedback → `POST /api/otto/conversations/:id/feedback` with `{ message_id, rating, comment }`

**Conversation Context:**
- User can attach Instagram media for analysis
- User can reference previous messages
- User can upload documents or screenshots for context
- Context documents are stored and reused

---

### 19. AI Analysis Modal
**Route:** Modal from Instagram feed or blog page  
**Purpose:** Quick AI analysis of specific content  
**Backend Endpoints Used:**
- `POST /api/otto/conversations` - Create analysis conversation
- `POST /api/otto/conversations/:id/messages` - Send analysis request

**Modal Components:**
```
AIAnalysisModal
├── Header
│   ├── "Analyze with Otto" title
│   ├── Close button
│   └── Conversation link (open in main chat)
├── Content Preview
│   ├── Instagram media or blog post preview
│   ├── Caption/content snippet
│   └── Basic stats
├── Analysis Options
│   ├── "Engagement Analysis" option
│   ├── "Audience Sentiment" option
│   ├── "Content Optimization" option
│   ├── "Trend Analysis" option
│   └── Custom question input
├── Analysis Results
│   ├── Loading state with animation
│   ├── AI response (streamed or chunked)
│   ├── Suggestions list
│   ├── Metrics breakdown
│   └── "Ask follow-up question" input
└── Actions
    ├── "Save to conversation" button
    ├── "Export analysis" button
    └── "Share" button (generate link)
```

**Data Flow:**
1. User clicks "Analyze" on Instagram post → Show modal
2. Select analysis type or enter custom question
3. Submit → `POST /api/otto/conversations` (create new)
4. → `POST /api/otto/conversations/:id/messages` with media context
5. Stream AI response in real-time
6. Display analysis results
7. User can ask follow-ups or return to main chat

---

## Subscription & Billing

### 20. Subscription Page
**Route:** `/subscription` or `/settings/subscription` or `/billing`  
**Purpose:** View subscription status and manage billing  
**Backend Endpoints Used:**
- `GET /api/subscription/status` - Get current subscription
- `POST /api/subscription/checkout` - Create checkout session
- `POST /api/subscription/cancel` - Cancel subscription

**Page Components:**
```
SubscriptionPage
├── Current Plan Section
│   ├── Plan name (Pro)
│   ├── Price display (with billing cycle)
│   ├── Features list
│   │   ├── ✓ Connected Instagram accounts
│   │   ├── ✓ AI assistant (Otto)
│   │   ├── ✓ Team members
│   │   ├── ✓ Advanced analytics
│   │   └── ✓ Priority support
│   ├── Next billing date
│   ├── Billing status (Active, Canceled, Expired)
│   ├── "Manage Subscription" button
│   └── "Invoice History" link
├── Plan Comparison (if applicable)
│   ├── "Upgrade to Business" CTA (future)
│   ├── Feature comparison table
│   ├── Pricing comparison
│   └── "Learn more" link
├── Billing Information
│   ├── Current billing method
│   ├── Cardholder name
│   ├── Last 4 digits
│   ├── Expiration date
│   ├── "Update payment method" button
│   └── "Change billing address" link
├── Invoices Section
│   ├── "Invoice History" heading
│   ├── Invoice list
│   │   └── For each invoice:
│   │       ├── Invoice number
│   │       ├── Date
│   │       ├── Amount
│   │       ├── Status (Paid, Pending)
│   │       ├── "Download" button
│   │       └── "View details" link
│   ├── Pagination
│   └── Export all invoices
├── Subscription Management
│   ├── "Cancel Subscription" button
│   ├── "Pause Subscription" button (if available)
│   └── "Contact Support" link
└── Billing Contact Info
    ├── Email input
    ├── Company name input
    ├── Address input
    ├── Tax ID input
    └── "Save billing info" button
```

**Data Flow:**
1. Page loads → `GET /api/subscription/status` (fetch current plan)
2. Display current subscription details
3. Show next billing date
4. Show invoice history
5. User clicks "Cancel" → Show confirmation dialog with reasons → `POST /api/subscription/cancel`
6. On success → Show "Subscription will be canceled at end of period" message

**Cancel Subscription Flow:**
```
CancelDialog
├── "Are you sure?" heading
├── "We'd love to know why you're leaving"
├── Radio buttons for reasons
│   ├── Too expensive
│   ├── Don't use all features
│   ├── Found alternative
│   ├── Technical issues
│   └── Other (text input)
├── Feedback textarea
├── "Keep my subscription" button
├── "Cancel subscription" button
└── "Contact support first" link
```

---

### 21. Checkout Page (Paddle Integration)
**Route:** `/checkout` or `/billing/checkout`  
**Purpose:** Complete subscription purchase  
**Backend Endpoints Used:**
- `POST /api/subscription/checkout` - Create checkout session

**Page Components:**
```
CheckoutPage
├── Left Panel (Order Summary)
│   ├── Plan name (Pro)
│   ├── Pricing breakdown
│   │   ├── Base price
│   │   ├── Tax calculation
│   │   ├── Discount (if applicable)
│   │   └── Total
│   ├── Billing cycle toggle (Monthly/Yearly)
│   ├── Savings display (if yearly discount)
│   ├── Features included
│   ├── Discount code input
│   └── Apply button
├── Right Panel (Paddle Checkout)
│   ├── Hosted Paddle checkout form
│   │   ├── Email input (prefilled)
│   │   ├── Billing information fields
│   │   ├── Payment method selection
│   │   ├── Payment details (card)
│   │   └── Terms & conditions checkbox
│   └── "Subscribe Now" button
└── Loading/Processing State
    ├── Processing animation
    └── "Redirecting..." message
```

**Data Flow:**
1. User selects plan → Click "Subscribe"
2. Frontend calls `POST /api/subscription/checkout` with `{ plan_id, billing_cycle }`
3. Backend returns Paddle checkout URL or embeds checkout form
4. Redirect to Paddle checkout or embed iframe
5. User completes payment on Paddle
6. Paddle redirects back to `/checkout/success` or similar
7. Backend webhook processes payment
8. Frontend redirects to `/dashboard` with success message

---

## Shared Components

### 22. Navigation Bar / Header
**Used on:** All pages after login  
**Components:**
```
Navbar
├── Left Side
│   ├── Logo (link to dashboard)
│   └── Main navigation (expandable on mobile)
│       ├── Dashboard
│       ├── Workspaces
│       ├── Instagram
│       ├── AI Assistant
│       └── Billing
├── Center Side
│   └── Search bar (global search)
│       ├── Search across posts, contacts, messages
│       ├── Keyboard shortcut: Cmd/Ctrl + K
│       └── Recent searches
├── Right Side
│   ├── Notifications bell
│   │   ├── Badge with count
│   │   ├── Dropdown with recent notifications
│   │   ├── Mark as read
│   │   └── View all link
│   ├── Help icon (with docs links)
│   ├── Settings dropdown
│   │   ├── Profile settings
│   │   ├── Account settings
│   │   ├── Preferences
│   │   ├── Subscription
│   │   ├── Help & support
│   │   └── Logout
│   └── User avatar (clickable for dropdown)
└── Mobile Menu
    ├── Hamburger icon
    ├── Full navigation when open
    └── Close button
```

**Data Flow:**
- Always fetch user profile on app load
- Show notification badge (count from notifications service)
- Highlight active page in navigation

---

### 23. Sidebar / Navigation Menu
**Used on:** Dashboard and workspace pages  
**Components:**
```
Sidebar
├── Workspace Section
│   ├── Active workspace selector
│   ├── Workspace icon
│   ├── Workspace name
│   ├── Workspace switcher (expand/collapse)
│   └── Quick workspace switcher (dropdown)
├── Main Navigation
│   ├── Overview/Dashboard (with home icon)
│   ├── Workspaces (expandable)
│   ├── Instagram (with status indicator)
│   ├── AI Assistant (with unread count)
│   ├── Subscription (with status)
│   ├── Team (with member count)
│   ├── Analytics (with icon)
│   └── Settings (with gear icon)
├── Secondary Navigation
│   ├── Help & Documentation
│   ├── Contact Support
│   ├── Keyboard Shortcuts (Cmd+?)
│   └── Changelog (What's new)
├── Footer
│   ├── Refyne logo
│   ├── Version number
│   ├── Status page link
│   └── Terms & Privacy
└── Collapse/Expand Toggle
    └── Minimize sidebar on desktop
```

**Features:**
- Highlight current active page
- Show status indicators (connected, pending, error)
- Collapsible on mobile (hamburger menu)
- Keyboard navigation (arrow keys)

---

### 24. Loading States & Skeletons
**Used on:** All data-loading pages  
**Components:**
```
SkeletonLoader (for each content type)
├── Skeleton heading (shimmer animation)
├── Skeleton card (shimmer)
├── Skeleton list items (multiple)
└── Skeleton form fields

Loading Indicators
├── Inline spinner (small operations)
├── Full-page spinner (page navigation)
├── Progress bar (long operations)
└── Countdown timer (retry logic)
```

---

### 25. Error States & Retry Logic
**Used on:** All error scenarios  
**Components:**
```
ErrorBoundary
├── Error title
├── Error description
├── Error code (if applicable)
├── "Retry" button
├── "Contact support" button
└── Home/Dashboard link

Specific Errors
├── 401 Unauthorized → Redirect to login
├── 403 Forbidden → Show "Access denied"
├── 404 Not Found → Show "Page not found"
├── 500 Server Error → Show "Something went wrong"
├── Network Error → Show "Check your connection"
└── Rate Limit Error → Show "Too many requests, try again soon"
```

---

### 26. Modal / Dialog Components
**Used on:** Confirmations, forms, info  
**Components:**
```
Modal
├── Backdrop (semi-transparent)
├── Modal Container
│   ├── Header
│   │   ├── Title
│   │   └── Close button (X)
│   ├── Content
│   │   └── (varies by modal type)
│   └── Footer
│       ├── Action buttons (Primary, Secondary)
│       └── Close button
└── Animations
    ├── Fade in/out
    ├── Slide from bottom (mobile)
    └── Scale animation

Dialog Types
├── Confirmation dialog (delete, logout, cancel subscription)
├── Input dialog (rename, invite)
├── Info dialog (help, details)
└── Form dialog (edit profile, create workspace)
```

---

### 27. Toast / Notification System
**Used on:** All user actions  
**Components:**
```
Toast Notification
├── Icon (success, error, warning, info)
├── Message text
├── Close button
├── Auto-dismiss after 4-5 seconds
├── Stack multiple toasts
├── Position (top-right, top-center, bottom-right)
└── Animations (slide in, fade out)

Toast Types
├── Success: "Changes saved successfully"
├── Error: "Failed to update profile"
├── Warning: "This action cannot be undone"
└── Info: "Your subscription renews on..."
```

---

## Frontend Architecture Recommendations

### State Management
- **Global State:** Authentication (user, tokens), notifications
- **Local State:** Form inputs, UI toggles, pagination
- **Libraries:** Redux, Zustand, or Context API + hooks

### API Integration
- **HTTP Client:** Axios or Fetch with interceptors
- **Auth Tokens:** Store in secure cookies (HttpOnly), auto-refresh
- **Error Handling:** Catch 401/403 → redirect to login
- **Loading States:** Show skeleton loaders during data fetch

### Routing Structure
```
/                              (redirect to /dashboard)
/auth/
  /login                       (login page)
  /register                    (registration)
  /verify-email                (email verification)
  /forgot-password             (password reset request)
  /reset-password              (password reset confirm)
/onboarding                    (multi-step onboarding)
/dashboard                     (main hub)
/workspaces
  /                            (list workspaces)
  /:id                         (workspace dashboard)
  /:id/team                    (team management)
  /join                        (accept invitation)
/instagram
  /                            (connection status)
  /feed                        (media feed)
  /analytics                   (analytics dashboard)
/otto                          (AI assistant)
  /                            (chat interface)
  /:id                         (specific conversation)
/settings
  /profile                     (user profile)
  /preferences                 (user settings)
  /subscription                (subscription status)
  /security                    (security settings)
/subscription
  /checkout                    (payment checkout)
  /success                     (payment confirmation)
/error
  /404                         (not found)
  /500                         (server error)
```

### Environment Configuration
```env
VITE_API_BASE_URL=https://api.refyne.io
VITE_APP_NAME=Refyne
VITE_APP_VERSION=1.0.0
VITE_PADDLE_CLIENT_TOKEN=xxx
VITE_INSTAGRAM_CLIENT_ID=xxx
VITE_SENTRY_DSN=xxx
```

---

## Integration Checklist

### Pre-Development
- [ ] Backend API documented and running
- [ ] Postman collection or OpenAPI spec available
- [ ] Database seeded with test data
- [ ] CORS configured for frontend URLs
- [ ] JWT secret configured in .env

### Authentication Setup
- [ ] Login page connected to `/api/auth/login`
- [ ] Token storage (cookies + local state)
- [ ] Auto token refresh on 401 response
- [ ] Logout clears tokens and redirects to `/login`
- [ ] Protected routes require auth

### Integration Testing
- [ ] Test login flow (register → verify → login)
- [ ] Test profile updates (name, username, avatar)
- [ ] Test workspace creation and team invites
- [ ] Test Instagram OAuth flow
- [ ] Test payment checkout (Paddle sandbox)
- [ ] Test AI chat (messages and responses)
- [ ] Test error handling (network, 500, rate limits)

### Deployment Checklist
- [ ] Frontend environment variables configured
- [ ] API base URL points to production backend
- [ ] Error reporting configured (Sentry)
- [ ] Analytics configured (if needed)
- [ ] CDN configured for static assets
- [ ] SSL/TLS certificate valid
- [ ] CORS origins updated for production domain

---

**Document Version:** 1.0  
**Last Updated:** 2026-04-18  
**Status:** Ready for Frontend Development
