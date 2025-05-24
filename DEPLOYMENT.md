# SnapURL Deployment Guide

## Docker

### Build and Run

```bash
docker build -t snapurl .
docker run -d -p 8080:8080 -v snapurl_data:/data snapurl
