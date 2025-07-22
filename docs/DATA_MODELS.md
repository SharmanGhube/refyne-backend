# Refyne Platform - Data Models & Relationships

## Document Overview
This document defines the complete data structure and relationships for the Refyne platform, supporting the features outlined in the Product Specification.

---

## 1. Core Entity Relationships

```
┌─────────────────┐    1:N    ┌─────────────────┐    1:N    ┌─────────────────┐
│     Users       │◄──────────│   Workspaces    │◄──────────│ Social Accounts │
│                 │           │                 │           │                 │
└─────────────────┘           └─────────────────┘           └─────────────────┘
                                        │ 1:N                         │ 1:N
                                        ▼                             ▼
                              ┌─────────────────┐           ┌─────────────────┐
                              │    Context      │           │      Media      │
                              │   Documents     │           │    (Posts)      │
                              └─────────────────┘           └─────────────────┘
                                                                      │ 1:N
                                                                      ▼
                                                            ┌─────────────────┐
                                                            │    Comments     │
                                                            └─────────────────┘
```

---

## 2. Database Schema Definitions

### 2.1 Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    
    -- Subscription & Account Info
    subscription_tier VARCHAR(20) NOT NULL DEFAULT 'starter' 
        CHECK (subscription_tier IN ('starter', 'professional', 'business', 'enterprise')),
    subscription_status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (subscription_status IN ('active', 'cancelled', 'past_due', 'trialing')),
    subscription_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Onboarding Data
    user_type VARCHAR(50), -- Individual Creator, Small Business, etc.
    primary_platform VARCHAR(20),
    follower_range VARCHAR(20),
    primary_challenge TEXT,
    content_type VARCHAR(50),
    onboarding_completed BOOLEAN DEFAULT FALSE,
    
    -- Account Status
    status VARCHAR(20) NOT NULL DEFAULT 'Pending' 
        CHECK (status IN ('Pending', 'Active', 'Banned', 'Deleted')),
    is_active BOOLEAN NOT NULL DEFAULT false,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    
    -- Security
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_subscription_tier ON users(subscription_tier);
CREATE INDEX idx_users_subscription_status ON users(subscription_status);
```

### 2.2 Workspaces Table
```sql
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Settings
    timezone VARCHAR(50) DEFAULT 'UTC',
    default_moderation_strictness VARCHAR(10) DEFAULT 'medium'
        CHECK (default_moderation_strictness IN ('low', 'medium', 'high', 'custom')),
    
    -- Limits based on subscription
    max_social_accounts INTEGER DEFAULT 1,
    max_team_members INTEGER DEFAULT 1,
    max_context_documents INTEGER DEFAULT 3,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_workspaces_owner ON workspaces(owner_user_id);
CREATE INDEX idx_workspaces_created_at ON workspaces(created_at);
```

### 2.3 Workspace Members Table
```sql
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer'
        CHECK (role IN ('admin', 'moderator', 'viewer')),
    
    -- Invitation tracking
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    joined_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'declined', 'removed')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(workspace_id, user_id)
);

-- Indexes
CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);
CREATE INDEX idx_workspace_members_role ON workspace_members(role);
```

### 2.4 Social Accounts Table
```sql
CREATE TABLE social_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Platform Info
    platform VARCHAR(20) NOT NULL DEFAULT 'instagram'
        CHECK (platform IN ('instagram', 'youtube', 'tiktok', 'twitter')),
    platform_user_id VARCHAR(255) NOT NULL, -- Instagram user ID
    username VARCHAR(100) NOT NULL,
    display_name VARCHAR(255),
    profile_picture_url TEXT,
    
    -- Account Metrics
    follower_count INTEGER DEFAULT 0,
    following_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,
    
    -- OAuth & API Info
    access_token TEXT, -- Encrypted
    refresh_token TEXT, -- Encrypted
    token_expires_at TIMESTAMP WITH TIME ZONE,
    token_last_refreshed TIMESTAMP WITH TIME ZONE,
    
    -- Sync Status
    sync_status VARCHAR(20) DEFAULT 'active'
        CHECK (sync_status IN ('active', 'error', 'paused', 'disconnected')),
    last_sync_at TIMESTAMP WITH TIME ZONE,
    sync_error_message TEXT,
    
    -- Settings
    auto_sync_enabled BOOLEAN DEFAULT TRUE,
    moderation_enabled BOOLEAN DEFAULT TRUE,
    auto_reply_enabled BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(platform, platform_user_id)
);

