# Deploy Focalboard to Render

Simple guide to deploy Focalboard on Render.com using the free tier.

## Prerequisites

- GitHub account with this repository
- Render.com free account

## Deployment Steps

### 1. Create PostgreSQL Database

1. Go to Render Dashboard → "New" → "PostgreSQL"
2. Name: `focalboard-db`
3. Database Name: `focalboard`
4. User: `focalboard`
5. Region: Choose closest to you
6. Plan: Free (shared CPU, 1GB storage)
7. Click "Create Database"
8. **Copy the External Database URL** (you'll need this)

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

### 3. Environment Variables

Add these environment variables:

```
DATABASE_URL=[Paste your PostgreSQL External Database URL here]
FOCALBOARD_SERVERROOT=https://your-service-name.onrender.com
FOCALBOARD_PORT=8000
FOCALBOARD_USESSL=true
FOCALBOARD_SECURECOOKIE=true
FOCALBOARD_TELEMETRY=false
FOCALBOARD_AUTHMODE=native
FOCALBOARD_ENABLEPUBLICSHAREDBOARDS=false
FOCALBOARD_FILESDRIVER=local
FOCALBOARD_FILESPATH=./data/files
```

**Important**: Replace `your-service-name` with your actual Render service name.

### 4. Deploy

1. Click "Create Web Service"
2. Render will build your Docker image and deploy
3. First build takes ~5-10 minutes
4. Your app will be available at `https://your-service-name.onrender.com`

## Notes

- **Free tier limitations**: App sleeps after 15 minutes of inactivity
- **Storage**: Files stored locally (lost on restart unless using paid persistent disk)
- **Database**: 1GB PostgreSQL on free tier
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