package events

import (
	"sync"
	"testing"
	"time"

	"github.com/feline-dis/go-radio-v2/internal/models"
)

func TestNewEventBus(t *testing.T) {
	eventBus := NewEventBus()
	if eventBus == nil {
		t.Fatal("Expected EventBus to be created, got nil")
	}
	if eventBus.handlers == nil {
		t.Fatal("Expected handlers map to be initialized")
	}
}

func TestSubscribeAndPublish(t *testing.T) {
	eventBus := NewEventBus()

	var receivedEvent Event
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event Event) {
		receivedEvent = event
		wg.Done()
	}

	eventBus.Subscribe("test_event", handler)

	testEvent := Event{
		Type:      "test_event",
		Payload:   "test payload",
		Timestamp: time.Now(),
	}

	eventBus.Publish(testEvent)

	// Wait for handler to be called
	wg.Wait()

	if receivedEvent.Type != "test_event" {
		t.Errorf("Expected event type 'test_event', got '%s'", receivedEvent.Type)
	}

	if receivedEvent.Payload != "test payload" {
		t.Errorf("Expected payload 'test payload', got '%v'", receivedEvent.Payload)
	}
}

func TestPublishSongChange(t *testing.T) {
	eventBus := NewEventBus()

	var receivedEvent Event
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event Event) {
		receivedEvent = event
		wg.Done()
	}

	eventBus.Subscribe(EventSongChange, handler)

	currentSong := &models.Song{
		YouTubeID: "test123",
		Title:     "Test Song",
		Artist:    "Test Artist",
		Duration:  180,
	}

	nextSong := &models.Song{
		YouTubeID: "test456",
		Title:     "Next Song",
		Artist:    "Next Artist",
		Duration:  200,
	}

	eventBus.PublishSongChange(currentSong, nextSong)

	// Wait for handler to be called
	wg.Wait()

	if receivedEvent.Type != EventSongChange {
		t.Errorf("Expected event type '%s', got '%s'", EventSongChange, receivedEvent.Type)
	}

	songChangeEvent, ok := receivedEvent.Payload.(SongChangeEvent)
	if !ok {
		t.Fatal("Expected payload to be SongChangeEvent")
	}

	if songChangeEvent.CurrentSong.YouTubeID != "test123" {
		t.Errorf("Expected current song ID 'test123', got '%s'", songChangeEvent.CurrentSong.YouTubeID)
	}

	if songChangeEvent.NextSong.YouTubeID != "test456" {
		t.Errorf("Expected next song ID 'test456', got '%s'", songChangeEvent.NextSong.YouTubeID)
	}
}

func TestPublishQueueUpdate(t *testing.T) {
	eventBus := NewEventBus()

	var receivedEvent Event
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event Event) {
		receivedEvent = event
		wg.Done()
	}

	eventBus.Subscribe(EventQueueUpdate, handler)

	queueInfo := &models.QueueInfo{
		CurrentSong: &models.Song{
			YouTubeID: "test123",
			Title:     "Test Song",
			Artist:    "Test Artist",
			Duration:  180,
		},
		NextSong: &models.Song{
			YouTubeID: "test456",
			Title:     "Next Song",
			Artist:    "Next Artist",
			Duration:  200,
		},
		Queue: []*models.Song{
			{YouTubeID: "test123", Title: "Test Song", Artist: "Test Artist", Duration: 180},
			{YouTubeID: "test456", Title: "Next Song", Artist: "Next Artist", Duration: 200},
		},
		Remaining: 120.5,
		StartTime: time.Now(),
	}

	eventBus.PublishQueueUpdate(queueInfo)

	// Wait for handler to be called
	wg.Wait()

	if receivedEvent.Type != EventQueueUpdate {
		t.Errorf("Expected event type '%s', got '%s'", EventQueueUpdate, receivedEvent.Type)
	}

	queueUpdateEvent, ok := receivedEvent.Payload.(QueueUpdateEvent)
	if !ok {
		t.Fatal("Expected payload to be QueueUpdateEvent")
	}

	if queueUpdateEvent.CurrentSong.YouTubeID != "test123" {
		t.Errorf("Expected current song ID 'test123', got '%s'", queueUpdateEvent.CurrentSong.YouTubeID)
	}

	if queueUpdateEvent.Remaining != 120.5 {
		t.Errorf("Expected remaining time 120.5, got %f", queueUpdateEvent.Remaining)
	}

	if len(queueUpdateEvent.Queue) != 2 {
		t.Errorf("Expected queue length 2, got %d", len(queueUpdateEvent.Queue))
	}
}

func TestMultipleHandlers(t *testing.T) {
	eventBus := NewEventBus()

	handler1Called := false
	handler2Called := false
	var wg sync.WaitGroup
	wg.Add(2)

	handler1 := func(event Event) {
		handler1Called = true
		wg.Done()
	}

	handler2 := func(event Event) {
		handler2Called = true
		wg.Done()
	}

	eventBus.Subscribe("test_event", handler1)
	eventBus.Subscribe("test_event", handler2)

	testEvent := Event{
		Type:      "test_event",
		Payload:   "test payload",
		Timestamp: time.Now(),
	}

	eventBus.Publish(testEvent)

	// Wait for both handlers to be called
	wg.Wait()

	if !handler1Called {
		t.Error("Expected handler1 to be called")
	}

	if !handler2Called {
		t.Error("Expected handler2 to be called")
	}
}

func TestHandlerPanicRecovery(t *testing.T) {
	eventBus := NewEventBus()

	handlerCalled := false
	var wg sync.WaitGroup
	wg.Add(1)

	// Create a handler that panics
	panicHandler := func(event Event) {
		panic("test panic")
	}

	// Create a normal handler
	normalHandler := func(event Event) {
		handlerCalled = true
		wg.Done()
	}

	eventBus.Subscribe("test_event", panicHandler)
	eventBus.Subscribe("test_event", normalHandler)

	testEvent := Event{
		Type:      "test_event",
		Payload:   "test payload",
		Timestamp: time.Now(),
	}

	// This should not cause the test to panic
	eventBus.Publish(testEvent)

	// Wait for the normal handler to be called
	wg.Wait()

	if !handlerCalled {
		t.Error("Expected normal handler to be called even after panic")
	}
}
