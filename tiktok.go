package ttscrape_go

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// TikTokSession represents a browser session for TikTok
type TikTokSession struct {
	Context    context.Context
	CancelFunc context.CancelFunc
	MsToken    string
	Proxy      string
	Headers    map[string]string
	Params     map[string]string
	BaseURL    string
	BrowserFree bool // Flag to indicate if this session operates without a browser
}

// TikTokAPI is the main API client for TikTok
type TikTokAPI struct {
	Sessions []*TikTokSession
	Logger   *log.Logger
	Headless bool // Flag to indicate if browser should run in headless mode
	BrowserFree bool // Flag to indicate if we should try to operate without a browser after initial setup
}

// NewTikTokAPI creates a new TikTok API client
func NewTikTokAPI(logLevel int) *TikTokAPI {
	logger := log.New(os.Stdout, "TikTokAPI: ", log.LstdFlags)
	return &TikTokAPI{
		Sessions: make([]*TikTokSession, 0),
		Logger:   logger,
		Headless: true, // Default to headless mode for better performance
		BrowserFree: false, // Default to using browser for compatibility
	}
}

// SetHeadless sets whether to use headless mode for browser sessions
func (api *TikTokAPI) SetHeadless(headless bool) {
	api.Headless = headless
}

// SetBrowserFree sets whether to operate without a browser after initial setup
func (api *TikTokAPI) SetBrowserFree(browserFree bool) {
	api.BrowserFree = browserFree
}

// CreateSessions creates browser sessions for TikTok
func (api *TikTokAPI) CreateSessions(ctx context.Context, numSessions int, msTokens []string, sleepAfter int, browser string) error {
	// If we have valid msTokens and BrowserFree is enabled, create browser-free sessions
	if api.BrowserFree && len(msTokens) >= numSessions {
		for i := 0; i < numSessions; i++ {
			if msTokens[i] == "" {
				continue // Skip empty tokens
			}
			
			session := &TikTokSession{
				MsToken:    msTokens[i],
				Headers:    createDefaultHeaders(),
				Params:     createDefaultParams(msTokens[i]),
				BaseURL:    "https://www.tiktok.com",
				BrowserFree: true,
			}
			
			api.Sessions = append(api.Sessions, session)
		}
		
		// If we have all the sessions we need, return early
		if len(api.Sessions) == numSessions {
			return nil
		}
	}
	
	// Otherwise, create browser sessions for the remaining slots
	remainingSessions := numSessions - len(api.Sessions)
	for i := 0; i < remainingSessions; i++ {
		msToken := ""
		if i < len(msTokens) && msTokens[i] != "" {
			msToken = msTokens[i]
		}

		session, err := api.createSession(ctx, "https://www.tiktok.com", msToken, "", sleepAfter)
		if err != nil {
			return err
		}

		api.Sessions = append(api.Sessions, session)
		time.Sleep(time.Duration(sleepAfter) * time.Second)
	}

	return nil
}

