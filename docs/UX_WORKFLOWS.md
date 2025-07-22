# Refyne Platform - User Experience & Workflows

## Document Overview
This document defines the complete user experience flows and workflows for the Refyne platform, covering all user interactions from onboarding to advanced features.

---

## 1. User Onboarding Flow

### 1.1 Registration & Setup
```
Step 1: Landing Page
├── "Get Started" CTA
├── Pricing page review
└── "Start Free Trial" button

Step 2: Account Creation
├── Email & password input
├── Account verification email
└── Email confirmation click

Step 3: Onboarding Questionnaire
├── User type selection
├── Platform & follower info
├── Challenges & content type
└── Goal setting

Step 4: Workspace Setup
├── Workspace name creation
├── Instagram account connection
└── Initial context upload (optional)

Step 5: Welcome Tour
├── Dashboard overview
├── Key features introduction
└── First automation setup
```

### 1.2 Detailed Onboarding Steps

#### **Step 3: Onboarding Questionnaire**
```
Screen 1: "Tell us about yourself"
┌─────────────────────────────────────────┐
│ What best describes you?                │
│ ○ Individual Creator/Influencer         │
│ ○ Small Business Owner                  │
│ ○ Marketing Agency                      │
│ ○ Enterprise/Large Company             │
│ ○ Other                                 │
│                                         │
│        [Continue]                       │
└─────────────────────────────────────────┘

Screen 2: "Your social media presence"
┌─────────────────────────────────────────┐
│ What's your primary platform?           │
│ ○ Instagram                             │
│ ○ YouTube                               │
│ ○ TikTok                               │
│ ○ Multiple platforms                    │
│                                         │
│ How many followers do you have?         │
│ ○ Under 1K    ○ 1K-10K                 │
│ ○ 10K-100K   ○ 100K-1M                │
│ ○ 1M+                                  │
│                                         │
│        [Continue]                       │
└─────────────────────────────────────────┘

Screen 3: "What's your biggest challenge?"
┌─────────────────────────────────────────┐
│ ☐ Too many comments to respond to       │
│ ☐ Dealing with spam/toxic comments      │
│ ☐ Understanding audience sentiment      │
│ ☐ Converting engagement to sales        │
│ ☐ All of the above                      │
│                                         │
│        [Continue]                       │
└─────────────────────────────────────────┘
```

#### **Step 4: Instagram Connection**
```
Instagram OAuth Flow:
1. "Connect your Instagram account"
2. Instagram authorization popup
3. Permission acceptance
4. Account data sync begins
5. "Account connected successfully!"
6. Initial post sync (last 30 days)
```

---

## 2. Main Dashboard Experience

