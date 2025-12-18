package provider

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/service"
	"go.uber.org/zap"
)

// Default base URL - Alibaba Cloud DashScope (Singapore region)
// For Beijing region, use: https://dashscope.aliyuncs.com/compatible-mode/v1
const (
	defaultWanaiBaseURL = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"
)

// WanAIProvider implements the Wan AI video generation provider
type WanAIProvider struct {
	*BaseProvider
	version       string
	serverBaseURL string // Base URL for the API server (for proxy endpoint)
}

// DashScopeGenerateRequest represents a DashScope video generation request
type DashScopeGenerateRequest struct {
	Model      string                    `json:"model"`
	Input      DashScopeInput            `json:"input"`
	Parameters DashScopeGenerationParams `json:"parameters"`
}

// DashScopeInput represents input for video generation
type DashScopeInput struct {
	Prompt         string   `json:"prompt"`
	NegativePrompt string   `json:"negative_prompt,omitempty"`
	ImgURL         string   `json:"img_url,omitempty"`        // For image-to-video
	AudioURL       string   `json:"audio_url,omitempty"`      // For audio sync
	RefImagesURL   []string `json:"ref_images_url,omitempty"` // For multi-image reference
	Function       string   `json:"function,omitempty"`       // For VACE models
}

// DashScopeGenerationParams represents generation parameters
type DashScopeGenerationParams struct {
	Size         string   `json:"size,omitempty"`       // Format: "1280*720" (for text-to-video)
	Resolution   string   `json:"resolution,omitempty"` // Format: "720P", "1080P" (for image-to-video)
	Duration     int      `json:"duration,omitempty"`
	PromptExtend bool     `json:"prompt_extend,omitempty"`
	Watermark    bool     `json:"watermark,omitempty"`
	Audio        bool     `json:"audio,omitempty"`
	AudioURL     string   `json:"audio_url,omitempty"` // Custom audio URL
	Seed         int64    `json:"seed,omitempty"`
	N            int      `json:"n,omitempty"`         // Number of videos (default: 1)
	ShotType     string   `json:"shot_type,omitempty"` // "single" or "multi" (for wan2.6-i2v)
	ObjOrBg      []string `json:"obj_or_bg,omitempty"` // For multi-image reference
}

// DashScopeGenerateResponse represents a DashScope generation response
type DashScopeGenerateResponse struct {
	RequestID string          `json:"request_id,omitempty"`
	Output    DashScopeOutput `json:"output,omitempty"`
	Usage     DashScopeUsage  `json:"usage,omitempty"`
	Code      string          `json:"code,omitempty"`
	Message   string          `json:"message,omitempty"`
}

// DashScopeOutput represents the output from DashScope
type DashScopeOutput struct {
	TaskID     string `json:"task_id,omitempty"`
	TaskStatus string `json:"task_status,omitempty"`
	VideoURL   string `json:"video_url,omitempty"`
	Video      string `json:"video,omitempty"`
	Code       string `json:"code,omitempty"`       // Error code in output
	Message    string `json:"message,omitempty"`    // Error message in output
}

// DashScopeUsage represents API usage information
type DashScopeUsage struct {
	TotalTokens int `json:"total_tokens,omitempty"`
}

// DashScopeTaskResponse represents a task status check response
type DashScopeTaskResponse struct {
	RequestID string          `json:"request_id,omitempty"`
	Output    DashScopeOutput `json:"output,omitempty"`
	Usage     DashScopeUsage  `json:"usage,omitempty"`
	Code      string          `json:"code,omitempty"`
	Message   string          `json:"message,omitempty"`
}

// NewWanAIProvider creates a new Wan AI provider
func NewWanAIProvider(apiKey string, version string, baseURL string, serverBaseURL string, logger *zap.Logger) service.VideoProvider {
	if version == "" {
		version = "2.5" // Default to 2.5 (wan2.6 model not available yet)
	}
	if baseURL == "" {
		baseURL = defaultWanaiBaseURL
	}
	return &WanAIProvider{
		BaseProvider:  NewBaseProvider(apiKey, baseURL, 10*time.Minute, logger),
		version:       version,
		serverBaseURL: serverBaseURL,
	}
}

// GetName returns the provider name
func (p *WanAIProvider) GetName() entity.AIProvider {
	return entity.ProviderWanAI
}

