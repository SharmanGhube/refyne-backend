# Deployment Guide - Refyne Backend

## Overview
This guide covers deploying the Refyne backend to Railway with CI/CD via GitHub Actions.

## Prerequisites
- GitHub account with repository access
- Railway account (https://railway.app)
- Git installed locally

---

## Step 1: Railway Setup

### 1.1 Create Railway Project
1. Go to https://railway.app and sign in
2. Click "New Project"
3. Select "Deploy from GitHub repo"
4. Connect your GitHub account and select `refyne-backend` repository
5. Railway will automatically detect the Dockerfile

### 1.2 Add PostgreSQL Database
1. In your Railway project, click "+ New"
2. Select "Database" → "Add PostgreSQL"
3. Railway will create and connect the database automatically

### 1.3 Add Redis
1. Click "+ New" again
2. Select "Database" → "Add Redis"
3. Railway will provision Redis

### 1.4 Configure Environment Variables
In Railway project → Backend service → Variables tab, add:

```bash
# Application
APP_ENV=production
APP_PORT=8080

# Database (Railway auto-provides these as DATABASE_URL, but we use separate vars)
DB_HOST=${{Postgres.PGHOST}}
DB_PORT=${{Postgres.PGPORT}}
DB_USER=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
DB_NAME=${{Postgres.PGDATABASE}}
DB_SSL_MODE=require
DB_MAX_CONNECTIONS=20
DB_MAX_IDLE_CONNECTIONS=10
AUTO_MIGRATE=true

# Redis
REDIS_HOST=${{Redis.REDIS_HOST}}
REDIS_PORT=${{Redis.REDIS_PORT}}
REDIS_PASSWORD=${{Redis.REDIS_PASSWORD}}
REDIS_DB=0

# JWT Configuration
JWT_SECRET=<GENERATE_SECURE_SECRET_HERE>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Frontend
FRONTEND_URL=https://yourfrontend.com

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@refyne.com

# Paddle (Subscription)
PADDLE_API_KEY=<your-paddle-api-key>
PADDLE_WEBHOOK_SECRET=<your-paddle-webhook-secret>
PADDLE_ENVIRONMENT=sandbox  # Change to 'production' when ready

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_STORE=redis
```

**Generate JWT Secret:**
```bash
openssl rand -base64 64
```

### 1.5 Get Railway Token for GitHub Actions
1. Go to Railway Account Settings → Tokens
2. Create new token named "GitHub Actions"
3. Copy the token (keep it safe!)

---

## Step 2: GitHub Secrets Setup

### 2.1 Add Railway Token
1. Go to your GitHub repository
2. Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `RAILWAY_TOKEN`
5. Value: Paste the Railway token from Step 1.5
6. Click "Add secret"

---

## Step 3: Deploy

### 3.1 Initial Deployment
Railway will auto-deploy on first setup. Monitor the deployment:
1. Go to Railway project → Deployments tab
2. Watch build logs
3. Once deployed, get the public URL from Settings → Domains

### 3.2 CI/CD Deployments
After initial setup, every push to `main` branch triggers:

1. **CI Workflow** (`.github/workflows/ci.yml`)
   - Linting (golangci-lint)
   - Testing with PostgreSQL + Redis
   - Security scanning (gosec)
   - Docker build validation
   - Code coverage upload

2. **Deploy Workflow** (`.github/workflows/deploy.yml`)
   - Automatic deployment to Railway
   - Only runs on `main` branch

### 3.3 Manual Deployment
If needed, deploy manually:
```bash
# Install Railway CLI
npm i -g @railway/cli

# Login
railway login

# Link to project
railway link

# Deploy
railway up
```

---

## Step 4: Verify Deployment

### 4.1 Health Checks
Test the deployed service:
```bash
# Basic health
curl https://your-app.railway.app/api/health

# Detailed health
curl https://your-app.railway.app/api/health/detailed

# Readiness
curl https://your-app.railway.app/api/health/ready
```

### 4.2 Test Authentication Flow
```bash
# Register user
curl -X POST https://your-app.railway.app/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "first_name": "Test",
    "last_name": "User"
  }'

# Request OTP
curl -X POST https://your-app.railway.app/api/auth/request-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'
```

---

## Step 5: Monitoring

### 5.1 Railway Logs
- Real-time logs: Railway Dashboard → Your Service → Logs
- Filter by severity, time range

### 5.2 Metrics
Railway provides:
- CPU usage
- Memory usage
- Network traffic
- Request counts

### 5.3 Alerts
Set up in Railway:
- Deploy notifications (Slack, Discord, Email)
- Resource usage alerts

---

## Step 6: Domain Setup (Optional)

### 6.1 Add Custom Domain
1. Railway Project → Settings → Domains
2. Click "Add Domain"
3. Enter your domain (e.g., `api.refyne.com`)
4. Add DNS records as instructed:
   - CNAME record pointing to Railway's domain
5. SSL automatically provisioned

---

## Troubleshooting

### Build Fails
- Check Dockerfile syntax
- Verify all dependencies in go.mod
- Review build logs in Railway/GitHub Actions

### Runtime Errors
- Check environment variables are set correctly
- Verify database migrations ran (`AUTO_MIGRATE=true`)
- Check service logs for stack traces

### Database Connection Issues
- Ensure PostgreSQL service is running
- Verify connection string format
- Check SSL mode setting (`DB_SSL_MODE=require`)

### Redis Connection Issues
- Ensure Redis service is running
- Verify REDIS_PASSWORD is set
- Check REDIS_HOST and REDIS_PORT

---

## Migration to AWS (Future)

When ready to migrate to AWS:

1. **Infrastructure as Code**
   - Set up Terraform/Pulumi for ECS Fargate
   - Create RDS PostgreSQL + ElastiCache Redis
   - Configure VPC, security groups, load balancer

2. **Data Migration**
   - PostgreSQL: `pg_dump` from Railway → `pg_restore` to RDS
   - Redis: Export/import or reconfigure

3. **Update GitHub Actions**
   - Change deploy.yml to use AWS credentials
   - Deploy to ECR → ECS instead of Railway

4. **DNS Update**
   - Point domain to AWS Application Load Balancer

**The same Dockerfile works on both platforms!**

---

## Support

- Railway Docs: https://docs.railway.app
- GitHub Actions: https://docs.github.com/actions
- Refyne Issues: https://github.com/refynehq/refyne-backend/issues