### 2.1 Dashboard Layout
```
┌─────────────────────────────────────────────────────────────────┐
│ REFYNE                          [Otto Chat] [Notifications] [⚙️] │
├─────────────────────────────────────────────────────────────────┤
│ 📊 Dashboard | 💬 Comments | 🤖 Automation | 📈 Analytics | 📚 Context
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ ┌─ Quick Stats ─────────────────────────────────────────────┐  │
│ │ Today: 23 comments | Sentiment: 😊 85% | Pending: 3      │  │
│ └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│ ┌─ Recent Activity ─────────┐ ┌─ Pending Actions ─────────┐    │
│ │ 2m ago: Auto-replied to   │ │ 🔍 3 comments need review │    │
│ │ @user about pricing       │ │ ✋ 2 responses waiting    │    │
│ │                           │ │    approval               │    │
│ │ 5m ago: Flagged toxic     │ │ 📝 1 new lead identified │    │
│ │ comment from @troll       │ │                           │    │
│ └───────────────────────────┘ └───────────────────────────┘    │
│                                                                 │
│ ┌─ Connected Accounts ──────────────────────────────────────┐  │
│ │ 📸 @fashionbrand_official (2.3K followers)               │  │
│ │    └─ 📊 15 new comments today | ⚡ Auto-reply: ON      │  │
│ └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Navigation Patterns
- **Quick Actions:** Always visible floating buttons for common tasks
- **Breadcrumbs:** Clear hierarchy (Workspace > Account > Post > Comment)
- **Contextual Sidebars:** Different sidebar content based on current view
- **Keyboard Shortcuts:** Power user features (J/K for navigation, R for reply)

---

## 3. Comment Management Workflows

### 3.1 Comment Moderation Interface
```
┌─────────────────────────────────────────────────────────────────┐
│ 💬 Comments Management                                          │
├─────────────────────────────────────────────────────────────────┤
│ Filters: [All] [Pending] [Flagged] [High Sentiment] [Leads]    │
│ Sort by: [Newest] [Sentiment] [Lead Score] [Engagement]        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ ┌─ Comment Card ────────────────────────────────────────────┐  │
│ │ 👤 @sarah_fashion • 2.1K followers • ✅ Verified         │  │
│ │ ⏰ 2 hours ago on "Summer Collection Post"                │  │
│ │                                                           │  │
│ │ "Omg this dress is STUNNING! 😍 Do you have it in blue?  │  │
│ │  Also what's your return policy?"                         │  │
│ │                                                           │  │
│ │ 🎯 Sentiment: 😊 Positive (0.85) | 🎪 Intent: Purchase   │  │
│ │ 🏆 Lead Score: 82/100 | 🏷️ Topics: product, inquiry     │  │
│ │                                                           │  │
│ │ 🤖 Otto suggests: "Product Inquiry Template"             │  │
│ │ 📝 "Hi Sarah! That dress would look amazing on you! We   │  │
│ │     have it in blue - I'll DM you the details! 💙"       │  │
│ │                                                           │  │
│ │ [✅ Approve] [✏️ Edit] [❌ Reject] [👁️ More Info]        │  │
│ └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Bulk Actions Interface
```
When multiple comments selected:
┌─────────────────────────────────────────┐
│ 5 comments selected                     │
│ [✅ Approve All] [❌ Reject All]        │
│ [🏷️ Add Tag] [📁 Mark as Lead]         │
│ [🚫 Block Users] [⚙️ Apply Rule]       │
└─────────────────────────────────────────┘
```

### 3.3 Comment Detail View
```
┌─────────────────────────────────────────────────────────────────┐
│ ← Back to Comments                                              │
├─────────────────────────────────────────────────────────────────┤
│ Comment Details                                                 │
│                                                                 │
│ 👤 User Profile                 📊 AI Analysis                 │
│ ├─ @sarah_fashion              ├─ Sentiment: +0.85             │
│ ├─ 2,143 followers             ├─ Toxicity: 0.02               │
│ ├─ Account age: 2 years        ├─ Intent: purchase_inquiry     │
│ ├─ Verification: ✅            ├─ Language: English            │
│ └─ Previous interactions: 3    └─ Confidence: 94%              │
│                                                                 │
│ 📝 Comment History                                              │
│ ├─ This comment (2h ago)                                       │
│ ├─ "Love your style!" (3 days ago) - No response               │
│ └─ "Where is this from?" (1 week ago) - Auto-replied           │
│                                                                 │
│ 🎯 Suggested Actions                                            │
│ ├─ Reply with "Product Inquiry" template                       │
│ ├─ Send DM with product details                                │
│ ├─ Add to CRM as "Warm Lead"                                   │
│ └─ Add to "Blue Dress Inquiries" campaign                      │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Automation Setup Workflows

### 4.1 Response Template Creation
```
Step 1: Template Basic Info
┌─────────────────────────────────────────┐
│ Create Response Template                │
│                                         │
│ Template Name: ________________         │
│ Description: ___________________        │
│                                         │
│ Template Type:                          │
│ ○ Comment Reply                         │
│ ○ DM Message                           │
│ ○ Email                                │
│                                         │
│        [Continue]                       │
└─────────────────────────────────────────┘