-- Indexes
CREATE INDEX idx_social_accounts_workspace ON social_accounts(workspace_id);
CREATE INDEX idx_social_accounts_platform ON social_accounts(platform);
CREATE INDEX idx_social_accounts_sync_status ON social_accounts(sync_status);
```

### 2.5 Context Documents Table
```sql
CREATE TABLE context_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Document Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50), -- FAQ, Products, Policies, etc.
    file_name VARCHAR(255),
    file_size INTEGER, -- in bytes
    file_type VARCHAR(10), -- pdf, docx, txt
    
    -- Content
    original_content TEXT, -- Extracted text content
    processed_content JSONB, -- AI-processed structure
    keywords TEXT[], -- Extracted keywords for quick matching
    
    -- Processing Status
    processing_status VARCHAR(20) DEFAULT 'pending'
        CHECK (processing_status IN ('pending', 'processing', 'completed', 'failed')),
    processing_error TEXT,
    
    -- Usage Stats
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_context_docs_workspace ON context_documents(workspace_id);
CREATE INDEX idx_context_docs_category ON context_documents(category);
CREATE INDEX idx_context_docs_keywords ON context_documents USING GIN(keywords);
CREATE INDEX idx_context_docs_processing_status ON context_documents(processing_status);
```

### 2.6 Context Assignments Table
```sql
CREATE TABLE context_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_document_id UUID NOT NULL REFERENCES context_documents(id) ON DELETE CASCADE,
    
    -- Assignment Target (either account or media, not both)
    social_account_id UUID REFERENCES social_accounts(id) ON DELETE CASCADE,
    media_id UUID REFERENCES media(id) ON DELETE CASCADE,
    
    -- Assignment Config
    priority INTEGER DEFAULT 1, -- Higher number = higher priority
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure only one target is specified
    CONSTRAINT check_single_target CHECK (
        (social_account_id IS NOT NULL AND media_id IS NULL) OR
        (social_account_id IS NULL AND media_id IS NOT NULL)
    )
);

-- Indexes
CREATE INDEX idx_context_assignments_document ON context_assignments(context_document_id);
CREATE INDEX idx_context_assignments_social_account ON context_assignments(social_account_id);
CREATE INDEX idx_context_assignments_media ON context_assignments(media_id);
```

### 2.7 Media Table (Instagram Posts)
```sql
CREATE TABLE media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    social_account_id UUID NOT NULL REFERENCES social_accounts(id) ON DELETE CASCADE,
    
    -- Platform Data
    platform_media_id VARCHAR(255) NOT NULL, -- Instagram media ID
    media_type VARCHAR(20) NOT NULL 
        CHECK (media_type IN ('image', 'video', 'carousel', 'reel', 'story')),
    
    -- Content
    caption TEXT,
    media_url TEXT,
    thumbnail_url TEXT,
    permalink TEXT,
    
    -- Metrics (synced from platform)
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    share_count INTEGER DEFAULT 0,
    
    -- AI Analysis
    sentiment_score DECIMAL(3,2), -- -1.0 to 1.0
    topic_tags TEXT[], -- AI-extracted topics
    
    -- Organization
    campaign_id UUID REFERENCES campaigns(id),
    is_archived BOOLEAN DEFAULT FALSE,
    
    -- Platform Timestamps
    published_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- System Timestamps
    first_synced_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_synced_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(social_account_id, platform_media_id)
);

