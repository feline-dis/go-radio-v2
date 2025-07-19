package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4"))

	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))
)

type setupStep int

const (
	stepWelcome setupStep = iota
	stepStorageType
	stepDataDir
	stepMetadataType
	stepS3Config
	stepConfirmation
	stepComplete
)

type model struct {
	step     setupStep
	cursor   int
	choices  []string
	config   map[string]string
	error    string
	quitting bool
}

func initialModel() model {
	return model{
		step:    stepWelcome,
		choices: []string{"Continue", "Exit"},
		config:  make(map[string]string),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			return m.handleSelection()
		}
	}
	return m, nil
}

func (m model) handleSelection() (tea.Model, tea.Cmd) {
	m.error = ""
	
	switch m.step {
	case stepWelcome:
		if m.cursor == 0 { // Continue
			m.step = stepStorageType
			m.choices = []string{"Local Files", "AWS S3"}
			m.cursor = 0
		} else { // Exit
			m.quitting = true
			return m, tea.Quit
		}
	
	case stepStorageType:
		if m.cursor == 0 { // Local Files
			m.config["FILE_STORAGE_TYPE"] = "local"
			m.step = stepDataDir
			m.choices = []string{"./data", "~/go-radio-data", "Custom path"}
			m.cursor = 0
		} else { // AWS S3
			m.config["FILE_STORAGE_TYPE"] = "s3"
			m.step = stepS3Config
			m.choices = []string{"Configure S3", "Skip for now"}
			m.cursor = 0
		}
	
	case stepDataDir:
		switch m.cursor {
		case 0:
			m.config["LOCAL_DATA_DIR"] = "./data"
		case 1:
			homeDir, _ := os.UserHomeDir()
			m.config["LOCAL_DATA_DIR"] = filepath.Join(homeDir, "go-radio-data")
		case 2:
			// TODO: Add text input for custom path
			m.config["LOCAL_DATA_DIR"] = "./data"
		}
		m.step = stepMetadataType
		m.choices = []string{"SQLite Database", "JSON Files"}
		m.cursor = 0
	
	case stepMetadataType:
		if m.cursor == 0 { // SQLite
			m.config["METADATA_STORAGE_TYPE"] = "sqlite"
			dataDir := m.config["LOCAL_DATA_DIR"]
			if dataDir == "" {
				dataDir = "./data"
			}
			m.config["SQLITE_DB_PATH"] = filepath.Join(dataDir, "radio.db")
		} else { // JSON
			m.config["METADATA_STORAGE_TYPE"] = "json"
			m.config["JSON_DATA_DIR"] = m.config["LOCAL_DATA_DIR"]
		}
		m.step = stepConfirmation
		m.choices = []string{"Apply Configuration", "Go Back", "Exit"}
		m.cursor = 0
	
	case stepS3Config:
		// TODO: Implement S3 configuration
		m.step = stepConfirmation
		m.choices = []string{"Apply Configuration", "Go Back", "Exit"}
		m.cursor = 0
	
	case stepConfirmation:
		switch m.cursor {
		case 0: // Apply
			err := m.writeConfig()
			if err != nil {
				m.error = err.Error()
			} else {
				m.step = stepComplete
				m.choices = []string{"Exit"}
				m.cursor = 0
			}
		case 1: // Go Back
			m.step = stepWelcome
			m.choices = []string{"Continue", "Exit"}
			m.cursor = 0
		case 2: // Exit
			m.quitting = true
			return m, tea.Quit
		}
	
	case stepComplete:
		m.quitting = true
		return m, tea.Quit
	}
	
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for using Go Radio setup!\n"
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render("Go Radio v2 Setup"))
	s.WriteString("\n\n")

	switch m.step {
	case stepWelcome:
		s.WriteString("Welcome to Go Radio v2!\n\n")
		s.WriteString("This setup wizard will help you configure your local radio service.\n")
		s.WriteString("You'll be able to choose between local file storage or cloud storage,\n")
		s.WriteString("configure metadata storage, and set up tunneling for external access.\n\n")

	case stepStorageType:
		s.WriteString("Choose your file storage backend:\n\n")
		s.WriteString("• Local Files: Store audio files on your local filesystem\n")
		s.WriteString("• AWS S3: Store audio files in Amazon S3 (requires AWS credentials)\n\n")

	case stepDataDir:
		s.WriteString("Choose your local data directory:\n\n")
		s.WriteString("This is where your audio files and database will be stored.\n\n")

	case stepMetadataType:
		s.WriteString("Choose your metadata storage:\n\n")
		s.WriteString("• SQLite Database: Single file database (recommended)\n")
		s.WriteString("• JSON Files: Store metadata in JSON files\n\n")

	case stepS3Config:
		s.WriteString("Configure AWS S3 settings:\n\n")
		s.WriteString("You'll need to set AWS credentials and bucket name in environment variables.\n\n")


	case stepConfirmation:
		s.WriteString("Configuration Summary:\n\n")
		for key, value := range m.config {
			s.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
		s.WriteString("\n")

	case stepComplete:
		s.WriteString("Setup complete!\n\n")
		s.WriteString("Configuration has been written to .env file.\n")
		s.WriteString("You can now run 'make run' to start your radio service.\n\n")
	}

	// Show error if any
	if m.error != "" {
		s.WriteString(errorStyle.Render("Error: " + m.error))
		s.WriteString("\n\n")
	}

	// Show choices
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			choice = selectedStyle.Render(choice)
		} else {
			choice = unselectedStyle.Render(choice)
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Use arrow keys to navigate, Enter to select, q or Ctrl+C to quit"))

	return s.String()
}

func (m model) writeConfig() error {
	// Create .env file with configuration
	envContent := "# Go Radio v2 Configuration\n"
	envContent += "# Generated by setup wizard\n\n"

	// Server configuration
	envContent += "PORT=8080\n"
	envContent += "LOG_LEVEL=info\n\n"

	// Storage configuration
	for key, value := range m.config {
		envContent += fmt.Sprintf("%s=%s\n", key, value)
	}

	// Add other default configuration
	envContent += "\n# JWT Configuration\n"
	envContent += "JWT_SECRET=your-secret-key-here\n"
	envContent += "JWT_EXPIRATION=24h\n\n"

	envContent += "# Admin Configuration\n"
	envContent += "ADMIN_USERNAME=admin\n"
	envContent += "ADMIN_PASSWORD=admin\n\n"

	envContent += "# YouTube Configuration\n"
	envContent += "YOUTUBE_API_KEY=your-youtube-api-key-here\n\n"

	// Create data directory if it doesn't exist
	if dataDir, exists := m.config["LOCAL_DATA_DIR"]; exists {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
		
		// Create audio subdirectory
		audioDir := filepath.Join(dataDir, "audio")
		if err := os.MkdirAll(audioDir, 0755); err != nil {
			return fmt.Errorf("failed to create audio directory: %w", err)
		}
	}

	// Write .env file
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

func main() {
	fmt.Println("Starting Go Radio v2 Setup...")
	
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}