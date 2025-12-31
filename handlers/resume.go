package handlers

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jsmonhq/jsmon-cli/resume"
)

// ResumeState holds the current resume state
type ResumeState struct {
	Config   *resume.Config
	Filename string
	Saved    bool
}

// NewResumeState creates a new resume state (only for file scanning)
// Note: API key and workspace ID are NOT stored in resume file for security
func NewResumeState(filename string, configType string, workspaceID, apiKey string) *ResumeState {
	return &ResumeState{
		Config: &resume.Config{
			Type:      configType,
			LastIndex: -1,
		},
		Filename: filename,
		Saved:    false,
	}
}

// Save saves the resume state
func (rs *ResumeState) Save() error {
	if rs.Config == nil {
		return fmt.Errorf("resume config is nil")
	}
	err := resume.Save(rs.Config, rs.Filename)
	if err == nil {
		rs.Saved = true
	}
	return err
}

// SetupSignalHandler sets up signal handling for graceful shutdown
func SetupSignalHandler(rs *ResumeState) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "\n%s[!] Interrupted. Saving resume state...%s\n", ColorRed, ColorReset)
		if rs != nil {
			if err := rs.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "%s[!] Failed to save resume state: %v%s\n", ColorRed, err, ColorReset)
			} else {
				fmt.Fprintf(os.Stderr, "%s[!] Resume state saved to: %s%s\n", ColorGreen, rs.Filename, ColorReset)
			}
		}
		os.Exit(1)
	}()
}