Step 2: Trigger Configuration
┌─────────────────────────────────────────┐
│ When should this template be used?      │
│                                         │
│ Keywords (comma separated):             │
│ [price, cost, how much, pricing]        │
│                                         │
│ Sentiment requirement:                  │
│ ○ Any sentiment                         │
│ ○ Positive only                         │
│ ○ Neutral/Positive only                │
│ ○ Negative only                         │
│                                         │
│ Intent type:                            │
│ ☐ Questions                             │
│ ☐ Purchase inquiries                    │
│ ☐ Complaints                            │
│ ☐ Compliments                           │
│                                         │
│        [Continue]                       │
└─────────────────────────────────────────┘

Step 3: Template Content
┌─────────────────────────────────────────┐
│ Template Message:                       │
│ ┌─────────────────────────────────────┐ │
│ │ Hi {username}! Thanks for asking    │ │
│ │ about pricing. I'll send you our    │ │
│ │ current rates in a DM right now! 📧 │ │
│ │                                     │ │
│ │ Available variables:                │ │
│ │ {username} - Commenter's name       │ │
│ │ {product} - Mentioned product       │ │
│ │ {brand} - Your brand name           │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Follow-up action:                       │
│ ○ None                                  │
│ ● Send DM with details                  │
│ ○ Add to email list                     │
│ ○ Add to CRM                           │
│                                         │
│        [Create Template]                │
└─────────────────────────────────────────┘
```

### 4.2 Moderation Rule Setup
```
Simple Rule Creator:
┌─────────────────────────────────────────┐
│ Create Moderation Rule                  │
│                                         │
│ Rule Name: ________________             │
│ Apply to: ○ All accounts                │
│           ● Specific account            │
│           ○ Specific posts              │
│                                         │
│ Rule Type:                              │
│ ○ Block specific words/phrases          │
│ ○ Toxicity threshold                    │
│ ○ Spam detection                        │
│ ● Competitor mentions                   │
│                                         │
│ Competitor names (one per line):        │
│ ┌─────────────────────────────────────┐ │
│ │ competitor_brand                    │ │
│ │ @competitor_handle                  │ │
│ │ rival_company                       │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Action: ○ Flag for review               │
│         ● Hide comment                  │
│         ○ Delete comment                │
│                                         │
│        [Create Rule]                    │
└─────────────────────────────────────────┘
```

---

## 5. Context Management Workflows

### 5.1 Context Document Upload
```
Step 1: Upload File
┌─────────────────────────────────────────┐
│ Add Context Document                    │
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │     Drag & drop files here          │ │
│ │            or                       │ │
│ │        [Browse Files]               │ │
│ │                                     │ │
│ │   Supported: PDF, Word, Text        │ │
│ │   Max size: 10MB per file           │ │
│ └─────────────────────────────────────┘ │
│                                         │
│ Document Name: ________________         │
│ Category: [FAQ] ▼                      │
│ Description: ___________________        │
│                                         │
│        [Upload]                         │
└─────────────────────────────────────────┘

Step 2: Processing Status
┌─────────────────────────────────────────┐
│ Processing Document...                  │
│                                         │
│ ✅ File uploaded                       │
│ ⏳ Extracting text content             │
│ ⏳ AI analysis in progress             │
│ ⏳ Indexing keywords                   │
│                                         │
│ [██████████████████████████████] 75%   │
│                                         │
│ This usually takes 30-60 seconds        │
└─────────────────────────────────────────┘

