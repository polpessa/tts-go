package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	ttscrape_go "github.com/fortindustries/ttscrape-go"
)

// SoundResult represents the result of a sound scraping operation
type SoundResult struct {
	SoundID string
	Info    map[string]interface{}
	Error   error
	Time    time.Duration
}

// ConcurrentSoundScraper demonstrates how to scrape multiple sounds concurrently
func ConcurrentSoundScraper(soundIDs []string, msToken string) {
	if msToken == "" {
		msToken = os.Getenv("ms_token")
		if msToken == "" {
			msToken = "fUpfvLbm0fvAb2-0k1pMGCA4DJDnHNm3ouzvhDE528p9DFZ3i0KxQqs-kAs8ebd2mEpwEH2haKR-6mGiILoOezdG6-H4J4WRA-YwlR05jUvhBRyIF8oTr0I9uHrrJFoa7UHqgQ=="
			fmt.Println("Using default ms_token")
		}
	}

	// If no sound IDs are provided, use these default ones
	if len(soundIDs) == 0 {
		soundIDs = []string{
			"6812253843712346882",
			"6770603781966039810",
			"7446022050302675728",
			"7450943781185538821",
			"7477555291099122449",
			"7277237345823230725",
			"7480249266632510225",
			"7479861742403619600",
			"7479738622171286289",
			"7479471349645462278",
		}
	}

	// Limit to 10 sounds if more are provided
	if len(soundIDs) > 10 {
		soundIDs = soundIDs[:10]
	}

	fmt.Printf("Scraping %d sounds concurrently...\n", len(soundIDs))
	startTotal := time.Now()

	// Create a new TikTok API client with optimized settings
	api := ttscrape_go.NewTikTokAPI(0)
	api.SetHeadless(true)
	api.SetBrowserFree(true)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create a single session
	fmt.Println("Creating session...")
	startSession := time.Now()
	err := api.CreateSessions(ctx, 1, []string{msToken}, 3, "chromium")
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}
	fmt.Printf("Session created in %v\n", time.Since(startSession))
	defer api.Close()

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	resultChan := make(chan SoundResult, len(soundIDs))

	// Start scraping each sound in a separate goroutine
	for _, soundID := range soundIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			startSound := time.Now()
			
			// Get sound info
			sound := api.Sound(id)
			info, err := sound.Info(map[string]interface{}{
				"session_index": 0,
			})
			
			// Send result to channel
			resultChan <- SoundResult{
				SoundID: id,
				Info:    info,
				Error:   err,
				Time:    time.Since(startSound),
			}
		}(soundID)
	}

	// Close the result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect and print results
	fmt.Println("\nResults:")
	fmt.Println("========")
	
	successCount := 0
	failCount := 0
	var totalTime time.Duration
	
	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("Sound %s: ERROR - %v (took %v)\n", result.SoundID, result.Error, result.Time)
			failCount++
		} else {
			// Convert to JSON for display
			infoJSON, _ := json.MarshalIndent(result.Info, "", "  ")
			fmt.Printf("Sound %s: SUCCESS (took %v)\n", result.SoundID, result.Time)
			
			// Print a small excerpt of the JSON to verify it's valid
			if len(infoJSON) > 200 {
				fmt.Printf("  JSON excerpt: %s...\n", infoJSON[:200])
			} else {
				fmt.Printf("  JSON excerpt: %s\n", infoJSON)
			}
			
			successCount++
		}
		totalTime += result.Time
	}

	// Print summary
	totalExecutionTime := time.Since(startTotal)
	fmt.Println("\nSummary:")
	fmt.Println("========")
	fmt.Printf("Total sounds: %d\n", len(soundIDs))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failCount)
	fmt.Printf("Average time per sound: %v\n", totalTime/time.Duration(len(soundIDs)))
	fmt.Printf("Total execution time: %v\n", totalExecutionTime)
	fmt.Printf("Effective concurrency speedup: %.2fx\n", float64(totalTime)/float64(totalExecutionTime))
}

// RunConcurrentSoundScraper is a main function that can be used to run the example
func RunConcurrentSoundScraper() {
	// Read sound IDs from the file
	soundIDs, err := readSoundIDsFromFile("support_files/sound_ids.csv")
	if err != nil {
		fmt.Printf("Error reading sound IDs: %v\n", err)
		// Use default sound IDs
		ConcurrentSoundScraper(nil, "")
		return
	}
	
	// Run the scraper with the first 10 sound IDs
	if len(soundIDs) > 10 {
		soundIDs = soundIDs[:10]
	}
	ConcurrentSoundScraper(soundIDs, "")
}

// readSoundIDsFromFile reads sound IDs from a CSV file
func readSoundIDsFromFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	// Split by newlines and filter out empty lines
	lines := make([]string, 0)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	
	return lines, nil
} 