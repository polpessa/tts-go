package ttscrape_go

import (
	"fmt"
	"sync"
)

// Sound represents a TikTok sound/music/song
type Sound struct {
	API      interface{} // Reference to the TikTokAPI
	ID       string
	Title    string
	Duration int
	Original bool
	AsDict   map[string]interface{}
	mu       sync.Mutex
}

// Info retrieves information about the sound
func (s *Sound) Info(options map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the API reference
	api, ok := s.API.(interface {
		MakeRequest(string, map[string]string, map[string]string, int) (map[string]interface{}, error)
	})
	if !ok {
		return nil, fmt.Errorf("invalid API reference")
	}

	// Set up URL parameters
	params := map[string]string{
		"musicId": s.ID,
	}

	// Get session index
	sessionIndex := 0
	if val, ok := options["session_index"]; ok {
		if idx, ok := val.(int); ok {
			sessionIndex = idx
		}
	}

	// Get ms_token
	if val, ok := options["ms_token"]; ok {
		if token, ok := val.(string); ok {
			params["msToken"] = token
		}
	}

	// Get headers
	headers := map[string]string{}
	if val, ok := options["headers"]; ok {
		if hdrs, ok := val.(map[string]string); ok {
			headers = hdrs
		}
	}

	// Make the request
	resp, err := api.MakeRequest(
		"https://www.tiktok.com/api/music/detail/",
		params,
		headers,
		sessionIndex,
	)
	if err != nil {
		return nil, err
	}

	// Check if response is valid
	if resp == nil {
		return nil, fmt.Errorf("TikTok returned an invalid response")
	}

	// Extract data
	s.AsDict = resp
	s.extractFromData()

	return resp, nil
}

// Videos retrieves videos that use this sound
func (s *Sound) Videos(count int, cursor int, options map[string]interface{}) (chan map[string]interface{}, error) {
	// Get the API reference
	api, ok := s.API.(interface {
		MakeRequest(string, map[string]string, map[string]string, int) (map[string]interface{}, error)
	})
	if !ok {
		return nil, fmt.Errorf("invalid API reference")
	}

	// Create a channel to send videos
	videos := make(chan map[string]interface{}, count)

	// Start a goroutine to fetch videos
	go func() {
		defer close(videos)

		currentCount := 0
		currentCursor := cursor

		for currentCount < count {
			// Set up URL parameters
			params := map[string]string{
				"musicID": s.ID,
				"count":   fmt.Sprintf("%d", 30), // Max count per request
				"cursor":  fmt.Sprintf("%d", currentCursor),
			}

			// Get session index
			sessionIndex := 0
			if val, ok := options["session_index"]; ok {
				if idx, ok := val.(int); ok {
					sessionIndex = idx
				}
			}

			// Get ms_token
			if val, ok := options["ms_token"]; ok {
				if token, ok := val.(string); ok {
					params["msToken"] = token
				}
			}

			// Get headers
			headers := map[string]string{}
			if val, ok := options["headers"]; ok {
				if hdrs, ok := val.(map[string]string); ok {
					headers = hdrs
				}
			}

			// Make the request
			resp, err := api.MakeRequest(
				"https://www.tiktok.com/api/music/item_list/",
				params,
				headers,
				sessionIndex,
			)
			if err != nil {
				return
			}

			// Check if response is valid
			if resp == nil {
				return
			}

			// Extract videos
			itemList, ok := resp["itemList"].([]interface{})
			if !ok {
				return
			}

			// No more videos
			if len(itemList) == 0 {
				return
			}

			// Send videos to channel
			for _, item := range itemList {
				if currentCount >= count {
					return
				}

				videoMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				videos <- videoMap
				currentCount++
			}

			// Update cursor for next page
			hasMore, ok := resp["hasMore"].(bool)
			if !ok || !hasMore {
				return
			}

			cursor, ok := resp["cursor"].(float64)
			if !ok {
				return
			}

			currentCursor = int(cursor)
		}
	}()

	return videos, nil
}

// extractFromData extracts data from the API response
func (s *Sound) extractFromData() {
	if s.AsDict == nil {
		return
	}

	// Extract music info
	musicInfo, ok := s.AsDict["musicInfo"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract title
	if title, ok := musicInfo["title"].(string); ok {
		s.Title = title
	}

	// Extract duration
	if duration, ok := musicInfo["duration"].(float64); ok {
		s.Duration = int(duration)
	}

	// Extract original
	if original, ok := musicInfo["original"].(bool); ok {
		s.Original = original
	}
} 