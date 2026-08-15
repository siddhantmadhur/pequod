# Ocelot Media Server

A free, open-source, self-hosted media server built from the ground up in Go — designed to be fast, beautiful, and extensible.

> **Status:** Early development. See the [roadmap](#roadmap) below for current progress.

## Why Ocelot?

Ocelot aims to be a modern alternative to Plex and Jellyfin, built without legacy tech debt. The goal is full feature parity with existing solutions while prioritizing a polished user experience and a clean, performant backend written entirely in Go.

## Features

- **Adaptive HLS Streaming** — On-the-fly transcoding via ffmpeg with dynamic seeking, segment tracking, and session management
- **Direct Play** — Serve original files without re-encoding when the client supports the format
- **Library Management** — Register directories, scan for media (movies & TV series), and organize content hierarchically
- **Metadata Matching** — Automatic metadata, episode info, and poster retrieval from [TMDB](https://www.themoviedb.org/)
- **Authentication** — JWT access/refresh token pairs with bcrypt password hashing and per-device session tracking
- **Setup Wizard** — Guided first-run flow for admin account creation and server configuration
- **Docker Support** — Single-image deployment

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22 |
| HTTP | [Echo](https://github.com/labstack/echo) |
| Database | SQLite ([go-sqlite3](https://github.com/mattn/go-sqlite3)) |
| Query Gen | [sqlc](https://sqlc.dev/) |
| Auth | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt |
| Transcoding | ffmpeg / ffprobe (HLS) |
| Config | TOML ([BurntSushi/toml](https://github.com/BurntSushi/toml)) |
| Metadata | [TMDB API v3](https://developer.themoviedb.org/docs) |

## Getting Started

### Prerequisites

- Go 1.22+
- C compiler (required by go-sqlite3 CGO bindings)
- ffmpeg & ffprobe
- A [TMDB Read Access Token](https://developer.themoviedb.org/docs/getting-started)

### Configuration

On first run, Ocelot generates a `config.toml` with sensible defaults (port `8080`, auto-generated JWT secret key). The persistent data directory can be set via environment variable:

```sh
export PERSISTENT_DATA=/path/to/data    # defaults to /data
export TMDB_READ_TOKEN=your_token       # required for metadata
```

### Run Locally

```sh
go build -o ocelot .
./ocelot
```

### Run with Docker

```sh
docker build -t ocelot .
docker run -e PERSISTENT_DATA=/data -e TMDB_READ_TOKEN=your_token -p 8080:8080 ocelot
```

> **Note:** The container requires ffmpeg and ffprobe installed for transcoding. Ensure the base image or a custom image includes them.

On first launch, navigate to the setup wizard to create your admin account.

## API Overview

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/wizard/get-first-user` | Check first-run setup status |
| `GET` | `/server/information` | Server info (hostname, OS, version) |
| `GET` | `/server/information/folders` | Browse filesystem for library setup |
| `POST` | `/auth/create/user` | Create a user account |
| `POST` | `/auth/login` | Authenticate and receive token pair |
| `POST` | `/auth/refresh` | Refresh an expired access token |
| `GET` | `/auth/user` | Get current user profile |
| `GET` | `/server/media/library` | List all libraries |
| `POST` | `/server/media/library` | Add a new library (triggers scan) |
| `GET` | `/media/library/:type/content` | Get content by library and media type |
| `GET` | `/media/library/:id/children` | Get child items (e.g. episodes) |
| `GET` | `/media/:id/playback/info` | Get playback info / start session |
| `GET` | `/media/:id/streams/:session/master.m3u8` | HLS master playlist |
| `GET` | `/media/:id/streams/:session/:segment/stream.ts` | HLS segment |
| `GET` | `/media/:id/direct/:fileName` | Direct file playback |

Authenticated endpoints require a valid `Bearer` token in the `Authorization` header.

## Roadmap

- [x] JWT auth with access/refresh tokens & session tracking
- [x] Admin user creation via setup wizard
- [x] HLS transcoding with dynamic seeking & segment management
- [x] Direct play support
- [x] Library scanning (movies & TV series)
- [x] Season/episode parsing and metadata matching via TMDB
- [ ] Secondary users with granular permissions
- [ ] Subtitle support
- [ ] Thumbnail generation
- [ ] Live directory change detection
- [ ] Per-user library access control
- [ ] Plugin & theming system
- [ ] Activity logging & analytics

## Planned Client


A cross-platform client is planned using Electron + React (desktop/web) and React Native (mobile), targeting **iOS, Android, macOS, Windows, Linux, and WebOS**.
As an alternative (while testing) the project should be able to be used via terminal commands like yt-dlp. More info on this coming. 


## Contributing

Bug reports and feature requests are welcome via [GitHub Issues](../../issues) — templates are provided for both.

## License

[GPL-3.0](LICENSE)

---

Created by [Siddhant Madhur](https://github.com/siddhantmadhur) — development started April 29, 2024.
