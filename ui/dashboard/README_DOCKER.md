# DRS Dashboard - Docker Setup

React + TypeScript dashboard for monitoring DRS (Data Recording System) modules.

## Features

- **Module Status Monitoring**: Real-time status of ECU modules (ecu0, ecu1, nas)
- **Recording Control**: Start/stop recording across all modules
- **Time Synchronization**: PTP sync status monitoring
- **Topic Rate Monitoring**: ROS2 topic rate analysis
- **System Control**: System-wide restart/shutdown operations
- **Auto-refresh**: 5-second interval data updates

## Docker Deployment

### Build and Run

```bash
# Build Docker image
docker build -t drs-dashboard .

# Run with Docker Compose
docker-compose up -d

# Or run directly
docker run -p 3000:80 drs-dashboard
```

The dashboard will be available at http://localhost:3000/

### Docker Features

- **Multi-stage build**: Optimized production image
- **Nginx reverse proxy**: Handles API requests and static files
- **Health checks**: Container health monitoring
- **CORS handling**: API proxy to avoid CORS issues
- **Gzip compression**: Optimized asset delivery
- **Security headers**: Basic security hardening

### Environment

- **API Gateway**: http://192.168.20.10:8080/api/v1
- **Container Port**: 3000 (host network)
- **Network**: host

## API Integration

The dashboard integrates with the DRS API Gateway:

- `/modules` - Module status
- `/recording/status` - Recording status
- `/ptp/status` - PTP synchronization
- `/modules/{hostname}/topics/status` - Topic rates
- System control endpoints for restart/shutdown

## Development Notes

- Uses shadcn/ui components with Tailwind CSS
- Implements 5-second auto-refresh for all data
- Graceful error handling with mock data fallback
- Responsive design for various screen sizes
- TypeScript for type safety
