# Static Admin

A modern web interface for managing static site content. Built for GitHub Pages and Jekyll sites.

## Features

- 🎨 Modern, intuitive interface
- 🔄 Seamless GitHub integration
- 📝 Rich text editor with Markdown support
- 🎯 Built for Jekyll and GitHub Pages
- 📱 Responsive design

## Development

### Requirements

- Go 1.23+
- Node.js 18+

### Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/josegonzalez/static-admin.git
   cd static-admin
   ```

2. Install backend dependencies:

   ```bash
   go mod download
   ```

3. Install frontend dependencies:

   ```bash
   cd frontend
   npm install
   ```

4. Create a GitHub OAuth App:
   - Go to GitHub Settings > Developer Settings > OAuth Apps
   - Create a new OAuth App
   - Set the callback URL to `http://localhost:8080/auth/github/callback`
   - Copy the Client ID and Client Secret

5. Create a `.env` file in the root directory:

   ```env
   GITHUB_CLIENT_ID=your_client_id
   GITHUB_CLIENT_SECRET=your_client_secret
   GITHUB_REDIRECT_URL=http://localhost:8080/auth/github/callback
   JWT_SECRET=your_jwt_secret
   SESSION_SECRET=your_session_secret
   ```

### Running

1. Start the backend server:

   ```bash
   go run main.go
   ```

2. Start the frontend development server:

   ```bash
   cd frontend
   npm run dev
   ```

3. Visit `http://localhost:3000`

### Deploying

We include a Dockerfile that compiles both the frontend and backend. This can be used to deploy to any Docker-compatible platform, and will respect the `PORT` environment variable. By default, it will run on port 5000.

To use it locally, run the following commands:

1. Build the docker image:

   ```bash
   docker build -t static-admin .
   ```

2. Run the docker container:

   ```bash
   docker run --rm -p 8080:8080 static-admin
   ```

3. Visit `http://localhost:8080`

### Releasing

1. Build the frontend:

   ```bash
   cd frontend
   npm run build
   ```

2. Generate the embedded frontend golang code:

   ```bash
   go generate ./...
   ```

3. Build the backend:

   ```bash
   go build -o static-admin main.go
   ```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE.md](LICENSE.md) for details
