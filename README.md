# HPTest - High Performance Test Backend

A lightweight, high-performance web backend written in Go for testing web servers, load balancers, and performance monitoring tools.

## 🚀 Features

- **Multi-protocol Support**: HTTP/1.1 with various response types
- **Dynamic Content Generation**: On-the-fly file and image generation
- **Performance Metrics**: Real-time request statistics and tracking
- **File Upload Handling**: Memory-efficient upload processing
- **Flexible Routing**: Catch-all routes and parameterized endpoints
- **Built-in Monitoring**: Automatic statistics logging and endpoint

## 📊 Endpoints

### Core Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Main dashboard with upload form |
| `/html` | ANY | Simple HTML response |
| `/genfile?size=N` | ANY | Generate file with specified size |
| `/upload` | POST | File upload endpoint |
| `/stats` | ANY | Request statistics |
| `/echo/{code}` | GET | Echo with custom HTTP status code |
| `/path/{any:*}` | GET | Catch-all route for path testing |

### Static Content

| Endpoint | Description |
|----------|-------------|
| `/style.css` | Sample CSS file |
| `/app.js` | Sample JavaScript file |
| `/img/{name}` | Dynamic image generation |
| `/dir/` | Directory listing |
| `/dir/{filename}` | Dynamic file serving |

## 🛠️ Installation

### Prerequisites
- Go 1.16 or higher
- Git

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd hptest

# Build the application
go build -o hptest main.go

# Run the server
./hptest