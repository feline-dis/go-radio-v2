package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type YouTubeService struct {
	apiKey     string
	httpClient *http.Client
}

type YouTubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

type YouTubeVideoResponse struct {
	Items []struct {
		ContentDetails struct {
			Duration string `json:"duration"`
		} `json:"contentDetails"`
	} `json:"items"`
}

type SearchResult struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Duration    string `json:"duration"`
}

func NewYouTubeService() (*YouTubeService, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("YOUTUBE_API_KEY environment variable is not set")
	}

	return &YouTubeService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (s *YouTubeService) SearchVideos(query string) ([]SearchResult, error) {
	// First, search for videos
	searchURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?part=snippet&q=%s&type=video&maxResults=10&key=%s",
		url.QueryEscape(query),
		s.apiKey,
	)

	resp, err := s.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search YouTube: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube API returned non-200 status code: %d", resp.StatusCode)
	}

	var searchResp YouTubeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Get video IDs for duration lookup
	videoIDs := make([]string, len(searchResp.Items))
	for i, item := range searchResp.Items {
		videoIDs[i] = item.ID.VideoID
	}

	// Get video durations
	durations, err := s.getVideoDurations(videoIDs)
	if err != nil {
		return nil, err
	}

	// Combine search results with durations
	results := make([]SearchResult, len(searchResp.Items))
	for i, item := range searchResp.Items {
		duration := "Unknown"
		if d, ok := durations[item.ID.VideoID]; ok {
			duration = d
		}

		results[i] = SearchResult{
			ID:          item.ID.VideoID,
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			Thumbnail:   item.Snippet.Thumbnails.Default.URL,
			Duration:    duration,
		}
	}

	return results, nil
}

func (s *YouTubeService) getVideoDurations(videoIDs []string) (map[string]string, error) {
	if len(videoIDs) == 0 {
		return nil, nil
	}

	// Join video IDs with commas
	ids := ""
	for i, id := range videoIDs {
		if i > 0 {
			ids += ","
		}
		ids += id
	}

	// Get video details
	detailsURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?part=contentDetails&id=%s&key=%s",
		ids,
		s.apiKey,
	)

	resp, err := s.httpClient.Get(detailsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get video details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube API returned non-200 status code: %d", resp.StatusCode)
	}

	var detailsResp YouTubeVideoResponse
	if err := json.NewDecoder(resp.Body).Decode(&detailsResp); err != nil {
		return nil, fmt.Errorf("failed to decode video details response: %w", err)
	}

	// Create a map of video IDs to durations
	durations := make(map[string]string)
	for i, item := range detailsResp.Items {
		durations[videoIDs[i]] = item.ContentDetails.Duration
	}

	return durations, nil
}
