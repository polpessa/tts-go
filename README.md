# TikTok Scraper Go

A Go implementation of the [TikTok-Api](https://github.com/davidteather/TikTok-Api) Python library. This library allows you to scrape data from TikTok, including user information, videos, sounds, and more.

## Features

- Fetch sound information
- Get videos associated with a sound
- Designed for high-performance scraping (millions of requests per day)
- Multiple performance modes:
  - Regular browser mode (visible Chrome window)
  - Headless browser mode (invisible Chrome window)
  - Browser-free mode (no browser needed, up to 95% faster)

## Installation

```bash
go get github.com/fortindustries/ttscrape-go
```

## Usage

### Sound Example

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fortindustries/ttscrape-go"
)

func main() {
	// Get ms_token from environment variable
	msToken := os.Getenv("ms_token")
	if msToken == "" {
		fmt.Println("ms_token environment variable is not set")
		return
	}

	// Sound ID to fetch
	soundID := "7016547803243022337"

	// Create a new TikTok API client
	api := ttscrape_go.NewTikTokAPI(0)

	// Optional: Configure performance options
	api.SetHeadless(true)      // Use headless browser (no visible window)
	api.SetBrowserFree(true)   // Use browser-free mode for maximum performance

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create sessions
	err := api.CreateSessions(ctx, 1, []string{msToken}, 3, "chromium")
	if err != nil {
		fmt.Printf("Error creating sessions: %v\n", err)
		return
	}
	defer api.Close()

	// Get sound
	sound := api.Sound(soundID)

	// Get sound info
	info, err := sound.Info(map[string]interface{}{
		"session_index": 0,
	})
	if err != nil {
		fmt.Printf("Error getting sound info: %v\n", err)
		return
	}

	// Print sound info
	infoJSON, _ := json.MarshalIndent(info, "", "  ")
	fmt.Printf("Sound Info: %s\n", infoJSON)

	// Get videos
	videos, err := sound.Videos(30, 0, map[string]interface{}{
		"session_index": 0,
	})
	if err != nil {
		fmt.Printf("Error getting videos: %v\n", err)
		return
	}

	// Print videos
	count := 0
	for video := range videos {
		videoJSON, _ := json.MarshalIndent(video, "", "  ")
		fmt.Printf("Video %d: %s\n", count+1, videoJSON)
		count++
	}

	fmt.Printf("Found %d videos\n", count)
}
```

## Performance Modes

This library offers three performance modes to suit different needs:

### 1. Regular Browser Mode

Uses a visible Chrome/Chromium browser window. This is useful for debugging but is the slowest option.

```go
api := ttscrape_go.NewTikTokAPI(0)
api.SetHeadless(false)
api.SetBrowserFree(false)
```

### 2. Headless Browser Mode

Uses an invisible Chrome/Chromium browser window. This is faster than regular mode but still requires a browser.

```go
api := ttscrape_go.NewTikTokAPI(0)
api.SetHeadless(true)  // Default is true
api.SetBrowserFree(false)
```

### 3. Browser-Free Mode

The fastest mode that doesn't require a browser after initial setup. This mode is up to 95% faster than regular mode.

```go
api := ttscrape_go.NewTikTokAPI(0)
api.SetBrowserFree(true)
```

#### Performance Comparison

| Mode             | Session Creation | API Requests | Total Time | Improvement |
| ---------------- | ---------------- | ------------ | ---------- | ----------- |
| Regular Browser  | ~10.4s           | ~0.8s        | ~11.2s     | Baseline    |
| Headless Browser | ~7.8s            | ~0.5s        | ~8.3s      | ~26% faster |
| Browser-Free     | ~0.00001s        | ~0.5s        | ~0.5s      | ~95% faster |

## Requirements

- Go 1.16 or higher
- Chrome/Chromium browser installed (for regular and headless modes)
- Valid `ms_token` for authentication

## MS Token

The `ms_token` is required for authentication with TikTok. You can obtain it by:

1. Logging into TikTok in your browser
2. Opening the developer tools (F12)
3. Going to the Application tab
4. Looking for the `ms_token` cookie under the Cookies section

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Based on the [TikTok-Api](https://github.com/davidteather/TikTok-Api) Python library by David Teather
