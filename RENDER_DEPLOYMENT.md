# Deploy Focalboard to Render

Simple guide to deploy Focalboard on Render.com using the free tier.

## Prerequisites

- GitHub account with this repository
- Render.com free account
- Optional: Supabase.com free account (for alternative database)

## Deployment Steps

### 1. Create PostgreSQL Database

**Option A: Render PostgreSQL (Free tier - 1GB)**

1. Go to Render Dashboard → "New" → "PostgreSQL"
2. Name: `focalboard-db`
3. Database Name: `focalboard`
4. User: `focalboard`
5. Region: Choose closest to you
6. Plan: Free (shared CPU, 1GB storage)
7. Click "Create Database"
8. **Copy the External Database URL** (you'll need this)

**Option B: Supabase PostgreSQL (Free tier - 500MB)**

1. Go to [supabase.com](https://supabase.com) → "New Project"
2. Organization: Create or select
3. Name: `focalboard`
4. Database Password: Generate strong password
5. Region: Choose closest to you
6. Plan: Free (500MB database, 50MB file storage)
7. Click "Create new project"
8. **Get the Pooler Connection String**:
   - Follow Supabase's guide: https://supabase.com/docs/guides/database/connecting-to-postgres#supavisor-session-mode
   - Use the **Session Mode** connection string (contains `pooler.supabase.com`)
   - Format: `postgresql://postgres.[PROJECT_ID]:[YOUR_PASSWORD]@aws-0-[REGION].pooler.supabase.com:5432/postgres`
   - **Important**: Use the pooler format to avoid IPv6 connection issues on Render

### 2. Create Web Service

1. Go to Render Dashboard → "New" → "Web Service"
2. Connect GitHub repository: `https://github.com/0xanonymeow/focalboard`
3. Configure:
   - **Name**: `focalboard` (or your preferred name)
   - **Region**: Same as your database
   - **Branch**: `main`
   - **Runtime**: `Docker`
   - **Dockerfile Path**: `./docker/Dockerfile`
   - **Docker Context**: `./`

### 3. File Storage Setup (Choose One)

**Option A: Local Storage (Simple, files lost on restart)**
```
FOCALBOARD_FILESDRIVER=local
FOCALBOARD_FILESPATH=./data/files
```

**Option B: AWS S3 (Persistent storage)**
1. Create AWS S3 bucket
2. Create IAM user with S3 access
3. Add these environment variables:
```
FOCALBOARD_FILESDRIVER=amazons3
FOCALBOARD_FILESS3CONFIG_BUCKET=your-bucket-name
FOCALBOARD_FILESS3CONFIG_REGION=us-east-1
FOCALBOARD_FILESS3CONFIG_ACCESSKEYID=your-access-key
FOCALBOARD_FILESS3CONFIG_SECRETACCESSKEY=your-secret-key
FOCALBOARD_FILESS3CONFIG_SSL=true
```

**Option C: Cloudflare R2 (S3-compatible, cheaper)**
1. **Create R2 bucket**:
   - Go to Cloudflare Dashboard → R2 Object Storage
   - Click "Create bucket"
   - Choose bucket name and location
2. **Create R2 API token**:
   - Go to Cloudflare Dashboard → Account Home → R2
   - Under **API** dropdown, select "Manage API tokens"
   - Choose "Create User API token" (or "Create Account API token" for account-wide access)
   - Under **Permissions**: Select "Object Read & Write"
   - (Optional) Scope token to specific bucket if desired
   - Click "Create User API token"
   - **Important**: Copy both the **Access Key ID** and **Secret Access Key** immediately (Secret Access Key is only shown once)
   - Note your **Account ID** from the R2 dashboard
3. **Add these environment variables**:
```
FOCALBOARD_FILESDRIVER=amazons3
FOCALBOARD_FILESS3CONFIG_BUCKET=your-bucket-name
FOCALBOARD_FILESS3CONFIG_REGION=auto
FOCALBOARD_FILESS3CONFIG_ACCESSKEYID=your-access-key-id
FOCALBOARD_FILESS3CONFIG_SECRETACCESSKEY=your-secret-access-key
FOCALBOARD_FILESS3CONFIG_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
FOCALBOARD_FILESS3CONFIG_SSL=true
```
Replace:
- `your-bucket-name`: Your R2 bucket name
- `your-access-key-id`: Access Key ID from step 2
- `your-secret-access-key`: Secret Access Key from step 2  
- `your-account-id`: Your Cloudflare Account ID

### 4. Environment Variables

Add these common environment variables plus your chosen storage option above:

```
DATABASE_URL=[Paste your PostgreSQL connection string here]
FOCALBOARD_PORT=8000
FOCALBOARD_USESSL=true
FOCALBOARD_SECURECOOKIE=true
FOCALBOARD_TELEMETRY=false
FOCALBOARD_AUTHMODE=native
FOCALBOARD_ENABLEPUBLICSHAREDBOARDS=false
```

**Note**: Skip `FOCALBOARD_SERVERROOT` for now - you'll add it after deployment.

### 5. Deploy

1. Click "Create Web Service"
2. Render will build your Docker image and deploy
3. First build takes ~5-10 minutes
4. Your app will be available at `https://your-service-name.onrender.com`

### 6. Add Server Root (After Deployment)

**Option A: Use Render's default URL**
1. Once deployed, note your actual Render URL (e.g., `https://focalboard-abc123.onrender.com`)
2. Go to your service → Environment
3. Add this environment variable:
   ```
   FOCALBOARD_SERVERROOT=https://your-actual-render-url.onrender.com
   ```
4. Click "Save Changes" - this will trigger a redeploy

**Option B: Use your own custom domain**
1. Set up custom domain in Render:
   - Go to your service → Settings → Custom Domains
   - Add your domain (e.g., `focalboard.yourdomain.com`)
   - Configure DNS CNAME to point to your Render service
2. Add this environment variable:
   ```
   FOCALBOARD_SERVERROOT=https://focalboard.yourdomain.com
   ```
3. Click "Save Changes" - this will trigger a redeploy

## Notes

- **Free tier limitations**: App sleeps after 15 minutes of inactivity
- **Storage**: 
  - Local: Files lost on restart (free)
  - S3/R2: Persistent storage (small cost for storage)
- **Database options**: 
  - Render PostgreSQL: 1GB storage, expires after 90 days
  - Supabase PostgreSQL: 500MB storage, no expiration, includes dashboard
- **Environment variables**: Your fixes ensure proper configuration override

## Testing

1. Visit your Render URL
2. Create account and test board creation
3. Verify port 8000 is working (should show in logs)
4. Test file uploads (stored locally)

## Troubleshooting

- Check build logs if deployment fails
- Verify DATABASE_URL is correctly set
- Ensure FOCALBOARD_SERVERROOT matches your actual URL
- **IPv6 connection errors**: Use Supabase's pooler connection string (e.g., `aws-0-region.pooler.supabase.com`) instead of the direct database connection
- **".env file not found" warning**: This is normal in production - the app will use environment variables