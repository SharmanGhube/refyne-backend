# Refyne Platform - Complete Product Specification

## Document Overview

**Project:** Refyne - AI-Powered Community Growth Platform  
**Version:** 1.0.0 (MVP Focus)  
**Last Updated:** July 22, 2025  
**Document Type:** Product Requirements Document (PRD)

## Related Documents
- [User Experience & Workflows](./UX_WORKFLOWS.md)
- [Data Models & Relationships](./DATA_MODELS.md)
- [AI & Automation Features](./AI_AUTOMATION.md)
- [Technical Specification](./TECHNICAL_SPECIFICATION.md)

---

## 1. Product Vision & Scope

### 1.1 Vision Statement
Refyne is the essential AI-powered community growth platform that helps creators and businesses turn their social media communities into their most valuable asset through intelligent protection, understanding, engagement, and growth tools powered by Otto, the AI assistant.

### 1.2 MVP Scope
**Primary Focus:** Instagram comment management and basic automation  
**Core Features:**
- Single workspace management
- Instagram account connection
- Basic comment moderation with Otto
- Simple automated responses
- Context-aware AI interactions
- Basic analytics and insights

**Out of Scope for MVP:**
- Multiple workspace management
- Team collaboration features
- Advanced workflow builder
- CRM integrations
- Lead generation pipelines
- Enterprise features

---

## 2. User Account & Authentication

### 2.1 Account Types
**Single Account Type:** All users have the same base account type with feature access determined by subscription tier.

### 2.2 Registration Process
**Required Information:**
- Email address (primary identifier)
- Password (minimum 8 characters, special chars required)

**Post-Registration Onboarding Questionnaire:**
1. **"What best describes you?"**
   - Individual Creator/Influencer
   - Small Business Owner
   - Marketing Agency
   - Enterprise/Large Company
   - Other

2. **"What's your primary social media platform?"**
   - Instagram
   - YouTube
   - TikTok
   - Multiple platforms
   - Other

3. **"How many followers do you have across all platforms?"**
   - Under 1K
   - 1K - 10K
   - 10K - 100K
   - 100K - 1M
   - 1M+

4. **"What's your biggest challenge with community management?"**
   - Too many comments to respond to
   - Dealing with spam/toxic comments
   - Understanding audience sentiment
   - Converting engagement to sales
   - All of the above

5. **"What type of content do you primarily create?"**
   - Lifestyle/Personal
   - Business/Professional
   - Entertainment
   - Education
   - E-commerce/Products
   - Other

### 2.3 Subscription Tiers

#### **Starter Tier - $29/month**
- 1 workspace
- 1 Instagram account
- Up to 1,000 comments processed/month
- Basic moderation (toxicity, spam)
- 3 context documents (max 10 pages each)
- Basic analytics dashboard
- Email support

#### **Professional Tier - $79/month**
- 1 workspace
- 3 Instagram accounts
- Up to 10,000 comments processed/month
- Advanced moderation (custom rules)
- Unlimited context documents
- Auto-reply to comments (10 templates)
- Advanced analytics & weekly reports
- Chat support
- Sentiment analysis

#### **Business Tier - $199/month**
- 3 workspaces
- 10 Instagram accounts
- Up to 50,000 comments processed/month
- Everything in Professional
- Team collaboration (up to 5 members)
- Custom moderation rules per account
- DM automation (basic)
- Lead identification
- Priority support

#### **Enterprise Tier - Custom Pricing**
- Unlimited workspaces
- Unlimited Instagram accounts
- Unlimited comment processing
- Everything in Business
- Advanced workflow automation
- CRM integrations
- White-label options
- Custom AI model training
- Dedicated account manager

---

## 3. Workspace Management

### 3.1 Workspace Structure
```
Refyne Account
└── Workspaces (limit based on tier)
    ├── Social Accounts (Instagram)
    ├── Context Library
    ├── Team Members & Permissions
    ├── Moderation Rules
    └── Analytics & Reports
```

### 3.2 Workspace Limits by Tier
- **Starter:** 1 workspace
- **Professional:** 1 workspace
- **Business:** 3 workspaces
- **Enterprise:** Unlimited

### 3.3 Team Collaboration (Business+ Tiers)

#### **Role Permissions:**

**Admin (Workspace Owner)**
- Full access to all features
- Invite/remove team members
- Manage billing and subscription
- Delete workspace
- Manage all social accounts

**Moderator**
- Manage moderation rules
- Approve/reject AI responses
- View all analytics
- Manage context library
- Cannot invite users or manage billing

**Viewer**
- View analytics only
- Cannot modify settings
- Cannot approve responses
- Read-only access to context

### 3.4 Workspace Sharing
- Users can be invited to multiple workspaces
- Each workspace invitation is role-specific
- Users maintain their own Refyne account across workspaces

---

## 4. Social Media Integration

