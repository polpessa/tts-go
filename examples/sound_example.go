package examples

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	ttscrape_go "github.com/fortindustries/ttscrape-go"
)

// SoundExample demonstrates how to use the TikTok API to fetch sound information
func SoundExample() {
	// Parse command line flags
	headless := flag.Bool("headless", true, "Run in headless mode")
	browserFree := flag.Bool("browser-free", true, "Use browser-free mode after initial setup")
	flag.Parse()

	// Start timing the entire process
	startTotal := time.Now()

	// Get ms_token from environment variable or use the provided one
	msToken := os.Getenv("ms_token")
	if msToken == "" {
		msToken = "fUpfvLbm0fvAb2-0k1pMGCA4DJDnHNm3ouzvhDE528p9DFZ3i0KxQqs-kAs8ebd2mEpwEH2haKR-6mGiILoOezdG6-H4J4WRA-YwlR05jUvhBRyIF8oTr0I9uHrrJFoa7UHqgQ=="
		fmt.Println("Using default ms_token")
	}

	// Sound ID to fetch
	soundID := "7016547803243022337"

	// Create a new TikTok API client
	fmt.Println("Creating TikTok API client...")
	startInit := time.Now()
	api := ttscrape_go.NewTikTokAPI(0)
	
	// Configure the API client
	api.SetHeadless(*headless)
	api.SetBrowserFree(*browserFree)
	
	fmt.Printf("API client created in %v\n", time.Since(startInit))
	fmt.Printf("Headless mode: %v\n", *headless)
	fmt.Printf("Browser-free mode: %v\n", *browserFree)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create sessions
	fmt.Println("Creating sessions...")
	startSession := time.Now()
	err := api.CreateSessions(ctx, 1, []string{msToken}, 3, "chromium")
	if err != nil {
		fmt.Printf("Error creating sessions: %v\n", err)
		return
	}
	fmt.Printf("Sessions created in %v\n", time.Since(startSession))
	defer api.Close()

	// Get sound
	fmt.Println("Getting sound...")
	sound := api.Sound(soundID)

	// Get sound info
	fmt.Println("Getting sound info...")
	startInfo := time.Now()
	info, err := sound.Info(map[string]interface{}{
		"session_index": 0,
	})
	if err != nil {
		fmt.Printf("Error getting sound info: %v\n", err)
		return
	}
	fmt.Printf("Sound info retrieved in %v\n", time.Since(startInfo))

	// Print sound info
	infoJSON, _ := json.MarshalIndent(info, "", "  ")
	fmt.Printf("Sound Info: %s\n", infoJSON)

	// Get videos
	fmt.Println("Getting videos...")
	startVideos := time.Now()
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
		if count == 0 {
			// Print time to first result
			fmt.Printf("First video received in %v\n", time.Since(startVideos))
		}
		videoJSON, _ := json.MarshalIndent(video, "", "  ")
		fmt.Printf("Video %d: %s\n", count+1, videoJSON)
		count++
	}
	fmt.Printf("All %d videos received in %v\n", count, time.Since(startVideos))

	// Print total time
	fmt.Printf("Total execution time: %v\n", time.Since(startTotal))
} 