-- Indexes
CREATE INDEX idx_media_social_account ON media(social_account_id);
CREATE INDEX idx_media_platform_id ON media(platform_media_id);
CREATE INDEX idx_media_published_at ON media(published_at);
CREATE INDEX idx_media_campaign ON media(campaign_id);
CREATE INDEX idx_media_sentiment ON media(sentiment_score);
```

### 2.8 Campaigns Table
```sql
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Campaign Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    campaign_type VARCHAR(50), -- Product Launch, Seasonal, etc.
    
    -- Date Range
    start_date DATE,
    end_date DATE,
    
    -- Settings
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_campaigns_workspace ON campaigns(workspace_id);
CREATE INDEX idx_campaigns_dates ON campaigns(start_date, end_date);
```

### 2.9 Comments Table
```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_id UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    
    -- Platform Data
    platform_comment_id VARCHAR(255) NOT NULL,
    parent_comment_id UUID REFERENCES comments(id), -- For replies
    
    -- Comment Content
    text_content TEXT NOT NULL,
    commenter_username VARCHAR(100) NOT NULL,
    commenter_display_name VARCHAR(255),
    commenter_profile_picture TEXT,
    commenter_follower_count INTEGER,
    commenter_is_verified BOOLEAN DEFAULT FALSE,
    
    -- AI Analysis
    sentiment_score DECIMAL(3,2), -- -1.0 to 1.0
    toxicity_score DECIMAL(3,2), -- 0.0 to 1.0
    intent_classification VARCHAR(50), -- question, complaint, praise, etc.
    extracted_topics TEXT[], -- AI-identified topics
    language_code VARCHAR(5), -- en, es, fr, etc.
    
    -- Moderation
    moderation_status VARCHAR(20) DEFAULT 'pending'
        CHECK (moderation_status IN ('pending', 'approved', 'hidden', 'deleted', 'flagged')),
    moderation_action_reason TEXT,
    moderated_by UUID REFERENCES users(id),
    moderated_at TIMESTAMP WITH TIME ZONE,
    auto_moderated BOOLEAN DEFAULT FALSE,
    
    -- Engagement Tracking
    like_count INTEGER DEFAULT 0,
    reply_count INTEGER DEFAULT 0,
    has_response BOOLEAN DEFAULT FALSE,
    response_type VARCHAR(20), -- auto, manual, none
    
    -- Lead Scoring
    lead_score INTEGER DEFAULT 0, -- 0-100
    is_potential_lead BOOLEAN DEFAULT FALSE,
    lead_intent_keywords TEXT[],
    
    -- Timestamps
    posted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    first_synced_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(media_id, platform_comment_id)
);

-- Indexes
CREATE INDEX idx_comments_media ON comments(media_id);
CREATE INDEX idx_comments_platform_id ON comments(platform_comment_id);
CREATE INDEX idx_comments_posted_at ON comments(posted_at);
CREATE INDEX idx_comments_moderation_status ON comments(moderation_status);
CREATE INDEX idx_comments_sentiment ON comments(sentiment_score);
CREATE INDEX idx_comments_lead_score ON comments(lead_score);
CREATE INDEX idx_comments_commenter ON comments(commenter_username);
```

### 2.10 Moderation Rules Table
```sql
CREATE TABLE moderation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Rule Target (workspace, account, or media level)
    social_account_id UUID REFERENCES social_accounts(id) ON DELETE CASCADE,
    media_id UUID REFERENCES media(id) ON DELETE CASCADE,
    
    -- Rule Definition
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rule_type VARCHAR(50) NOT NULL 
        CHECK (rule_type IN ('keyword_filter', 'toxicity_threshold', 'spam_detection', 'competitor_mention', 'custom')),
    
    -- Rule Configuration (JSON for flexibility)
    rule_config JSONB NOT NULL,
    /* Example configurations:
    {
        "keywords": ["competitor1", "competitor2"],
        "action": "hide",
        "case_sensitive": false
    }
    
    {
        "toxicity_threshold": 0.8,
        "action": "delete",
        "notify_admin": true
    }
    */
    
    -- Rule Behavior
    action VARCHAR(20) NOT NULL DEFAULT 'flag'
        CHECK (action IN ('flag', 'hide', 'delete', 'approve')),
    strictness VARCHAR(10) DEFAULT 'medium'
        CHECK (strictness IN ('low', 'medium', 'high')),
    require_approval BOOLEAN DEFAULT TRUE,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 1, -- Higher number = higher priority
    
    -- Usage Stats
    triggered_count INTEGER DEFAULT 0,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_moderation_rules_workspace ON moderation_rules(workspace_id);
