# Docker MongoDB Setup for Smart 360

This setup provides a simple MongoDB development environment using Docker Compose.

## Quick Start

1. **Start MongoDB with Docker:**
   ```bash
   docker-compose up -d
   ```

2. **Access MongoDB:**
   - **Application:** `mongodb://localhost:27017`
   - **Mongo Express (Web UI):** http://localhost:8081
     - Username: `admin`
     - Password: `admin123`

3. **Start the application:**
   ```bash
   # Copy the environment file
   cp .env.example .env
   
   # Start the backend
   go run main.go
   ```

## Environment Configuration

Update your `.env` file with MongoDB settings:

```bash
# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=smart360

# OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback

# Application Configuration
FRONTEND_URL=http://localhost:5173
JWT_SECRET=your-random-secret-key-change-in-production
```

## Docker Services

### MongoDB
- **Image:** `mongo:8.0` (latest version)
- **Port:** `27017:27017`
- **Database:** `smart360`
- **Authentication:** Enabled with admin user
- **Data Persistence:** Uses Docker volume `mongodb_data`

### Mongo Express (Optional)
- **Image:** `mongo-express:latest`
- **Port:** `8081:8081`
- **Web UI:** http://localhost:8081
- **Credentials:** admin/admin123

## Database Initialization

The MongoDB container automatically runs the initialization script at `scripts/mongo-init.js` which:

1. Creates the `smart360` database
2. Sets up collections with validation rules
3. Creates performance indexes
4. Prepares the database for the application

## Development Workflow

1. **Start MongoDB:**
   ```bash
   docker-compose up -d mongodb
   ```

2. **Run application:**
   ```bash
   go run main.go
   ```

3. **Seed data:** The application will automatically seed development data on first run

4. **Access database:** Use Mongo Express at http://localhost:8081 or connect with MongoDB Compass

## Useful Commands

```bash
# Start all services
docker-compose up -d

# Start only MongoDB
docker-compose up -d mongodb

# View logs
docker-compose logs -f mongodb

# Stop all services
docker-compose down

# Stop and remove data (WARNING: deletes all data)
docker-compose down -v
```

## Production Considerations

For production deployment:

1. **Change default passwords** in docker-compose.yml
2. **Use environment variables** for sensitive data
3. **Enable authentication** with proper user roles
4. **Set up backups** for MongoDB data
5. **Use MongoDB Atlas** for managed database service

## Troubleshooting

### MongoDB Connection Issues
- Ensure Docker is running
- Check if port 27017 is available
- Verify MongoDB container is healthy: `docker-compose ps`

### Data Persistence
- Data is stored in Docker volume `mongodb_data`
- To reset data: `docker-compose down -v` then `docker-compose up -d`

### Performance
- For better performance, consider increasing Docker memory allocation
- Monitor MongoDB logs: `docker-compose logs mongodb`

## Next Steps

1. Configure your OAuth credentials in `.env`
2. Start the frontend application
3. Access the application at http://localhost:5173
4. Use Mongo Express to explore the database structure
