# Refyne Backend - Optimal Folder Structure

## Document Overview
**Project:** Refyne Platform Backend Architecture  
**Purpose:** Define maintainable, flexible, and scalable folder structure  
**Architecture:** Domain-Driven Design (DDD) with Clean Architecture principles  
**Language:** Go 1.24.4+  
**Last Updated:** July 22, 2025  

---

## 🏗️ Complete Folder Structure

```
refyne-backend/
├── .env                                    # Environment variables (dev)
├── .env.example                           # Environment template
├── .gitignore                             # Git ignore patterns
├── .air.toml                             # Air hot reloading config
├── .golangci.yml                         # Linting configuration
├── Dockerfile                            # Production container
├── Dockerfile.dev                       # Development container
├── docker-compose.yml                   # Multi-service dev setup
├── docker-compose.prod.yml              # Production compose
├── go.mod                                # Go module definition
├── go.sum                                # Dependency checksums
├── Makefile                              # Build automation
├── README.md                             # Project documentation
├── CONTRIBUTING.md                       # Development guidelines
│
├── api/                                  # API specifications
│   ├── openapi/                         # OpenAPI/Swagger specs
│   │   ├── v1/
│   │   │   ├── auth.yaml
│   │   │   ├── workspaces.yaml
│   │   │   ├── instagram.yaml
│   │   │   ├── moderation.yaml
│   │   │   ├── analytics.yaml
│   │   │   └── otto.yaml
│   │   └── refyne-api-v1.yaml          # Combined API spec
│   └── postman/                         # Postman collections
│       ├── Refyne-Dev.postman_collection.json
│       └── Refyne-Prod.postman_collection.json
│
├── bin/                                  # Compiled binaries
│   ├── app                              # Main application binary
│   ├── migrator                         # Database migration tool
│   └── seeder                          # Database seeding tool
│
├── build/                               # Build and packaging
│   ├── ci/                             # CI/CD scripts
│   │   ├── github-actions/
│   │   ├── docker/
│   │   └── k8s/                        # Kubernetes manifests
│   └── package/                        # Release packages
│
├── cmd/                                 # Application entry points
│   ├── api/                            # Main API server
│   │   ├── main.go                     # Primary application entry
│   │   ├── wire.go                     # Dependency injection
│   │   └── wire_gen.go                 # Generated DI code
│   ├── migrator/                       # Database migration tool
│   │   └── main.go
│   ├── seeder/                         # Database seeding tool
│   │   └── main.go
│   └── worker/                         # Background job worker
│       └── main.go
│
├── configs/                            # Configuration templates
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   ├── config.test.yaml
│   └── docker/
│       ├── postgres.conf
│       └── redis.conf
│
├── deployments/                        # Deployment configurations
│   ├── docker/
│   │   └── docker-compose.override.yml
│   ├── k8s/                           # Kubernetes manifests
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── secret.yaml
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── ingress.yaml
│   └── terraform/                      # Infrastructure as Code
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
│
├── docs/                               # Project documentation
│   ├── api/                           # API documentation
│   │   ├── authentication.md
│   │   ├── rate-limiting.md
│   │   └── error-handling.md
│   ├── architecture/                  # Architecture docs
│   │   ├── domain-design.md
│   │   ├── data-flow.md
│   │   └── security.md
│   ├── deployment/                    # Deployment guides
│   │   ├── local-development.md
│   │   ├── staging.md
│   │   └── production.md
│   ├── TECHNICAL_SPECIFICATION.md    # Technical spec
│   ├── PRODUCT_SPECIFICATION.md      # Product requirements
│   ├── DATA_MODELS.md                # Database schemas
│   ├── UX_WORKFLOWS.md               # User experience flows
│   ├── AI_AUTOMATION.md              # AI & automation features
│   ├── MVP_IMPLEMENTATION_GUIDE.md   # Implementation guide
│   ├── FEATURE_ENHANCEMENTS.md       # Future enhancements
│   ├── MVP_TODO_LIST.md              # Development TODO
│   ├── FOLDER_STRUCTURE.md           # This document
│   └── strategy/                      # Business strategy docs
│       ├── 01_Brand_Identity.md
│       ├── 02_Product_Bible.md
│       ├── 03_Technical_Plan.md
│       └── 04_Roadmap_and_GTM.md
│
├── internal/                          # Private application code
│   ├── api/                          # HTTP layer (Gin framework)
│   │   ├── router.go                 # Main router setup
│   │   ├── wire.go                   # API layer DI
│   │   ├── middlewares/              # HTTP middlewares
│   │   │   ├── auth.go              # JWT authentication
│   │   │   ├── cors.go              # CORS handling
│   │   │   ├── logging.go           # Request logging
│   │   │   ├── rate_limit.go        # Rate limiting
│   │   │   ├── request_id.go        # Request ID generation
│   │   │   ├── recovery.go          # Panic recovery
│   │   │   └── validation.go        # Input validation
│   │   └── responses/               # Standardized API responses
│   │       ├── success.go
│   │       ├── error.go
│   │       └── pagination.go
│   │
│   ├── bootstrap/                    # Application bootstrapping
│   │   ├── app.go                   # Main application struct
│   │   ├── wire.go                  # Bootstrap DI
│   │   ├── server.go                # HTTP server management
│   │   └── graceful_shutdown.go     # Graceful shutdown logic
│   │
│   ├── config/                       # Configuration management
│   │   ├── config.go                # Configuration struct
│   │   ├── wire.go                  # Config DI
│   │   ├── validation.go            # Config validation
│   │   └── loader.go                # Config loading logic
│   │
│   ├── database/                     # Database layer
│   │   ├── connection.go            # Database connections
│   │   ├── init.go                  # Database initialization
│   │   ├── pool.go                  # Connection pooling
│   │   ├── wire.go                  # Database DI
│   │   ├── health.go                # Database health checks
│   │   ├── transaction.go           # Transaction management
│   │   ├── migrations/              # Database migrations
│   │   │   ├── migrations.go        # Migration runner
│   │   │   ├── 000001_create_users_table.up.sql
│   │   │   ├── 000001_create_users_table.down.sql
│   │   │   ├── 000002_create_workspaces_table.up.sql
│   │   │   ├── 000002_create_workspaces_table.down.sql
│   │   │   ├── 000003_create_social_accounts_table.up.sql
│   │   │   ├── 000003_create_social_accounts_table.down.sql
│   │   │   ├── 000004_create_media_table.up.sql
│   │   │   ├── 000004_create_media_table.down.sql
│   │   │   ├── 000005_create_comments_table.up.sql
│   │   │   ├── 000005_create_comments_table.down.sql
│   │   │   ├── 000006_create_context_documents_table.up.sql
│   │   │   ├── 000006_create_context_documents_table.down.sql
│   │   │   ├── 000007_create_response_templates_table.up.sql
│   │   │   ├── 000007_create_response_templates_table.down.sql
│   │   │   ├── 000008_create_moderation_rules_table.up.sql
│   │   │   ├── 000008_create_moderation_rules_table.down.sql
│   │   │   └── 000009_create_analytics_tables.up.sql
│   │   └── seeds/                   # Database seeders
│   │       ├── users.sql
│   │       ├── workspaces.sql
│   │       └── response_templates.sql
│   │
│   ├── domain/                       # Business domain layer (DDD)
│   │   ├── auth/                    # Authentication domain
│   │   │   ├── wire.go              # Auth domain DI
│   │   │   ├── entities/            # Domain entities
│   │   │   │   ├── session.go
│   │   │   │   └── token.go
│   │   │   ├── value_objects/       # Value objects
│   │   │   │   ├── email.go
│   │   │   │   └── password.go
│   │   │   ├── repositories/        # Repository interfaces
│   │   │   │   ├── auth_repository.go
│   │   │   │   └── session_repository.go
│   │   │   ├── services/            # Domain services
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── password_service.go
│   │   │   │   ├── token_service.go
│   │   │   │   └── onboarding_service.go
│   │   │   ├── handlers/            # HTTP handlers
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── refresh_handler.go
│   │   │   │   └── logout_handler.go
│   │   │   ├── routes/              # Route definitions
│   │   │   │   └── auth_routes.go
│   │   │   ├── models/              # Data models
│   │   │   │   ├── login_request.go
│   │   │   │   ├── register_request.go
│   │   │   │   └── auth_response.go
│   │   │   ├── utils/               # Domain utilities
│   │   │   │   ├── jwt.go
│   │   │   │   ├── password.go
│   │   │   │   └── validation.go
│   │   │   └── errors/              # Domain-specific errors
│   │   │       ├── auth_errors.go
│   │   │       └── validation_errors.go
│   │   │
│   │   ├── user/                    # User management domain
│   │   │   ├── wire.go              # User domain DI
│   │   │   ├── entities/            # Domain entities
│   │   │   │   └── user.go
│   │   │   ├── value_objects/       # Value objects
│   │   │   │   ├── user_id.go
│   │   │   │   ├── username.go
│   │   │   │   └── user_status.go
│   │   │   ├── repositories/        # Repository interfaces
│   │   │   │   ├── user_repository.go
│   │   │   │   └── user_settings_repository.go
│   │   │   ├── services/            # Domain services
│   │   │   │   ├── user_service.go
│   │   │   │   └── profile_service.go
│   │   │   ├── handlers/            # HTTP handlers
│   │   │   │   ├── user_handler.go
│   │   │   │   └── profile_handler.go
│   │   │   ├── routes/              # Route definitions
│   │   │   │   └── user_routes.go
│   │   │   ├── models/              # Data models
│   │   │   │   ├── user.go
│   │   │   │   ├── user_settings.go
│   │   │   │   └── profile.go
│   │   │   ├── account/             # User account subdomain
│   │   │   │   ├── wire.go
│   │   │   │   ├── handlers/
│   │   │   │   ├── services/
│   │   │   │   ├── repositories/
│   │   │   │   └── models/
│   │   │   └── errors/              # Domain-specific errors
│   │   │       └── user_errors.go
│   │   │
│   │   ├── workspace/               # Workspace management domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── workspace.go
│   │   │   │   └── workspace_member.go
│   │   │   ├── value_objects/
│   │   │   │   ├── workspace_id.go
│   │   │   │   ├── workspace_role.go
│   │   │   │   └── member_status.go
│   │   │   ├── repositories/
│   │   │   │   ├── workspace_repository.go
│   │   │   │   └── member_repository.go
│   │   │   ├── services/
│   │   │   │   ├── workspace_service.go
│   │   │   │   ├── member_service.go
│   │   │   │   └── permissions_service.go
│   │   │   ├── handlers/
│   │   │   │   ├── workspace_handler.go
│   │   │   │   └── member_handler.go
│   │   │   ├── routes/
│   │   │   │   └── workspace_routes.go
│   │   │   ├── models/
│   │   │   │   ├── workspace.go
│   │   │   │   ├── workspace_member.go
│   │   │   │   └── workspace_settings.go
│   │   │   └── errors/
│   │   │       └── workspace_errors.go
│   │   │
│   │   ├── instagram/               # Instagram integration domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── social_account.go
│   │   │   │   ├── instagram_media.go
│   │   │   │   └── instagram_comment.go
│   │   │   ├── value_objects/
│   │   │   │   ├── access_token.go
│   │   │   │   ├── media_type.go
│   │   │   │   └── sync_status.go
│   │   │   ├── repositories/
│   │   │   │   ├── social_account_repository.go
│   │   │   │   ├── media_repository.go
│   │   │   │   └── comment_repository.go
│   │   │   ├── services/
│   │   │   │   ├── oauth_service.go
│   │   │   │   ├── sync_service.go
│   │   │   │   ├── webhook_service.go
│   │   │   │   └── api_client_service.go
│   │   │   ├── handlers/
│   │   │   │   ├── oauth_handler.go
│   │   │   │   ├── sync_handler.go
│   │   │   │   └── webhook_handler.go
│   │   │   ├── routes/
│   │   │   │   └── instagram_routes.go
│   │   │   ├── models/
│   │   │   │   ├── social_account.go
│   │   │   │   ├── media.go
│   │   │   │   └── comment.go
│   │   │   ├── client/              # Instagram API client
│   │   │   │   ├── instagram_client.go
│   │   │   │   ├── oauth.go
│   │   │   │   ├── media.go
│   │   │   │   └── comments.go
│   │   │   └── errors/
│   │   │       └── instagram_errors.go
│   │   │
│   │   ├── moderation/              # AI moderation domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── moderation_rule.go
│   │   │   │   ├── moderation_action.go
│   │   │   │   └── comment_analysis.go
│   │   │   ├── value_objects/
│   │   │   │   ├── sentiment_score.go
│   │   │   │   ├── toxicity_score.go
│   │   │   │   └── confidence_level.go
│   │   │   ├── repositories/
│   │   │   │   ├── moderation_repository.go
│   │   │   │   └── analysis_repository.go
│   │   │   ├── services/
│   │   │   │   ├── ai_analysis_service.go
│   │   │   │   ├── moderation_service.go
│   │   │   │   ├── queue_service.go
│   │   │   │   └── rule_engine_service.go
│   │   │   ├── handlers/
│   │   │   │   ├── moderation_handler.go
│   │   │   │   └── analysis_handler.go
│   │   │   ├── routes/
│   │   │   │   └── moderation_routes.go
│   │   │   ├── models/
│   │   │   │   ├── moderation_rule.go
│   │   │   │   ├── moderation_action.go
│   │   │   │   └── comment_analysis.go
│   │   │   ├── ai/                  # AI integration
│   │   │   │   ├── gemini_client.go
│   │   │   │   ├── prompt_templates.go
│   │   │   │   ├── response_parser.go
│   │   │   │   └── cost_tracker.go
│   │   │   └── errors/
│   │   │       └── moderation_errors.go
│   │   │
│   │   ├── context/                 # Context management domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── context_document.go
│   │   │   │   └── context_assignment.go
│   │   │   ├── value_objects/
│   │   │   │   ├── document_type.go
│   │   │   │   ├── file_size.go
│   │   │   │   └── relevance_score.go
│   │   │   ├── repositories/
│   │   │   │   ├── document_repository.go
│   │   │   │   └── assignment_repository.go
│   │   │   ├── services/
│   │   │   │   ├── document_service.go
│   │   │   │   ├── upload_service.go
│   │   │   │   ├── extraction_service.go
│   │   │   │   └── matching_service.go
│   │   │   ├── handlers/
│   │   │   │   ├── document_handler.go
│   │   │   │   └── upload_handler.go
│   │   │   ├── routes/
│   │   │   │   └── context_routes.go
│   │   │   ├── models/
│   │   │   │   ├── context_document.go
│   │   │   │   └── context_assignment.go
│   │   │   ├── processors/          # Document processing
│   │   │   │   ├── pdf_processor.go
│   │   │   │   ├── word_processor.go
│   │   │   │   └── text_processor.go
│   │   │   └── errors/
│   │   │       └── context_errors.go
│   │   │
│   │   ├── otto/                    # Otto AI chat domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── chat_session.go
│   │   │   │   ├── chat_message.go
│   │   │   │   └── query_intent.go
│   │   │   ├── value_objects/
│   │   │   │   ├── message_type.go
│   │   │   │   ├── session_id.go
│   │   │   │   └── intent_confidence.go
│   │   │   ├── repositories/
│   │   │   │   ├── chat_repository.go
│   │   │   │   └── session_repository.go
│   │   │   ├── services/
│   │   │   │   ├── chat_service.go
│   │   │   │   ├── nlp_service.go
│   │   │   │   ├── query_service.go
│   │   │   │   └── websocket_service.go
│   │   │   ├── handlers/
│   │   │   │   ├── chat_handler.go
│   │   │   │   └── websocket_handler.go
│   │   │   ├── routes/
│   │   │   │   └── otto_routes.go
│   │   │   ├── models/
│   │   │   │   ├── chat_message.go
│   │   │   │   ├── chat_session.go
│   │   │   │   └── query_response.go
│   │   │   ├── intelligence/        # Otto AI brain
│   │   │   │   ├── intent_classifier.go
│   │   │   │   ├── query_processor.go
│   │   │   │   ├── response_generator.go
│   │   │   │   └── personality.go
│   │   │   └── errors/
│   │   │       └── otto_errors.go
│   │   │
│   │   ├── automation/              # Response automation domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── response_template.go
│   │   │   │   ├── automated_response.go
│   │   │   │   └── approval_queue.go
│   │   │   ├── value_objects/
│   │   │   │   ├── template_type.go
│   │   │   │   ├── trigger_keyword.go
│   │   │   │   └── approval_status.go
│   │   │   ├── repositories/
│   │   │   │   ├── template_repository.go
│   │   │   │   └── response_repository.go
│   │   │   ├── services/
│   │   │   │   ├── template_service.go
│   │   │   │   ├── automation_service.go
│   │   │   │   ├── matching_service.go
│   │   │   │   └── posting_service.go
│   │   │   ├── handlers/
│   │   │   │   ├── template_handler.go
│   │   │   │   └── automation_handler.go
│   │   │   ├── routes/
│   │   │   │   └── automation_routes.go
│   │   │   ├── models/
│   │   │   │   ├── response_template.go
│   │   │   │   └── automated_response.go
│   │   │   ├── engine/              # Automation engine
│   │   │   │   ├── rule_matcher.go
│   │   │   │   ├── template_renderer.go
│   │   │   │   └── response_scheduler.go
│   │   │   └── errors/
│   │   │       └── automation_errors.go
│   │   │
│   │   ├── analytics/               # Analytics and insights domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── analytics_metric.go
│   │   │   │   ├── trend_analysis.go
│   │   │   │   └── insight.go
│   │   │   ├── value_objects/
│   │   │   │   ├── metric_type.go
│   │   │   │   ├── date_range.go
│   │   │   │   └── trend_direction.go
│   │   │   ├── repositories/
│   │   │   │   ├── metrics_repository.go
│   │   │   │   └── insights_repository.go
│   │   │   ├── services/
│   │   │   │   ├── analytics_service.go
│   │   │   │   ├── calculation_service.go
│   │   │   │   ├── aggregation_service.go
│   │   │   │   └── trend_service.go
│   │   │   ├── handlers/
│   │   │   │   ├── analytics_handler.go
│   │   │   │   └── insights_handler.go
│   │   │   ├── routes/
│   │   │   │   └── analytics_routes.go
│   │   │   ├── models/
│   │   │   │   ├── analytics_metric.go
│   │   │   │   ├── dashboard_data.go
│   │   │   │   └── insight.go
│   │   │   ├── calculators/         # Metric calculators
│   │   │   │   ├── sentiment_calculator.go
│   │   │   │   ├── engagement_calculator.go
│   │   │   │   └── trend_calculator.go
│   │   │   └── errors/
│   │   │       └── analytics_errors.go
│   │   │
│   │   ├── dashboard/               # Dashboard domain
│   │   │   ├── wire.go
│   │   │   ├── entities/
│   │   │   │   ├── dashboard_widget.go
│   │   │   │   └── activity_feed.go
│   │   │   ├── repositories/
│   │   │   │   └── dashboard_repository.go
│   │   │   ├── services/
│   │   │   │   ├── dashboard_service.go
│   │   │   │   └── widget_service.go
│   │   │   ├── handlers/
│   │   │   │   └── dashboard_handler.go
│   │   │   ├── routes/
│   │   │   │   └── dashboard_routes.go
│   │   │   ├── models/
│   │   │   │   ├── dashboard_data.go
│   │   │   │   └── widget_config.go
│   │   │   └── errors/
│   │   │       └── dashboard_errors.go
│   │   │
│   │   └── email/                   # Email notifications domain
│   │       ├── wire.go
│   │       ├── entities/
│   │       │   └── email_template.go
│   │       ├── services/
│   │       │   ├── email_service.go
│   │       │   └── template_service.go
│   │       ├── models/
│   │       │   └── email.go
│   │       ├── templates/           # Email templates
│   │       │   ├── welcome.html
│   │       │   ├── password_reset.html
│   │       │   └── weekly_report.html
│   │       ├── jobs/                # Email job workers
│   │       │   ├── periodic/
│   │       │   │   └── weekly_report.go
│   │       │   └── scheduled/
│   │       │       └── notification.go
│   │       └── errors/
│   │           └── email_errors.go
│   │
│   └── shared/                      # Shared infrastructure
│       ├── registry/                # Dependency registry
│       │   ├── handler_registry.go  # HTTP handler registry
│       │   ├── service_registry.go  # Service registry
│       │   └── wire.go              # Registry DI
│       ├── river/                   # Background job queue
│       │   ├── queue.go             # Queue client
│       │   ├── service.go           # Queue service
│       │   ├── worker.go            # Job worker
│       │   ├── jobs/                # Job definitions
│       │   │   ├── email_job.go
│       │   │   ├── sync_job.go
│       │   │   ├── analysis_job.go
│       │   │   └── cleanup_job.go
│       │   ├── wire.go              # Queue DI
│       │   └── errors.go            # Queue errors
│       ├── cache/                   # Caching layer
│       │   ├── redis_client.go      # Redis client
│       │   ├── cache_service.go     # Cache service
│       │   ├── cache_keys.go        # Cache key constants
│       │   └── wire.go              # Cache DI
│       ├── storage/                 # File storage
│       │   ├── local_storage.go     # Local file storage
│       │   ├── cloud_storage.go     # Cloud storage (S3/GCS)
│       │   ├── storage_service.go   # Storage service interface
│       │   └── wire.go              # Storage DI
│       ├── monitoring/              # Monitoring and metrics
│       │   ├── prometheus.go        # Prometheus metrics
│       │   ├── health_check.go      # Health check endpoints
│       │   ├── profiling.go         # Performance profiling
│       │   └── wire.go              # Monitoring DI
│       ├── events/                  # Event system
│       │   ├── event_bus.go         # Event bus implementation
│       │   ├── event_handler.go     # Event handler interface
│       │   ├── publishers/          # Event publishers
│       │   │   ├── user_events.go
│       │   │   └── comment_events.go
│       │   ├── subscribers/         # Event subscribers
│       │   │   ├── notification_subscriber.go
│       │   │   └── analytics_subscriber.go
│       │   └── wire.go              # Events DI
│       ├── websocket/               # WebSocket support
│       │   ├── hub.go               # WebSocket hub
│       │   ├── client.go            # WebSocket client
│       │   ├── connection.go        # Connection management
│       │   └── wire.go              # WebSocket DI
│       └── utils/                   # Shared utilities
│           ├── crypto/              # Cryptographic utilities
│           │   ├── hash.go
│           │   ├── encrypt.go
│           │   └── random.go
│           ├── time/                # Time utilities
│           │   ├── timezone.go
│           │   └── formatting.go
│           ├── validator/           # Custom validators
│           │   ├── email.go
│           │   ├── phone.go
│           │   └── password.go
│           └── http/                # HTTP utilities
│               ├── client.go
│               ├── retry.go
│               └── rate_limiter.go
│
├── logs/                            # Application logs
│   ├── app.log                     # Main application log
│   ├── error.log                   # Error logs
│   ├── access.log                  # HTTP access logs
│   ├── database.log                # Database query logs
│   └── audit.log                   # Security audit logs
│
├── pkg/                            # Public packages (reusable)
│   ├── error/                      # Error handling package
│   │   ├── core.go                 # Core error types
│   │   ├── codes.go                # Error codes
│   │   ├── handler.go              # Error handler
│   │   └── wrapper.go              # Error wrapper
│   ├── logging/                    # Logging package
│   │   ├── logging.go              # Logger setup
│   │   ├── wire.go                 # Logging DI
│   │   ├── formatters/             # Log formatters
│   │   │   ├── json.go
│   │   │   └── text.go
│   │   └── hooks/                  # Log hooks
│   │       ├── file_hook.go
│   │       └── sentry_hook.go
│   ├── metrics/                    # Metrics package
│   │   ├── metrics.go              # Metrics definitions
│   │   ├── prometheus.go           # Prometheus integration
│   │   ├── collectors/             # Custom collectors
│   │   │   ├── http_collector.go
│   │   │   └── db_collector.go
│   │   └── middleware.go           # Metrics middleware
│   ├── security/                   # Security utilities
│   │   ├── jwt/                    # JWT utilities
│   │   │   ├── token.go
│   │   │   ├── claims.go
│   │   │   └── validator.go
│   │   ├── crypto/                 # Cryptography
│   │   │   ├── bcrypt.go
│   │   │   ├── aes.go
│   │   │   └── rsa.go
│   │   └── rate_limit/             # Rate limiting
│   │       ├── limiter.go
│   │       └── memory_store.go
│   ├── pagination/                 # Pagination utilities
│   │   ├── paginator.go
│   │   ├── cursor.go
│   │   └── response.go
│   ├── validation/                 # Validation utilities
│   │   ├── validator.go
│   │   ├── rules.go
│   │   └── custom_rules.go
│   └── migration/                  # Migration utilities
│       ├── migrator.go
│       ├── runner.go
│       └── schema.go
│
├── scripts/                        # Development and deployment scripts
│   ├── build/                     # Build scripts
│   │   ├── build.sh               # Build application
│   │   ├── docker-build.sh        # Docker build
│   │   └── cross-compile.sh       # Cross-platform builds
│   ├── dev/                       # Development scripts
│   │   ├── setup.sh              # Development setup
│   │   ├── reset-db.sh           # Database reset
│   │   ├── seed-db.sh            # Database seeding
│   │   └── test.sh               # Run tests
│   ├── deploy/                    # Deployment scripts
│   │   ├── deploy.sh             # General deployment
│   │   ├── migrate.sh            # Database migration
│   │   └── rollback.sh           # Rollback deployment
│   └── maintenance/               # Maintenance scripts
│       ├── backup.sh             # Database backup
│       ├── cleanup.sh            # Log cleanup
│       └── health-check.sh       # Health verification
│
├── storage/                       # File storage (development)
│   ├── uploads/                   # User uploads
│   │   ├── documents/            # Context documents
│   │   ├── images/               # Images
│   │   └── temp/                 # Temporary files
│   ├── cache/                     # File cache
│   └── exports/                   # Data exports
│
├── test/                         # Test files
│   ├── integration/              # Integration tests
│   │   ├── api/                 # API integration tests
│   │   │   ├── auth_test.go
│   │   │   ├── workspace_test.go
│   │   │   └── instagram_test.go
│   │   ├── database/            # Database integration tests
│   │   │   ├── migration_test.go
│   │   │   └── transaction_test.go
│   │   └── external/            # External service tests
│   │       ├── instagram_api_test.go
│   │       └── gemini_api_test.go
│   ├── unit/                    # Unit tests
│   │   ├── domain/              # Domain unit tests
│   │   │   ├── auth/
│   │   │   │   ├── services_test.go
│   │   │   │   └── handlers_test.go
│   │   │   ├── user/
│   │   │   └── workspace/
│   │   ├── pkg/                 # Package unit tests
│   │   │   ├── error_test.go
│   │   │   ├── logging_test.go
│   │   │   └── metrics_test.go
│   │   └── shared/              # Shared component tests
│   │       ├── cache_test.go
│   │       └── queue_test.go
│   ├── mocks/                   # Mock implementations
│   │   ├── repositories/        # Repository mocks
│   │   │   ├── user_mock.go
│   │   │   └── workspace_mock.go
│   │   ├── services/            # Service mocks
│   │   │   ├── auth_mock.go
│   │   │   └── email_mock.go
│   │   └── external/            # External service mocks
│   │       ├── instagram_mock.go
│   │       └── gemini_mock.go
│   ├── fixtures/                # Test data fixtures
│   │   ├── users.json
│   │   ├── workspaces.json
│   │   └── comments.json
│   ├── testdata/                # Test files and data
│   │   ├── documents/           # Test documents
│   │   ├── images/              # Test images
│   │   └── responses/           # API response samples
│   └── e2e/                     # End-to-end tests
│       ├── user_journey_test.go
│       ├── automation_flow_test.go
│       └── analytics_test.go
│
└── tmp/                         # Temporary files
    ├── pids/                    # Process IDs
    ├── uploads/                 # Temporary uploads
    └── logs/                    # Temporary logs
```