// createDefaultHeaders creates a default set of headers for browser-free sessions
func createDefaultHeaders() map[string]string {
	return map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept":          "application/json, text/plain, */*",
		"Referer":         "https://www.tiktok.com/",
		"Origin":          "https://www.tiktok.com",
	}
}

// createDefaultParams creates default parameters for browser-free sessions
func createDefaultParams(msToken string) map[string]string {
	params := map[string]string{
		"aid":                "1988",
		"app_language":       "en",
		"app_name":           "tiktok_web",
		"browser_language":   "en-US",
		"browser_name":       "Mozilla",
		"browser_online":     "true",
		"browser_platform":   "MacIntel",
		"browser_version":    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
		"channel":            "tiktok_web",
		"cookie_enabled":     "true",
		"device_id":          fmt.Sprintf("%d", rand.Int63()),
		"device_platform":    "web",
		"focus_state":        "true",
		"from_page":          "fyp",
		"history_len":        "1",
		"is_fullscreen":      "false",
		"is_page_visible":    "true",
		"language":           "en",
		"os":                 "mac",
		"priority_region":    "",
		"referer":            "",
		"region":             "US",
		"screen_height":      "1080",
		"screen_width":       "1920",
		"tz_name":            "America/New_York",
		"webcast_language":   "en",
	}
	
	if msToken != "" {
		params["msToken"] = msToken
	}
	
	return params
}

// createSession creates a single browser session
func (api *TikTokAPI) createSession(ctx context.Context, startURL string, msToken string, proxy string, sleepAfter int) (*TikTokSession, error) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"),
	}
	
	// Add headless option if enabled
	if api.Headless {
		opts = append(opts, chromedp.Headless)
	}

	if proxy != "" {
		opts = append(opts, chromedp.ProxyServer(proxy))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	browserCtx, _ := chromedp.NewContext(allocCtx)

	session := &TikTokSession{
		Context:    browserCtx,
		CancelFunc: cancel,
		MsToken:    msToken,
		Proxy:      proxy,
		Headers:    make(map[string]string),
		Params:     make(map[string]string),
		BaseURL:    "https://www.tiktok.com",
		BrowserFree: false,
	}

	// Navigate to TikTok
	err := chromedp.Run(browserCtx, chromedp.Navigate(startURL))
	if err != nil {
		cancel()
		return nil, err
	}

	// Set up session parameters
	err = api.setSessionParams(session)
	if err != nil {
		cancel()
		return nil, err
	}

	time.Sleep(time.Duration(sleepAfter) * time.Second)
	
	// If browser-free mode is enabled and we have a valid msToken, convert this to a browser-free session
	if api.BrowserFree && session.MsToken != "" {
		// Save the important data
		msToken := session.MsToken
		headers := session.Headers
		params := session.Params
		
		// Close the browser
		session.CancelFunc()
		
		// Create a browser-free session with the same data
		return &TikTokSession{
			MsToken:    msToken,
			Headers:    headers,
			Params:     params,
			BaseURL:    "https://www.tiktok.com",
			BrowserFree: true,
		}, nil
	}
	
	return session, nil
}

// setSessionParams sets the session parameters
func (api *TikTokAPI) setSessionParams(session *TikTokSession) error {
	var userAgent, language string

	err := chromedp.Run(session.Context,
		chromedp.Evaluate(`navigator.userAgent`, &userAgent),
		chromedp.Evaluate(`navigator.language || navigator.userLanguage`, &language),
	)
	if err != nil {
		return err
	}

	session.Headers = map[string]string{
		"User-Agent":      userAgent,
		"Accept-Language": language,
		"Accept":          "application/json, text/plain, */*",
		"Referer":         "https://www.tiktok.com/",
		"Origin":          "https://www.tiktok.com",
	}

	session.Params = map[string]string{
		"aid":                "1988",
		"app_language":       language,
		"app_name":           "tiktok_web",
		"browser_language":   language,
		"browser_name":       "Mozilla",
		"browser_online":     "true",
		"browser_platform":   "MacIntel",
		"browser_version":    userAgent,
		"channel":            "tiktok_web",
		"cookie_enabled":     "true",
		"device_id":          fmt.Sprintf("%d", rand.Int63()),
		"device_platform":    "web",
		"focus_state":        "true",
		"from_page":          "fyp",
		"history_len":        "1",
		"is_fullscreen":      "false",
		"is_page_visible":    "true",
		"language":           language,
		"os":                 "mac",
		"priority_region":    "",
		"referer":            "",
		"region":             "US",
		"screen_height":      "1080",
		"screen_width":       "1920",
		"tz_name":            "America/New_York",
		"webcast_language":   language,
	}

	if session.MsToken != "" {
		session.Params["msToken"] = session.MsToken
	}

	return nil
}

// MakeRequest makes an HTTP request to TikTok
func (api *TikTokAPI) MakeRequest(urlStr string, params map[string]string, headers map[string]string, sessionIndex int) (map[string]interface{}, error) {
	if sessionIndex >= len(api.Sessions) {
		return nil, fmt.Errorf("session index out of range")
	}

	session := api.Sessions[sessionIndex]
	
	// For browser-free sessions, just use HTTP client
	if session.BrowserFree {
		return api.makeHTTPRequest(session, urlStr, params, headers)
	}
	
	// Merge params
	mergedParams := make(map[string]string)
	for k, v := range session.Params {
		mergedParams[k] = v
	}
	for k, v := range params {
		mergedParams[k] = v
	}

	// Build URL with params
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	q := parsedURL.Query()
	for k, v := range mergedParams {
		q.Set(k, v)
	}
	parsedURL.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range session.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// makeHTTPRequest makes a direct HTTP request without using a browser
func (api *TikTokAPI) makeHTTPRequest(session *TikTokSession, urlStr string, params map[string]string, headers map[string]string) (map[string]interface{}, error) {
	// Merge params
	mergedParams := make(map[string]string)
	for k, v := range session.Params {
		mergedParams[k] = v
	}
	for k, v := range params {
		mergedParams[k] = v
	}

	// Build URL with params
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	q := parsedURL.Query()
	for k, v := range mergedParams {
		q.Set(k, v)
	}
	parsedURL.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range session.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Close closes all sessions
func (api *TikTokAPI) Close() {
	for _, session := range api.Sessions {
		if !session.BrowserFree && session.CancelFunc != nil {
			session.CancelFunc()
		}
	}
	api.Sessions = nil
}

// Sound returns a new Sound object
func (api *TikTokAPI) Sound(id string) *Sound {
	return &Sound{
		API: api,
		ID:  id,
	}
} 