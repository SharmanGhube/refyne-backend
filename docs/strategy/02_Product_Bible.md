# 02: Refyne Product Bible

This document is the master list of all features for the Refyne platform, organized by the four pillars of community growth.

## Pillar 1: Protect - Create a Safe & Positive Space

1.  **AI Toxic Comment Detection:** Uses Gemini API with custom prompt engineering for high-accuracy, context-aware toxicity scoring.
2.  **Spam & Bot Detection:** Employs pattern recognition (regex for emoji/link spam) and ML models to detect bot-like behavior.
3.  **Contextual Filtering:** Goes beyond keywords with advanced NLP, slang dictionaries, and brand-specific context to reduce false positives.
4.  **Custom Ruleset Filters:** A UI rule builder allows users to define keywords, regex patterns, and sentiment thresholds for moderation.
5.  **Manual Review Mode:** A task queue for low-confidence decisions, allowing for a human-in-the-loop review process.
6.  **Soft Mute/Shadowban:** Hides comments from public view while keeping them visible to the original commenter via platform API calls.
7.  **Repeat Offender Tagging:** A user behavior tracking system with a strike system to identify and penalize repeat offenders.

## Pillar 2: Understand - Gain Actionable Audience Insights

8.  **Sentiment Analysis:** LLM-based sentiment scoring (positive, negative, neutral, and nuanced emotions) per comment, post, and campaign.
9.  **Brand Safety Score:** A weighted algorithm that calculates an overall community health score based on toxicity, engagement quality, and sentiment.
10. **Comment Volume Tracker:** Time-series data collection and visualization for engagement patterns.
11. **Top Keywords/Emojis:** NLP extraction and frequency analysis to identify trending topics and reactions.
12. **Feedback Extractor:** A classification model that distinguishes between constructive criticism and hate speech.
13. **Time-Based Heatmaps:** Temporal analysis of engagement and toxicity patterns to find optimal posting times.
14. **Commenter Insights:** User clustering and engagement pattern analysis to identify top fans, potential trolls, and new followers.

## Pillar 3: Engage - Build Loyalty on Autopilot

15. **Auto-Replies to FAQs:** Intent recognition (e.g., "pricing," "where to buy") to provide templated, helpful responses.
16. **Auto-Pin Top Comments:** A quality-scoring algorithm identifies and automatically pins the most positive and relevant comments.
17. **Auto-React to Fans:** Positive sentiment detection triggers automated reactions (e.g., a heart emoji) to build goodwill.
18. **Smart DMs to Supporters:** Identifies high-value supporters and sends templated outreach messages.
19. **Response Queue for Teams:** A prioritized task system with suggested replies for team-based community management.

## Pillar 4: Grow - Turn Engagement into Results

20. **Automated Lead Capture:** A visual workflow builder that allows users to create automations to capture leads from comments and DMs.
21. **DM Sales Funnels:** Conversational AI (chatbot) that guides potential customers through a sales journey within their DMs.
22. **CRM & Email Integration:** Seamlessly sync captured leads and customer data to third-party tools like HubSpot, Salesforce, and Mailchimp.
23. **Automation Pipeline:** The core engine that allows users to connect triggers (e.g., new comment with keyword) to a sequence of actions (e.g., reply, send DM, add to CRM).

## Cross-Pillar & Platform Features

-   **Alerts & Reporting:** Weekly summary emails, sentiment spike alerts, toxic surge warnings, exportable reports, and LLM-generated campaign recaps.
-   **Workspace Management:** Campaign grouping, multiple workspaces, team roles & permissions (RBAC).
-   **Platform Infrastructure:** Authentication, billing dashboard, subscription management, usage tracking.
-   **Premium/Enterprise:** White-label dashboards, public API & webhooks, pre-publish risk scanning, custom AI model training.
