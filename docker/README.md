# Docker Image Management Guide

This guide provides instructions for managing Docker images for both development and production environments. It covers how to publish new releases to GitHub Packages using GitHub Actions, and how to build and tag images locally for development.

## Publishing a New Release to GitHub Packages

To publish a new release of Docker image to GitHub Packages:

1. Navigate to the **Actions** tab in the repository.
2. Find the workflow named **Publish Docker image to GitHub Packages** or similar.
3. Click on the **Run workflow** dropdown button, usually located on the right side.
4. Select the reference for publishing (branch or tag).
5. Click the **Run workflow** button to start the manual workflow.

To specify image name set repository variable `AUTH_SERVER__IMAGE_NAME`.
This process will build and push the production Docker image to GitHub Packages.

**Tagging options**. Image tag depends on the reference of your choise:
- `main` branch reference will result in tagging the image as `latest`;
- `v*` tag (e.g. `v1.0.0`) will just set an image tag as whatever goes after `v`.

## Building Images Locally

To build Docker images for development and production environments locally, you can use the following commands from the root of your project directory:

### Development Image

To build a Docker image for development, which includes additional tools and configurations helpful during the development process, run this command from the **root** folder of the project:

```sh
docker build -f docker/dev/Dockerfile -t your-image:your-tag .
```

### Production Image

Similarly, for building a production-ready Docker image, use the following command:

```sh
docker build -f docker/prod/Dockerfile -t your-image:your-tag .
```

## Remote Server Deployment

To utilize this image from the GitHub Container Registry, you might require an access token with the `packages:read` permission.