Step 3: Assignment Selection
┌─────────────────────────────────────────┐
│ Apply this context to:                  │
│                                         │
│ Social Accounts:                        │
│ ☐ @fashionbrand_official               │
│ ☐ @fashionbrand_sales                  │
│                                         │
│ Specific Campaigns:                     │
│ ☐ Summer Collection 2025                │
│ ☐ Back to School Promotion             │
│                                         │
│ ✅ Apply to all current and future     │
│    content (recommended)                │
│                                         │
│        [Save Assignment]                │
└─────────────────────────────────────────┘
```

### 5.2 Context Library Management
```
┌─────────────────────────────────────────────────────────────────┐
│ 📚 Context Library                                              │
├─────────────────────────────────────────────────────────────────┤
│ [+ Add Document] [📥 Import] [🏷️ Manage Categories]            │
│                                                                 │
│ Filter: [All] [FAQ] [Products] [Policies] | Sort: [Recent] ▼   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ ┌─ Document Card ──────────────────────────────────────────┐   │
│ │ 📄 Frequently Asked Questions v2.0                      │   │
│ │ 📁 FAQ • 💾 2.3MB • 📅 Updated 2 days ago              │   │
│ │                                                         │   │
│ │ Applied to: @fashionbrand_official, Summer Collection   │   │
│ │ Usage: 47 times this month | Effectiveness: 89%        │   │
│ │                                                         │   │
│ │ Recent topics matched: sizing, returns, shipping       │   │
│ │                                                         │   │
│ │ [✏️ Edit] [🎯 Assignments] [📊 Analytics] [🗑️ Delete] │   │
│ └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│ ┌─ Document Card ──────────────────────────────────────────┐   │
│ │ 📦 Product Catalog - Summer 2025                        │   │
│ │ 📁 Products • 💾 8.7MB • 📅 Updated 1 week ago          │   │
│ │                                                         │   │
│ │ Applied to: @fashionbrand_official only                 │   │
│ │ Usage: 23 times this month | Effectiveness: 92%        │   │
│ │                                                         │   │
│ │ [✏️ Edit] [🎯 Assignments] [📊 Analytics] [🗑️ Delete] │   │
│ └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Otto Chat Interface

### 6.1 Otto Chat Widget
```
┌─────────────────────────────────────────┐
│ 🤖 Chat with Otto                      │
├─────────────────────────────────────────┤
│                                         │
│ Otto: Hi! I'm here to help you analyze │
│ your community and automate responses.  │
│ What would you like to know?            │
│                                         │
│ Quick actions:                          │
│ 💭 Analyze sentiment for recent posts   │
│ 📊 Show top performing content          │
│ 🎯 Find potential leads                 │
│ ⚙️ Help me create a new rule           │
│                                         │
├─────────────────────────────────────────┤
│ Type your question...          [Send]   │
└─────────────────────────────────────────┘
```

### 6.2 Otto Conversation Examples
```
User: "What are people saying about my latest post?"

Otto: I analyzed 23 comments on your "Summer Dress Collection" post from 2 hours ago:

📊 Sentiment Breakdown:
• 😊 Positive: 78% (18 comments)
• 😐 Neutral: 17% (4 comments)  
• 😟 Negative: 4% (1 comment)

🎯 Key Themes:
• "love the colors" (mentioned 6 times)
• "where to buy" (mentioned 4 times)
• "price point" (mentioned 3 times)

🚨 Action Items:
• 4 comments asking about availability - should I auto-reply with the product link?
• 1 complaint about sizing - flagged for manual response

Would you like me to set up an auto-reply for purchase inquiries?
```

### 6.3 Otto Proactive Suggestions
```
Notification bubble appears:
┌─────────────────────────────────────────┐
│ 🤖 Otto noticed something...            │
│                                         │
│ Your "Blue Dress" post is getting 3x   │
│ more engagement than usual! 📈          │
│                                         │
│ People are asking about:                │
│ • Sizing (7 comments)                   │
│ • Price (5 comments)                    │
│ • Availability (4 comments)             │
│                                         │
│ Should I create auto-replies for these  │
│ common questions?                       │
│                                         │
│ [Yes, create templates] [Not now]       │
└─────────────────────────────────────────┘
```

---

## 7. Analytics & Reporting Interface

