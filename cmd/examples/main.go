package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
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
	Videos  []map[string]interface{}
}

func main() {
	// Define command line flags
	exampleType := flag.String("example", "concurrent", "Type of example to run (sound, performance, concurrent)")
	flag.Parse()

	// Check if ms_token is set
	msToken := os.Getenv("ms_token")
	if msToken == "" {
		msToken = "fUpfvLbm0fvAb2-0k1pMGCA4DJDnHNm3ouzvhDE528p9DFZ3i0KxQqs-kAs8ebd2mEpwEH2haKR-6mGiILoOezdG6-H4J4WRA-YwlR05jUvhBRyIF8oTr0I9uHrrJFoa7UHqgQ=="
		fmt.Println("Warning: ms_token environment variable not set. Using default token.")
	}

	// Run the selected example
	switch *exampleType {
	case "sound":
		fmt.Println("Running Sound Example...")
		runSoundExample(msToken)
	case "performance":
		fmt.Println("Running Performance Comparison Example...")
		runPerformanceComparison(msToken)
	case "concurrent":
		fmt.Println("Running Concurrent Sound Scraper Example...")
		runConcurrentSoundScraper(msToken)
	default:
		fmt.Printf("Unknown example type: %s\n", *exampleType)
		fmt.Println("Available examples: sound, performance, concurrent")
		os.Exit(1)
	}
}

// runSoundExample demonstrates how to use the TikTok API to fetch sound information
func runSoundExample(msToken string) {
	startTotal := time.Now()

	// Create a new TikTok API client
	fmt.Println("Creating TikTok API client...")
	startClient := time.Now()
	api := ttscrape_go.NewTikTokAPI(0)
	api.SetHeadless(true)
	api.SetBrowserFree(true)
	fmt.Printf("TikTok API client created in %v\n", time.Since(startClient))

	// Create a session
	fmt.Println("Creating session...")
	startSession := time.Now()
	ctx := context.Background()
	err := api.CreateSessions(ctx, 1, []string{msToken}, 3, "chromium")
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}
	fmt.Printf("Session created in %v\n", time.Since(startSession))
	defer api.Close()

	// Get sound info
	fmt.Println("Getting sound info...")
	startSound := time.Now()
	soundID := "7277237345823230725" // Example sound ID
	sound := api.Sound(soundID)
	info, err := sound.Info(map[string]interface{}{
		"session_index": 0,
	})
	if err != nil {
		fmt.Printf("Error getting sound info: %v\n", err)
		return
	}
	fmt.Printf("Sound info retrieved in %v\n", time.Since(startSound))

	// Print sound info
	fmt.Println("\nSound Info:")
	fmt.Println("==========")
	infoJSON, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(infoJSON))

	// Get videos
	fmt.Println("\nGetting videos...")
	startVideos := time.Now()
	videos, err := sound.Videos(5, 0, map[string]interface{}{
		"session_index": 0,
	})
	if err != nil {
		fmt.Printf("Error getting videos: %v\n", err)
		return
	}
	fmt.Printf("Videos retrieved in %v\n", time.Since(startVideos))

	// Print videos
	fmt.Println("\nVideos:")
	fmt.Println("=======")
	videoCount := 0
	for video := range videos {
		videoCount++
		fmt.Printf("Video %d:\n", videoCount)
		videoJSON, _ := json.MarshalIndent(video, "", "  ")
		fmt.Println(string(videoJSON))
	}

	fmt.Printf("\nTotal execution time: %v\n", time.Since(startTotal))
}