CREATE INDEX idx_moderation_rules_social_account ON moderation_rules(social_account_id);
CREATE INDEX idx_moderation_rules_type ON moderation_rules(rule_type);
CREATE INDEX idx_moderation_rules_active ON moderation_rules(is_active);
```

### 2.11 Response Templates Table
```sql
CREATE TABLE response_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Template Info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    template_type VARCHAR(50) DEFAULT 'comment_reply'
        CHECK (template_type IN ('comment_reply', 'dm_message', 'email')),
    
    -- Trigger Configuration
    trigger_keywords TEXT[], -- Keywords that activate this template
    trigger_sentiment VARCHAR(20), -- positive, negative, neutral, any
    trigger_intent VARCHAR(50), -- question, complaint, purchase_intent, etc.
    
    -- Template Content
    content TEXT NOT NULL, -- Template with variables like {username}, {product_name}
    variables JSONB, -- Available variables and their descriptions
    
    -- Behavior
    requires_approval BOOLEAN DEFAULT TRUE,
    auto_send BOOLEAN DEFAULT FALSE,
    follow_up_action VARCHAR(50), -- send_dm, add_to_crm, none
    
    -- Usage Stats
    usage_count INTEGER DEFAULT 0,
    success_rate DECIMAL(3,2) DEFAULT 0.0, -- Based on user feedback
    last_used_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_response_templates_workspace ON response_templates(workspace_id);
CREATE INDEX idx_response_templates_keywords ON response_templates USING GIN(trigger_keywords);
CREATE INDEX idx_response_templates_type ON response_templates(template_type);
CREATE INDEX idx_response_templates_active ON response_templates(is_active);
```

### 2.12 Automated Responses Table
```sql
CREATE TABLE automated_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    template_id UUID REFERENCES response_templates(id),
    
    -- Response Details
    response_type VARCHAR(20) NOT NULL DEFAULT 'comment_reply'
        CHECK (response_type IN ('comment_reply', 'dm_message', 'email')),
    content TEXT NOT NULL,
    
    -- Processing Status
    status VARCHAR(20) DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'sent', 'failed')),
    
    -- Platform Response
    platform_response_id VARCHAR(255), -- ID returned by platform API
    platform_error TEXT,
    
    -- Approval Workflow
    requires_approval BOOLEAN DEFAULT TRUE,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    
    -- Timestamps
    scheduled_for TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_automated_responses_comment ON automated_responses(comment_id);
CREATE INDEX idx_automated_responses_template ON automated_responses(template_id);
CREATE INDEX idx_automated_responses_status ON automated_responses(status);
CREATE INDEX idx_automated_responses_scheduled ON automated_responses(scheduled_for);
```

---

## 3. Data Relationships & Business Rules

### 3.1 Subscription Tier Limits

```sql
-- Function to check workspace limits
CREATE OR REPLACE FUNCTION check_workspace_limits(user_uuid UUID)
RETURNS BOOLEAN AS $$
DECLARE
    user_tier VARCHAR(20);
    current_workspaces INTEGER;
    max_allowed INTEGER;
BEGIN
    -- Get user's subscription tier
    SELECT subscription_tier INTO user_tier FROM users WHERE id = user_uuid;
    
    -- Count current workspaces
    SELECT COUNT(*) INTO current_workspaces 
    FROM workspaces 
    WHERE owner_user_id = user_uuid AND deleted_at IS NULL;
    
    -- Set limits based on tier
    max_allowed := CASE user_tier
        WHEN 'starter' THEN 1
        WHEN 'professional' THEN 1
        WHEN 'business' THEN 3
        WHEN 'enterprise' THEN 999999
        ELSE 1
    END;
    
    RETURN current_workspaces < max_allowed;
END;
$$ LANGUAGE plpgsql;
```

### 3.2 Context Document Limits

```sql
-- Function to check context document limits
CREATE OR REPLACE FUNCTION check_context_limits(workspace_uuid UUID)
RETURNS BOOLEAN AS $$
DECLARE
    owner_tier VARCHAR(20);
    current_docs INTEGER;
    max_allowed INTEGER;