// GenerateVideo initiates video generation with DashScope (Wan AI)
func (p *WanAIProvider) GenerateVideo(ctx context.Context, req service.GenerationRequest) (*entity.GenerationResult, error) {
	// Build the request according to DashScope API format
	// Use image-to-video for better results (always use i2v model)

	// Force 5 seconds for faster generation (override any template defaults)
	duration := 5 // Always use 5 seconds for speed
	// Only allow user to override if they explicitly request longer
	if req.Params.Duration > 0 && req.Params.Duration > 5 {
		// User wants longer, but cap at 15 seconds
		if req.Params.Duration > 15 {
			duration = 15
		} else {
			// Round to nearest valid value (5, 10, or 15)
			if req.Params.Duration <= 7 {
				duration = 5
			} else if req.Params.Duration <= 12 {
				duration = 10
			} else {
				duration = 15
			}
		}
	}

	// Get image URL from template thumbnail, or use default test image
	imgURL := req.ThumbnailURL

	// Handle image URL - proxy ALL external images through our backend to avoid access issues
	if imgURL != "" {
		// Check if it's an external URL (not already proxied)
		if strings.HasPrefix(imgURL, "http://") || strings.HasPrefix(imgURL, "https://") {
			// Check if it's already a local/proxied URL
			if !strings.Contains(imgURL, p.serverBaseURL) && !strings.HasPrefix(imgURL, "/") {
				// External URL - download and cache it locally, then serve from static endpoint
				// This ensures DashScope can access it
				cachedURL, err := p.downloadAndCacheImage(ctx, imgURL, req.TemplateID)
				if err != nil {
					p.logger.Warn("Failed to cache external image, using default",
						zap.String("template_id", req.TemplateID),
						zap.String("thumbnail_url", imgURL),
						zap.Error(err),
					)
					imgURL = "https://cdn.translate.alibaba.com/r/wanx-demo-1.png"
				} else {
					imgURL = cachedURL
					p.logger.Info("Using cached external image",
						zap.String("template_id", req.TemplateID),
						zap.String("original_url", req.ThumbnailURL),
						zap.String("cached_url", imgURL),
					)
				}
			} else {
				// Already proxied or local URL
				p.logger.Info("Using template thumbnail for image-to-video",
					zap.String("template_id", req.TemplateID),
					zap.String("thumbnail_url", imgURL),
				)
			}
		} else {
			// Relative URL, use as-is
			p.logger.Info("Using template thumbnail for image-to-video",
				zap.String("template_id", req.TemplateID),
				zap.String("thumbnail_url", imgURL),
			)
		}
	} else {
		// No template thumbnail, use default test image
		imgURL = "https://cdn.translate.alibaba.com/r/wanx-demo-1.png"
		p.logger.Info("Using default test image (no template thumbnail)",
			zap.String("template_id", req.TemplateID),
		)
	}

	// Use wan2.6-i2v for image-to-video (better quality)
	modelName := "wan2.6-i2v" // Image-to-video model

	// Force 720P for faster generation (override any template defaults)
	// Default to 720P for speed (can use 480P for even faster, 1080P for quality)
	resolution := "720P" // Always default to 720P for speed
	if req.Params.Resolution != "" {
		res := string(req.Params.Resolution)
		switch res {
		case "480p", "480P":
			resolution = "480P" // Fastest option (2-3 minutes)
		case "720p", "720P":
			resolution = "720P" // Fast option (3-5 minutes) - DEFAULT
		case "1080p", "1080P":
			resolution = "1080P" // Slower but highest quality (5-10 minutes)
		default:
			resolution = "720P" // Always default to 720P
		}
	}

	// Use only the user's prompt (template base prompt is ignored)
	finalPrompt := req.Prompt

	dashScopeInput := DashScopeInput{
		Prompt: finalPrompt,
		ImgURL: imgURL,
	}

	dashScopeParams := DashScopeGenerationParams{
		Resolution:   resolution, // Use resolution for i2v models (default 720P for speed)
		Duration:     duration,   // Default 5 seconds for faster generation
		PromptExtend: false,      // Disable auto-extension to use exact user prompt
		Watermark:    false,
		Audio:        true,     // Enable audio for i2v
		ShotType:     "single", // Default to single shot (faster than multi-shot)
	}

	// Build DashScope request
	dashScopeReq := DashScopeGenerateRequest{
		Model:      modelName,
		Input:      dashScopeInput,
		Parameters: dashScopeParams,
	}

	body, err := json.Marshal(dashScopeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// DashScope text-to-video endpoint
	// Extract base URL and construct proper endpoint
	baseURL := p.baseURL
	// Use the video-generation/video-synthesis endpoint for text-to-video
	url := fmt.Sprintf("%s/services/aigc/video-generation/video-synthesis", baseURL)
	// Replace /compatible-mode/v1 with /api/v1 if needed
	if baseURL == "https://dashscope-intl.aliyuncs.com/compatible-mode/v1" {
		url = "https://dashscope-intl.aliyuncs.com/api/v1/services/aigc/video-generation/video-synthesis"
	} else if baseURL == "https://dashscope.aliyuncs.com/compatible-mode/v1" {
		url = "https://dashscope.aliyuncs.com/api/v1/services/aigc/video-generation/video-synthesis"
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	httpReq.Header.Set("X-DashScope-Async", "enable") // Enable async mode

	p.logger.Info("DashScope (Wan AI) API request - Image-to-Video",
		zap.String("url", url),
		zap.String("model", modelName),
		zap.String("version", p.version),
		zap.String("prompt", req.Prompt),
		zap.String("image_url", imgURL),
		zap.String("resolution", resolution),
		zap.Int("duration", duration),
	)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("DashScope API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(bodyBytes)),
		)
		return nil, fmt.Errorf("DashScope API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var dashScopeResp DashScopeGenerateResponse
	if err := json.Unmarshal(bodyBytes, &dashScopeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if dashScopeResp.Code != "" && dashScopeResp.Code != "Success" {
		return nil, fmt.Errorf("DashScope error: %s - %s", dashScopeResp.Code, dashScopeResp.Message)
	}

	if dashScopeResp.Output.TaskID == "" {
		return nil, fmt.Errorf("DashScope response missing task_id")
	}

	p.logger.Info("DashScope video generation started",
		zap.String("task_id", dashScopeResp.Output.TaskID),
		zap.String("status", dashScopeResp.Output.TaskStatus),
	)

	return &entity.GenerationResult{
		ProviderJobID: dashScopeResp.Output.TaskID,
		VideoURL:      dashScopeResp.Output.VideoURL,
		ThumbnailURL:  "",
		Duration:      duration,
	}, nil
}

// GetProgress retrieves generation progress from DashScope
func (p *WanAIProvider) GetProgress(ctx context.Context, providerJobID string) (*entity.Progress, error) {
	// DashScope uses task status endpoint
	baseURL := p.baseURL
	url := fmt.Sprintf("%s/tasks/%s", baseURL, providerJobID)
	// Replace /compatible-mode/v1 with /api/v1 if needed
	if baseURL == "https://dashscope-intl.aliyuncs.com/compatible-mode/v1" {
		url = fmt.Sprintf("https://dashscope-intl.aliyuncs.com/api/v1/tasks/%s", providerJobID)
	} else if baseURL == "https://dashscope.aliyuncs.com/compatible-mode/v1" {
		url = fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/tasks/%s", providerJobID)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("DashScope status check error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(bodyBytes)),
		)
		return nil, fmt.Errorf("DashScope status check error: %d", resp.StatusCode)
	}

	var taskResp DashScopeTaskResponse
	if err := json.Unmarshal(bodyBytes, &taskResp); err != nil {
		p.logger.Error("Failed to decode DashScope response",
			zap.String("body", string(bodyBytes)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log full response for debugging
	p.logger.Info("DashScope task response",
		zap.String("task_id", providerJobID),
		zap.String("code", taskResp.Code),
		zap.String("message", taskResp.Message),
		zap.String("task_status", taskResp.Output.TaskStatus),
		zap.String("video_url", taskResp.Output.VideoURL),
		zap.Any("full_response", string(bodyBytes)),
	)

	// Check for API-level errors first
	if taskResp.Code != "" && taskResp.Code != "Success" {
		// Log the error for debugging
		p.logger.Error("DashScope API error in task status",
			zap.String("task_id", providerJobID),
			zap.String("code", taskResp.Code),
			zap.String("message", taskResp.Message),
		)
		return nil, fmt.Errorf("DashScope error: %s - %s", taskResp.Code, taskResp.Message)
	}

	// Map DashScope task status to our progress
	status := taskResp.Output.TaskStatus
	var progress int
	var stage string
	var message string
	videoURL := taskResp.Output.VideoURL
	if videoURL == "" {
		videoURL = taskResp.Output.Video // Sometimes it's in Video field
	}

	// Log status for debugging
	p.logger.Info("DashScope task status",
		zap.String("task_id", providerJobID),
		zap.String("status", status),
		zap.String("video_url", videoURL),
		zap.String("code", taskResp.Code),
		zap.String("message", taskResp.Message),
		zap.Any("output", taskResp.Output),
	)

	switch status {
	case "SUCCEEDED", "succeeded":
		progress = 100
		stage = "COMPLETED"
		message = "Video generation completed"
		if videoURL == "" {
			p.logger.Warn("DashScope task succeeded but no video URL",
				zap.String("task_id", providerJobID),
			)
		} else {
			p.logger.Info("DashScope task completed successfully",
				zap.String("task_id", providerJobID),
				zap.String("video_url", videoURL),
			)
		}
	case "RUNNING", "running", "PROCESSING", "processing":
		progress = 50
		stage = "PROCESSING"
		message = "Video generation in progress"
	case "PENDING", "pending", "QUEUED", "queued":
		progress = 10
		stage = "PENDING"
		message = "Video generation queued"
	case "FAILED", "failed", "ERROR", "error":
		progress = 0
		stage = "FAILED"
		// Try to extract error message from multiple sources
		errorMsg := taskResp.Message
		// Check output.message first (most common location for DashScope errors)
		if errorMsg == "" && taskResp.Output.Message != "" {
			errorMsg = taskResp.Output.Message
		}
		// Check output.code
		if errorMsg == "" && taskResp.Output.Code != "" {
			errorMsg = taskResp.Output.Code
		}
		// Check top-level code
		if errorMsg == "" && taskResp.Code != "" {
			errorMsg = taskResp.Code
		}
		// Fallback: Try to extract from full response JSON (for edge cases)
		if errorMsg == "" {
			var fullResp map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &fullResp); err == nil {
				if output, ok := fullResp["output"].(map[string]interface{}); ok {
					if msg, ok := output["message"].(string); ok && msg != "" {
						errorMsg = msg
					} else if code, ok := output["code"].(string); ok && code != "" {
						errorMsg = code
					}
				}
				if errorMsg == "" {
					if msg, ok := fullResp["message"].(string); ok && msg != "" {
						errorMsg = msg
					} else if code, ok := fullResp["code"].(string); ok && code != "" {
						errorMsg = code
					}
				}
			}
		}
		if errorMsg == "" {
			errorMsg = "Video generation failed (unknown reason)"
		}
		message = fmt.Sprintf("Video generation failed: %s", errorMsg)
		p.logger.Error("DashScope task failed",
			zap.String("task_id", providerJobID),
			zap.String("error", errorMsg),
			zap.String("code", taskResp.Code),
			zap.String("message", taskResp.Message),
			zap.String("full_response", string(bodyBytes)),
		)
	default:
		progress = 30
		stage = "PROCESSING"
		message = fmt.Sprintf("Video generation: %s", status)
		p.logger.Warn("Unknown DashScope task status",
			zap.String("task_id", providerJobID),
			zap.String("status", status),
		)
	}

	progressResult := &entity.Progress{
		Percent: progress,
		Stage:   stage,
		Message: message,
	}

	// Store video URL in a way that can be retrieved (we'll need to update entity.Progress)
	// For now, log it and the worker will fetch it from the task response
	if videoURL != "" {
		p.logger.Info("DashScope video URL available",
			zap.String("task_id", providerJobID),
			zap.String("video_url", videoURL),
		)
	}

	return progressResult, nil
}

// GetVideoURL retrieves the video URL from a completed DashScope task
func (p *WanAIProvider) GetVideoURL(ctx context.Context, providerJobID string) (string, error) {
	// Fetch task status to get video URL
	baseURL := p.baseURL
	url := fmt.Sprintf("%s/tasks/%s", baseURL, providerJobID)
	// Replace /compatible-mode/v1 with /api/v1 if needed
	if baseURL == "https://dashscope-intl.aliyuncs.com/compatible-mode/v1" {
		url = fmt.Sprintf("https://dashscope-intl.aliyuncs.com/api/v1/tasks/%s", providerJobID)
	} else if baseURL == "https://dashscope.aliyuncs.com/compatible-mode/v1" {
		url = fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/tasks/%s", providerJobID)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DashScope task status error: %d", resp.StatusCode)
	}

	var taskResp DashScopeTaskResponse
	if err := json.Unmarshal(bodyBytes, &taskResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if taskResp.Code != "" && taskResp.Code != "Success" {
		return "", fmt.Errorf("DashScope error: %s - %s", taskResp.Code, taskResp.Message)
	}

	// Get video URL from response
	videoURL := taskResp.Output.VideoURL
	if videoURL == "" {
		videoURL = taskResp.Output.Video
	}

	return videoURL, nil
}

// CancelGeneration cancels an ongoing generation
// Note: DashScope may not support cancellation, this is a placeholder
func (p *WanAIProvider) CancelGeneration(ctx context.Context, providerJobID string) error {
	// DashScope doesn't have a standard cancel endpoint
	// Task will complete or fail on its own
	p.logger.Warn("CancelGeneration called but DashScope may not support cancellation",
		zap.String("task_id", providerJobID),
	)
	return fmt.Errorf("cancellation not supported by DashScope")
}

// GetCapabilities returns provider capabilities
func (p *WanAIProvider) GetCapabilities() service.ProviderCapabilities {
	return service.ProviderCapabilities{
		Name:            "Wan AI",
		MaxDuration:     60,
		MaxResolution:   entity.Resolution4K,
		SupportedRatios: []entity.AspectRatio{entity.AspectRatio16x9, entity.AspectRatio9x16, entity.AspectRatio1x1},
		EstimatedTime:   20,
		QualityTier:     "premium",
		SupportsStyles:  true,
		CostPerSecond:   0.03,
	}
}

// HealthCheck performs a health check
func (p *WanAIProvider) HealthCheck(ctx context.Context) (*service.ProviderHealth, error) {
	// Make a simple request to check if DashScope API is available
	// Try to list models or make a minimal request
	baseURL := p.baseURL
	url := fmt.Sprintf("%s/models", baseURL)
	// Replace /compatible-mode/v1 with /api/v1 if needed
	if baseURL == "https://dashscope-intl.aliyuncs.com/compatible-mode/v1" {
		url = "https://dashscope-intl.aliyuncs.com/api/v1/models"
	} else if baseURL == "https://dashscope.aliyuncs.com/compatible-mode/v1" {
		url = "https://dashscope.aliyuncs.com/api/v1/models"
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	start := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		return &service.ProviderHealth{
			IsHealthy:    false,
			ResponseTime: responseTime,
			ErrorRate:    1.0,
			LastChecked:  time.Now().Unix(),
		}, nil
	}
	defer resp.Body.Close()

	// 200 OK or 401 Unauthorized (API is reachable, just auth issue) means API is healthy
	isHealthy := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized

	return &service.ProviderHealth{
		IsHealthy:    isHealthy,
		QueueDepth:   0, // Would need to be tracked separately
		ResponseTime: responseTime,
		ErrorRate:    0.0,
		LastChecked:  time.Now().Unix(),
	}, nil
}

// downloadAndCacheImage downloads an image from a URL and caches it locally
// Returns the URL to access the cached image via the static endpoint
func (p *WanAIProvider) downloadAndCacheImage(ctx context.Context, imageURL string, templateID string) (string, error) {
	// Create cache directory if it doesn't exist
	cacheDir := "./static/temp-images"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Generate cache filename from URL hash
	hash := md5.Sum([]byte(imageURL))
	filename := hex.EncodeToString(hash[:]) + ".png"
	cachePath := filepath.Join(cacheDir, filename)

	// Check if already cached
	if _, err := os.Stat(cachePath); err == nil {
		// File exists, return cached URL
		// Use full URL with server base URL so DashScope can access it
		if p.serverBaseURL != "" {
			// Use HTTPS if available, but DashScope should be able to access it
			return fmt.Sprintf("%s/temp-images/%s", p.serverBaseURL, filename), nil
		}
		return fmt.Sprintf("/temp-images/%s", filename), nil
	}

	// Download the image with retry logic and better error handling
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Bypass SSL for problematic domains
		},
	}
	
	// Try HTTPS first, fallback to HTTP if needed
	var resp *http.Response
	var err error
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "image/*,*/*")
		req.Header.Set("Referer", imageURL)

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		
		// If HTTPS failed and we haven't tried HTTP yet, try HTTP
		if attempt == 0 && strings.HasPrefix(imageURL, "https://") {
			imageURL = strings.Replace(imageURL, "https://", "http://", 1)
			continue
		}
		
		if resp != nil {
			resp.Body.Close()
		}
		
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	
	if err != nil {
		return "", fmt.Errorf("failed to download image after %d attempts: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Save to cache
	file, err := os.Create(cachePath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(cachePath) // Clean up on error
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	// Return URL to cached image (use full HTTPS URL so DashScope can access it)
	if p.serverBaseURL != "" {
		return fmt.Sprintf("%s/temp-images/%s", p.serverBaseURL, filename), nil
	}
	return fmt.Sprintf("/temp-images/%s", filename), nil
}