// runPerformanceComparison demonstrates the performance differences between different modes
func runPerformanceComparison(msToken string) {
	soundID := "7277237345823230725" // Example sound ID
	
	// Run with regular browser mode
	fmt.Println("\n1. Regular Browser Mode:")
	fmt.Println("=======================")
	startRegular := time.Now()
	
	api1 := ttscrape_go.NewTikTokAPI(0)
	api1.SetHeadless(false)
	api1.SetBrowserFree(false)
	
	ctx1 := context.Background()
	err1 := api1.CreateSessions(ctx1, 1, []string{msToken}, 3, "chromium")
	if err1 != nil {
		fmt.Printf("Error creating session: %v\n", err1)
		return
	}
	
	sound1 := api1.Sound(soundID)
	info1, err1 := sound1.Info(map[string]interface{}{
		"session_index": 0,
	})
	if err1 != nil {
		fmt.Printf("Error getting sound info: %v\n", err1)
	} else {
		fmt.Printf("Sound info retrieved successfully. Title: %v\n", info1["title"])
	}
	
	regularTime := time.Since(startRegular)
	fmt.Printf("Total time: %v\n", regularTime)
	api1.Close()
	
	// Run with headless browser mode
	fmt.Println("\n2. Headless Browser Mode:")
	fmt.Println("========================")
	startHeadless := time.Now()
	
	api2 := ttscrape_go.NewTikTokAPI(0)
	api2.SetHeadless(true)
	api2.SetBrowserFree(false)
	
	ctx2 := context.Background()
	err2 := api2.CreateSessions(ctx2, 1, []string{msToken}, 3, "chromium")
	if err2 != nil {
		fmt.Printf("Error creating session: %v\n", err2)
		return
	}
	
	sound2 := api2.Sound(soundID)
	info2, err2 := sound2.Info(map[string]interface{}{
		"session_index": 0,
	})
	if err2 != nil {
		fmt.Printf("Error getting sound info: %v\n", err2)
	} else {
		fmt.Printf("Sound info retrieved successfully. Title: %v\n", info2["title"])
	}
	
	headlessTime := time.Since(startHeadless)
	fmt.Printf("Total time: %v\n", headlessTime)
	api2.Close()
	
	// Run with browser-free mode
	fmt.Println("\n3. Browser-Free Mode:")
	fmt.Println("====================")
	startBrowserFree := time.Now()
	
	api3 := ttscrape_go.NewTikTokAPI(0)
	api3.SetHeadless(true)
	api3.SetBrowserFree(true)
	
	ctx3 := context.Background()
	err3 := api3.CreateSessions(ctx3, 1, []string{msToken}, 3, "chromium")
	if err3 != nil {
		fmt.Printf("Error creating session: %v\n", err3)
		return
	}
	
	sound3 := api3.Sound(soundID)
	info3, err3 := sound3.Info(map[string]interface{}{
		"session_index": 0,
	})
	if err3 != nil {
		fmt.Printf("Error getting sound info: %v\n", err3)
	} else {
		fmt.Printf("Sound info retrieved successfully. Title: %v\n", info3["title"])
	}
	
	browserFreeTime := time.Since(startBrowserFree)
	fmt.Printf("Total time: %v\n", browserFreeTime)
	api3.Close()
	
	// Print performance comparison
	fmt.Println("\nPerformance Comparison:")
	fmt.Println("======================")
	fmt.Printf("Regular Browser Mode:  %v\n", regularTime)
	fmt.Printf("Headless Browser Mode: %v (%.1f%% faster than regular)\n", 
		headlessTime, 100*(float64(regularTime)-float64(headlessTime))/float64(regularTime))
	fmt.Printf("Browser-Free Mode:     %v (%.1f%% faster than regular)\n", 
		browserFreeTime, 100*(float64(regularTime)-float64(browserFreeTime))/float64(regularTime))
}

// runConcurrentSoundScraper demonstrates how to scrape multiple sounds concurrently
func runConcurrentSoundScraper(msToken string) {
	// Read sound IDs from the file
	soundIDs, err := readSoundIDsFromFile("../../support_files/sound_ids.csv")
	if err != nil {
		fmt.Printf("Error reading sound IDs: %v\n", err)
		// Use default sound IDs
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
	err = api.CreateSessions(ctx, 1, []string{msToken}, 3, "chromium")
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
			
			// Get videos for this sound (limit to 5 videos per sound)
			var videos []map[string]interface{}
			if err == nil {
				videoChan, videoErr := sound.Videos(5, 0, map[string]interface{}{
					"session_index": 0,
				})
				
				if videoErr == nil {
					// Collect videos from channel
					for video := range videoChan {
						videos = append(videos, video)
					}
				}
			}
			
			// Send result to channel
			resultChan <- SoundResult{
				SoundID: id,
				Info:    info,
				Error:   err,
				Time:    time.Since(startSound),
				Videos:  videos,
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
	
	// Create a map to store all results for JSON output
	resultsMap := make(map[string]interface{})
	
	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("Sound %s: ERROR - %v (took %v)\n", result.SoundID, result.Error, result.Time)
			resultsMap[result.SoundID] = map[string]interface{}{
				"error": result.Error.Error(),
				"time_ms": result.Time.Milliseconds(),
			}
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
			
			// Print video count
			fmt.Printf("  Videos found: %d\n", len(result.Videos))
			
			// Store the result in the map
			resultsMap[result.SoundID] = map[string]interface{}{
				"info": result.Info,
				"videos": result.Videos,
				"video_count": len(result.Videos),
				"time_ms": result.Time.Milliseconds(),
			}
			
			successCount++
		}
		totalTime += result.Time
	}

	// Save results to a JSON file
	resultsOutput := map[string]interface{}{
		"results": resultsMap,
		"metadata": map[string]interface{}{
			"total_sounds": len(soundIDs),
			"successful": successCount,
			"failed": failCount,
			"average_time_ms": totalTime.Milliseconds() / int64(len(soundIDs)),
			"total_execution_time_ms": time.Since(startTotal).Milliseconds(),
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}
	
	resultsJSON, err := json.MarshalIndent(resultsOutput, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling results to JSON: %v\n", err)
	} else {
		err = os.WriteFile("results.json", resultsJSON, 0644)
		if err != nil {
			fmt.Printf("Error writing results to file: %v\n", err)
		} else {
			fmt.Println("Results saved to results.json")
		}
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

// readSoundIDsFromFile reads sound IDs from a CSV file
func readSoundIDsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var soundIDs []string
	
	// Try to read as CSV first
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err == nil && len(records) > 0 {
		// Successfully read as CSV
		for _, record := range records {
			if len(record) > 0 && record[0] != "" {
				soundIDs = append(soundIDs, strings.TrimSpace(record[0]))
			}
		}
	} else {
		// If CSV reading fails, try reading as plain text
		file.Seek(0, 0) // Reset file pointer to beginning
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		
		// Split by newlines and filter out empty lines
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				soundIDs = append(soundIDs, line)
			}
		}
	}
	
	return soundIDs, nil
} 