### 4.1 Instagram Integration (MVP)

#### **Connection Process:**
1. User clicks "Connect Instagram Account"
2. OAuth flow redirects to Instagram
3. User authorizes Refyne app
4. System stores access token and basic profile info
5. Auto-sync begins for future content

#### **Data Synchronization:**
**What We Sync:**
- Posts (photos, videos, carousels, reels)
- Comments on posts (new only, no historical)
- Post metrics (likes, views, shares)
- Follower count and basic audience data

**Sync Frequency:**
- Real-time for new comments (webhook)
- Hourly for post metrics
- Daily for follower data

#### **Account Limitations:**
- One Instagram account per workspace (Starter/Professional)
- One Instagram account can only belong to one workspace
- No cross-workspace account sharing

### 4.2 Content Organization

#### **Media Management:**
- All posts auto-imported after connection
- Users can organize posts into campaigns
- Campaign examples:
  - "Summer Collection Launch"
  - "Black Friday Promotion"
  - "Product Reviews"

#### **Campaign Features:**
- Group related posts
- Apply campaign-specific moderation rules
- Campaign-level analytics
- Bulk context assignment

---

## 5. Context & Knowledge Management

### 5.1 Context Types & Formats

#### **Supported Content Types:**
- **FAQs:** Common questions and answers
- **Product Information:** Catalogs, specifications, pricing
- **Brand Guidelines:** Tone, voice, messaging do's and don'ts
- **Competitor Information:** What not to promote or mention
- **Company Policies:** Return policies, shipping info, etc.
- **Seasonal Information:** Holiday hours, special promotions

#### **Supported File Formats:**
- PDF (up to 50 pages)
- Word Documents (.docx)
- Plain Text (.txt)
- Maximum file size: 10MB per document

### 5.2 Context Processing
- Files uploaded are processed by AI to extract key information
- Text is indexed for quick retrieval during AI operations
- Context is automatically included in relevant AI prompts

### 5.3 Context Assignment Strategy

#### **Workspace-Level Context Library:**
- All context documents uploaded to workspace
- Shared across all social accounts in workspace
- Organized by categories (FAQ, Products, Policies, etc.)

#### **Account-Level Context Selection:**
- For each Instagram account, users select which context applies
- Example: Fashion brand account only uses fashion FAQ and product catalog
- Business account uses all contexts including policies

#### **Media-Level Context Override:**
- For specific posts/campaigns, users can override context
- Example: Product launch post only uses new product context
- Temporary context for limited-time promotions

### 5.4 Context Management UI
```
Context Library
├── General FAQ (Applied to: @fashion_brand, @business_account)
├── Product Catalog - Summer 2025 (Applied to: @fashion_brand)
├── Shipping Policies (Applied to: All accounts)
├── Competitor Guidelines (Applied to: @fashion_brand)
└── Holiday Schedule (Applied to: @business_account)
```

---

## 6. Otto AI Core Functionality

### 6.1 AI Foundation
- **Primary Engine:** Google Gemini API
- **Context Integration:** Relevant context automatically included in all AI prompts
- **Learning:** No model training, relies on context and prompt engineering

### 6.2 The Four Pillars Implementation

#### **Pillar 1: Protect (Moderation)**

**Default Moderation Categories:**
- Spam and promotional content
- Hate speech and toxicity
- Sexual or inappropriate content
- Violence and harmful content
- Competitor mentions

**Custom Rules Engine:**
```
Rule Examples:
- "Delete any comments mentioning [competitor names]"
- "Flag comments asking about suicide or self-harm for manual review"
- "Hide comments with excessive emojis (>5)"
- "Auto-approve comments from verified accounts"
```

**Action Options by Strictness:**
- **Low:** Flag for review only
- **Medium:** Auto-hide and flag for review
- **High:** Auto-delete and log action
- **Custom:** User-defined actions per rule

**Whitelist Features:**
- VIP users (never moderate their comments)
- Approved keywords (never flag these terms)
- Team members and collaborators

#### **Pillar 2: Understand (Analytics)**

**Sentiment Analysis:**
- Overall sentiment score per post
- Sentiment trends over time
- Sentiment by comment topic
- Sentiment alerts for negative spikes

**Topic & Trend Detection:**
- Most discussed topics in comments
- Trending keywords and phrases
- Competitor mention tracking
- Product/service feedback themes

**Engagement Insights:**
- Response rate to different content types
- Best performing posting times
- Audience behavior patterns
- Comment quality metrics

**Reporting Schedule:**
- Real-time dashboard updates
- Daily sentiment summaries
- Weekly comprehensive reports (email)
- Urgent alerts for significant changes

#### **Pillar 3: Engage (Automation)**

**Auto-Reply System:**
- Keyword-triggered responses
- Sentiment-based responses
- FAQ matching and auto-answers
- Custom templates with variables