---

## 🎯 Architecture Principles

### **1. Domain-Driven Design (DDD)**
- **Clear domain boundaries** with separate folders for each business domain
- **Entities, Value Objects, and Services** properly separated
- **Repository pattern** for data access abstraction
- **Domain events** for loose coupling between domains

### **2. Clean Architecture**
- **Dependency inversion** - inner layers don't depend on outer layers
- **Interface segregation** - small, focused interfaces
- **Single responsibility** - each package has one clear purpose
- **Separation of concerns** - business logic separate from infrastructure

### **3. Go Best Practices**
- **`internal/` package** for private application code
- **`pkg/` package** for reusable public packages
- **Wire dependency injection** for compile-time DI
- **Clear module structure** following Go conventions

---

## 📋 Folder Responsibilities

### **Core Application (`internal/`)**

#### **API Layer (`internal/api/`)**
- HTTP routing and middleware
- Request/response handling
- API versioning
- CORS, rate limiting, authentication

#### **Domain Layer (`internal/domain/`)**
- **Business logic and rules**
- **Entity definitions and validation**
- **Repository interfaces**
- **Domain services**
- **HTTP handlers for each domain**
- **Route definitions**

#### **Infrastructure (`internal/shared/`)**
- **Database connections and migrations**
- **Caching (Redis)**
- **Background job processing**
- **File storage**
- **Monitoring and metrics**
- **WebSocket support**

