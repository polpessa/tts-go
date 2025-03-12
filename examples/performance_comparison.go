package examples

import (
	"context"
	"fmt"
	"os"
	"time"

	ttscrape_go "github.com/fortindustries/ttscrape-go"
)

// PerformanceComparison demonstrates the performance differences between different modes
func PerformanceComparison() {
	// Get ms_token from environment variable or use the provided one
	msToken := os.Getenv("ms_token")
	if msToken == "" {
		msToken = "fUpfvLbm0fvAb2-0k1pMGCA4DJDnHNm3ouzvhDE528p9DFZ3i0KxQqs-kAs8ebd2mEpwEH2haKR-6mGiILoOezdG6-H4J4WRA-YwlR05jUvhBRyIF8oTr0I9uHrrJFoa7UHqgQ=="
		fmt.Println("Using default ms_token")
	}

	// Sound ID to fetch
	soundID := "7016547803243022337"

	// Run tests for different configurations
	fmt.Println("=== Performance Comparison ===")
	
	// Test 1: Regular browser mode
	fmt.Println("\n=== Test 1: Regular Browser Mode ===")
	runTest(soundID, msToken, false, false)
	
	// Test 2: Headless browser mode
	fmt.Println("\n=== Test 2: Headless Browser Mode ===")
	runTest(soundID, msToken, true, false)
	
	// Test 3: Browser-free mode
	fmt.Println("\n=== Test 3: Browser-Free Mode ===")
	runTest(soundID, msToken, true, true)
	
	fmt.Println("\n=== Performance Comparison Complete ===")
}

func runTest(soundID, msToken string, headless, browserFree bool) {
	startTotal := time.Now()
	
	// Create a new TikTok API client
	fmt.Println("Creating TikTok API client...")
	api := ttscrape_go.NewTikTokAPI(0)
	
	// Configure the API client
	api.SetHeadless(headless)
	api.SetBrowserFree(browserFree)
	
	fmt.Printf("Configuration: Headless=%v, BrowserFree=%v\n", headless, browserFree)

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
	sessionTime := time.Since(startSession)
	fmt.Printf("Sessions created in %v\n", sessionTime)
	defer api.Close()

	// Get sound
	sound := api.Sound(soundID)

	// Get sound info
	fmt.Println("Getting sound info...")
	startInfo := time.Now()
	_, err = sound.Info(map[string]interface{}{
		"session_index": 0,
	})
	if err != nil {
		fmt.Printf("Error getting sound info: %v\n", err)
		return
	}
	infoTime := time.Since(startInfo)
	fmt.Printf("Sound info retrieved in %v\n", infoTime)

	// Get videos
	fmt.Println("Getting videos...")
	startVideos := time.Now()
	videos, err := sound.Videos(5, 0, map[string]interface{}{
		"session_index": 0,
	})
	if err != nil {
		fmt.Printf("Error getting videos: %v\n", err)
		return
	}

	// Count videos
	count := 0
	var firstVideoTime time.Duration
	for range videos {
		if count == 0 {
			firstVideoTime = time.Since(startVideos)
			fmt.Printf("First video received in %v\n", firstVideoTime)
		}
		count++
	}
	videosTime := time.Since(startVideos)
	fmt.Printf("All %d videos received in %v\n", count, videosTime)

	// Print total time
	totalTime := time.Since(startTotal)
	fmt.Printf("Total execution time: %v\n", totalTime)
	
	// Print summary
	fmt.Println("\n=== Performance Summary ===")
	fmt.Printf("Session creation: %v\n", sessionTime)
	fmt.Printf("Sound info retrieval: %v\n", infoTime)
	fmt.Printf("First video retrieval: %v\n", firstVideoTime)
	fmt.Printf("All videos retrieval: %v\n", videosTime)
	fmt.Printf("Total execution: %v\n", totalTime)
} 