### 7.1 Analytics Dashboard
```
┌─────────────────────────────────────────────────────────────────┐
│ 📈 Analytics Dashboard                                          │
├─────────────────────────────────────────────────────────────────┤
│ Time Period: [Last 7 days] ▼ | Account: [All accounts] ▼      │
│                                                                 │
│ ┌─ Key Metrics ──────────────────────────────────────────────┐ │
│ │ 📊 Comments: 156 (+23%)   😊 Sentiment: 82% (+5%)        │ │
│ │ 🤖 Auto-replies: 34       🎯 Leads: 12 (+8 from last wk) │ │
│ └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│ ┌─ Sentiment Trend ─────────┐ ┌─ Top Topics ─────────────────┐ │
│ │        😊                 │ │ 1. Product inquiries (34%)   │ │
│ │       ╱ ╲                │ │ 2. Sizing questions (18%)    │ │
│ │      ╱   ╲               │ │ 3. Compliments (15%)         │ │
│ │ ────╱─────╲────          │ │ 4. Shipping info (12%)       │ │
│ │           ╲               │ │ 5. Color options (8%)        │ │
│ │            ╲              │ │                              │ │
│ │   Mon Tue Wed Thu Fri     │ │ [View detailed breakdown]    │ │
│ └───────────────────────────┘ └──────────────────────────────┘ │
│                                                                 │
│ ┌─ Automation Performance ──────────────────────────────────┐  │
│ │ Template: "Product Inquiry" - Used 23 times, 94% success  │  │
│ │ Template: "Sizing Help" - Used 12 times, 87% success      │  │
│ │ Rule: "Hide Competitor" - Triggered 3 times               │  │
│ └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## 8. Mobile Experience Considerations

### 8.1 Mobile Priority Features
```
Mobile App Features (Responsive Web MVP):
✅ Comment moderation (approve/reject)
✅ Urgent notifications
✅ Otto chat (simplified)
✅ Basic analytics view
❌ Complex rule creation
❌ Context document upload
❌ Detailed analytics
```

### 8.2 Mobile Comment Moderation
```
┌─────────────────────────────┐
│ 💬 3 pending comments       │
├─────────────────────────────┤
│ @sarah_fashion              │
│ "Love this dress! Do you... │
│ 😊 +0.85 | 🎯 Purchase      │
│                             │
│ 🤖 Suggest: Product reply   │
│ [✅] [✏️] [❌] [👁️]        │
├─────────────────────────────┤
│ @fashion_lover              │
│ "Where can I buy this? It's │
│ 😊 +0.72 | 🎯 Purchase      │
│                             │
│ [✅] [✏️] [❌] [👁️]        │
└─────────────────────────────┘
```

---

## 9. Error States & Edge Cases

### 9.1 Instagram Connection Issues
```
Connection Failed State:
┌─────────────────────────────────────────┐
│ ⚠️ Instagram Connection Lost            │
│                                         │
│ Your Instagram account was disconnected │
│ due to expired permissions.             │
│                                         │
│ What this means:                        │
│ • No new comments will be synced        │
│ • Automation is paused                  │
│ • Historical data is preserved          │
│                                         │
│ [Reconnect Instagram] [Learn More]      │
└─────────────────────────────────────────┘
```

### 9.2 Subscription Limit Warnings
```
Approaching Limit Warning:
┌─────────────────────────────────────────┐
│ 🚨 Comment Limit Warning                │
│                                         │
│ You've processed 950/1,000 comments     │
│ this month (95% of your limit).         │
│                                         │
│ What happens next:                      │
│ • New comments will queue for next month│
│ • Upgrade to continue processing        │
│                                         │
│ [Upgrade Plan] [View Usage] [Dismiss]   │
└─────────────────────────────────────────┘
```

---

This UX specification ensures a smooth, intuitive experience while maintaining the powerful functionality that makes Refyne valuable to creators and businesses.