BEGIN
    -- Get workspace owner's subscription tier
    SELECT u.subscription_tier INTO owner_tier
    FROM users u
    JOIN workspaces w ON u.id = w.owner_user_id
    WHERE w.id = workspace_uuid;
    
    -- Count current context documents
    SELECT COUNT(*) INTO current_docs
    FROM context_documents
    WHERE workspace_id = workspace_uuid AND deleted_at IS NULL;
    
    -- Set limits based on tier
    max_allowed := CASE owner_tier
        WHEN 'starter' THEN 3
        WHEN 'professional' THEN 999999
        WHEN 'business' THEN 999999
        WHEN 'enterprise' THEN 999999
        ELSE 3
    END;
    
    RETURN current_docs < max_allowed;
END;
$$ LANGUAGE plpgsql;
```

### 3.3 Data Retention Policies

```sql
-- Archive old comments (Enterprise tier: unlimited, others: 1 year)
CREATE OR REPLACE FUNCTION archive_old_comments()
RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER := 0;
    cutoff_date TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Archive comments older than 1 year for non-enterprise accounts
    cutoff_date := CURRENT_TIMESTAMP - INTERVAL '1 year';
    
    UPDATE comments 
    SET is_archived = TRUE
    WHERE posted_at < cutoff_date
    AND id IN (
        SELECT c.id FROM comments c
        JOIN media m ON c.media_id = m.id
        JOIN social_accounts sa ON m.social_account_id = sa.id
        JOIN workspaces w ON sa.workspace_id = w.id
        JOIN users u ON w.owner_user_id = u.id
        WHERE u.subscription_tier != 'enterprise'
    );
    
    GET DIAGNOSTICS archived_count = ROW_COUNT;
    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;
```

---

## 4. API Response Models

### 4.1 User Profile Response
```json
{
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "subscription": {
        "tier": "professional",
        "status": "active",
        "expires_at": "2025-08-22T00:00:00Z"
    },
    "onboarding": {
        "completed": true,
        "user_type": "Individual Creator",
        "primary_platform": "Instagram",
        "follower_range": "10K - 100K"
    },
    "created_at": "2025-01-01T00:00:00Z"
}
```

### 4.2 Workspace Response
```json
{
    "id": "uuid",
    "name": "Fashion Brand Workspace",
    "description": "Managing our fashion brand's social presence",
    "owner": {
        "id": "uuid",
        "username": "owner_username"
    },
    "settings": {
        "timezone": "America/New_York",
        "default_moderation_strictness": "medium"
    },
    "limits": {
        "max_social_accounts": 3,
        "max_team_members": 5,
        "max_context_documents": 999999
    },
    "current_usage": {
        "social_accounts": 2,
        "team_members": 3,
        "context_documents": 15
    },
    "team_members": [
        {
            "user_id": "uuid",
            "username": "team_member",
            "role": "moderator",
            "joined_at": "2025-01-15T00:00:00Z"
        }
    ],
    "created_at": "2025-01-01T00:00:00Z"
}
```

### 4.3 Comment with Analysis Response
```json
{
    "id": "uuid",
    "media_id": "uuid",
    "platform_comment_id": "instagram_comment_id",
    "content": {
        "text": "Love this outfit! Where can I buy it?",
        "commenter": {
            "username": "fashion_lover123",
            "display_name": "Sarah Johnson",
            "is_verified": false,
            "follower_count": 1250
        }
    },
    "analysis": {
        "sentiment_score": 0.85,
        "toxicity_score": 0.05,
        "intent_classification": "purchase_inquiry",
        "topics": ["fashion", "product_inquiry"],
        "language": "en",
        "lead_score": 78
    },
    "moderation": {
        "status": "approved",
        "auto_moderated": false,
        "rules_triggered": []
    },
    "response": {
        "has_response": true,
        "response_type": "auto",
        "template_used": "Product Inquiry Template",
        "status": "sent"
    },
    "timestamps": {
        "posted_at": "2025-07-22T10:30:00Z",
        "first_synced_at": "2025-07-22T10:31:00Z"
    }
}
```

---

This data model supports all the features outlined in the Product Specification while maintaining flexibility for future enhancements. The schema includes proper indexing for performance and constraints to ensure data integrity.