### **Public Packages (`pkg/`)**
- **Reusable utilities** that could be extracted as libraries
- **Error handling framework**
- **Logging infrastructure**
- **Security utilities**
- **Common validation and pagination**

### **External Integration**
- **Instagram API client** in `internal/domain/instagram/client/`
- **Google Gemini AI** in `internal/domain/moderation/ai/`
- **Email service** in `internal/domain/email/`

---

## 🔧 Development Guidelines

### **Adding New Features**
1. **Identify the domain** - which business area does this belong to?
2. **Create domain structure** - entities, repositories, services, handlers
3. **Define interfaces** - repository and service interfaces first
4. **Implement business logic** - in domain services
5. **Add HTTP layer** - handlers and routes
6. **Wire dependencies** - update wire.go files
7. **Add tests** - unit tests for each layer

### **Database Changes**
1. **Create migration files** in `internal/database/migrations/`
2. **Update domain models** in relevant domain folder
3. **Update repository interfaces** and implementations
4. **Add integration tests** for new queries

### **New Domain Addition**
```
internal/domain/new_domain/
├── wire.go
├── entities/
├── value_objects/
├── repositories/
├── services/
├── handlers/
├── routes/
├── models/
└── errors/
```

### **Testing Strategy**
- **Unit tests** for each service and handler
- **Integration tests** for database operations
- **End-to-end tests** for complete workflows
- **Mocks** for external dependencies
- **Test fixtures** for consistent test data