**Reply Rules Engine:**
```
Rule Examples:
- IF comment contains "price" OR "cost" → Reply with pricing template
- IF comment sentiment is negative → Flag for manual response
- IF comment asks question in FAQ → Auto-reply with FAQ answer
- IF comment from VIP user → Always reply with personalized template
```

**Human-in-the-Loop:**
- All automated responses queued for approval (default)
- Auto-approve for trusted templates (optional)
- Manual override for all responses
- Bulk approval interface

**Response Templates:**
```json
{
  "name": "Pricing Inquiry",
  "trigger_keywords": ["price", "cost", "how much"],
  "response": "Hi {username}! Thanks for your interest. I'll send you our pricing details in a DM right now! 💌",
  "follow_up_action": "send_dm"
}
```

#### **Pillar 4: Grow (Lead Generation) - Limited in MVP**

**Basic Lead Identification:**
- Comments indicating purchase intent
- Questions about specific products/services
- Requests for more information

**Simple Lead Actions:**
- Auto-reply with CTA
- Send follow-up DM (manually triggered)
- Flag as potential lead in dashboard

---

## 7. User Experience & Interface

### 7.1 Main Dashboard

#### **Dashboard Overview:**
```
Recent Activity Feed
├── New Comments (last 24 hours)
├── Moderation Actions Taken
├── Automated Responses Sent
└── Lead Opportunities Identified

Quick Stats
├── Comment Volume (today vs yesterday)
├── Sentiment Score (current)
├── Response Rate (last 7 days)
└── Active Social Accounts

Pending Actions
├── Comments Awaiting Approval (count)
├── Flagged Content to Review (count)
└── Response Templates to Approve (count)
```

### 7.2 Navigation Structure
```
Main Navigation
├── Dashboard (overview)
├── Comments (moderation interface)
├── Automation (rules and templates)
├── Analytics (insights and reports)
├── Context (knowledge management)
├── Settings (account and workspace)
└── Otto Chat (AI assistant interface)
```

### 7.3 Mobile Experience
**MVP:** Responsive web only, no native app
**Key Mobile Features:**
- Comment moderation on-the-go
- Urgent notification alerts
- Quick approval/rejection of responses
- Basic analytics viewing

---

## 8. Workflow & Automation System

### 8.1 MVP Automation Features

#### **Simple Triggers:**
- New comment posted
- Specific keyword mentioned
- Sentiment threshold reached
- Manual user trigger

#### **Basic Actions:**
- Reply to comment (template-based)
- Send DM (manual trigger only in MVP)
- Flag for review
- Add note to comment

#### **Conditional Logic (Simple):**
```
IF comment contains "shipping" 
AND sentiment is neutral/positive 
THEN reply with shipping FAQ template

IF comment contains competitor name 
THEN hide comment and notify admin
```

### 8.2 Testing & Simulation
- **Preview Mode:** Show what automation would do without taking action
- **Test Comments:** System generates sample comments to test rules
- **Rollback:** Ability to undo automated actions within 24 hours

---

## 9. Better Alternatives & Recommendations

### 9.1 Subscription Tier Improvements
**Recommendation:** Add usage-based pricing for enterprise clients
- Base enterprise fee + per-comment processing fee
- Allows for true scalability and fair pricing

### 9.2 Context Management Enhancement
**Recommendation:** Smart Context Suggestions
- AI analyzes comments and suggests which context documents are most relevant
- Auto-tagging of context documents for easier organization

### 9.3 Lead Generation Enhancement
**Recommendation:** Intent Scoring System
- Score comments 1-10 based on purchase intent
- Priority queue for high-intent comments
- Automated nurture sequences for different intent levels

### 9.4 Team Collaboration Improvement
**Recommendation:** Activity Feed & Notifications
- Real-time activity feed showing team actions
- Role-based notifications (moderators only see moderation alerts)
- Comment assignment system for team accountability

---

## 10. MVP Feature Priority

### 10.1 Phase 1 (Core MVP)
- [ ] User registration and authentication
- [ ] Single workspace creation
- [ ] Instagram OAuth connection
- [ ] Basic comment sync and display
- [ ] Simple moderation rules (toxic/spam detection)
- [ ] Context document upload and basic processing
- [ ] Otto chat interface for basic queries

### 10.2 Phase 2 (Enhanced MVP)
- [ ] Auto-reply templates and keyword triggers
- [ ] Sentiment analysis and basic analytics
- [ ] Email notifications and weekly reports
- [ ] Manual response approval system
- [ ] Campaign organization for posts

### 10.3 Phase 3 (Pre-Scale)
- [ ] Multiple social accounts per workspace
- [ ] Advanced moderation rules
- [ ] Team collaboration (Business tier)
- [ ] Basic lead identification
- [ ] Mobile-optimized interface

---

**Next Steps:** Create detailed UX workflows and data models to support this specification.