---

## 🚀 Benefits of This Structure

### **1. Maintainability**
- **Clear separation of concerns**
- **Easy to locate and modify code**
- **Minimal coupling between domains**
- **Consistent patterns across all domains**

### **2. Scalability**
- **Independent domain development**
- **Easy to extract microservices later**
- **Clear dependency management**
- **Horizontal scaling ready**

### **3. Testing**
- **Easy to mock dependencies**
- **Clear test organization**
- **Fast unit test execution**
- **Comprehensive test coverage**

### **4. Team Collaboration**
- **Domain ownership possible**
- **Minimal merge conflicts**
- **Clear code review boundaries**
- **Easy onboarding for new developers**

### **5. Future Growth**
- **Easy to add new social platforms** (TikTok, YouTube)
- **Simple feature extension**
- **Microservices extraction ready**
- **Plugin architecture possible**

---

## 📝 Migration from Current Structure

### **Current → New Structure Mapping**

```
Current                          →  New Structure
├── internal/domain/auth/        →  internal/domain/auth/
├── internal/domain/user/        →  internal/domain/user/
├── internal/domain/email/       →  internal/domain/email/
├── internal/api/                →  internal/api/
├── internal/bootstrap/          →  internal/bootstrap/
├── internal/config/             →  internal/config/
├── internal/database/           →  internal/database/
├── internal/shared/             →  internal/shared/
└── pkg/                         →  pkg/

New Additions:
├── internal/domain/workspace/    (New domain)
├── internal/domain/instagram/    (New domain)
├── internal/domain/moderation/   (New domain)
├── internal/domain/context/      (New domain)
├── internal/domain/otto/         (New domain)
├── internal/domain/automation/   (New domain)
├── internal/domain/analytics/    (New domain)
└── internal/domain/dashboard/    (New domain)
```

### **Implementation Steps**
1. **Week 1**: Set up foundation domains (workspace, instagram)
2. **Week 2-3**: Add AI domains (moderation, context)
3. **Week 4-5**: Implement Otto and automation domains
4. **Week 6-7**: Add analytics and dashboard domains
5. **Week 8**: Refactor existing code to new structure
6. **Week 9**: Add comprehensive testing structure
7. **Week 10**: Documentation and deployment setup

---

## 🎉 Conclusion

This folder structure provides a **solid foundation** for building the Refyne MVP while ensuring **long-term maintainability** and **scalability**. It follows Go best practices, implements Domain-Driven Design principles, and provides clear separation of concerns that will support your team as the product grows.

The structure is designed to handle all the features outlined in your MVP TODO list while remaining flexible enough to accommodate future enhancements and potential microservices extraction.
