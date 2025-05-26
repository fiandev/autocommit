package service

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"
	"google.golang.org/genai"

	"encoding/json"
)

//go:embed prompts/luminai_prompt.md
var luminaiPrompt string

type LuminaiService struct {
	luminaiPrompt string
}

// LuminaiCommitOptions contains options for commit generation
type LuminaiCommitOptions struct {
	StageAll    *bool
	UserContext *string
	Model       *string
	NoConfirm   *bool
	Quiet       *bool
	Push        *bool
	DryRun      *bool
	ShowDiff    *bool
	MaxLength   *int
	Language    *string
	Issue       *string
	NoVerify    *bool
}

// LuminaiPreCommitData contains data about the changes to be committed
type LuminaiPreCommitData struct {
	Files        []string
	Diff         string
	RelatedFiles map[string]string
	Issue        string
}

var (
	luminaiService *LuminaiService
	luminaiOnce    sync.Once
)

func NewLuminaiService() *LuminaiService {
	luminaiOnce.Do(func() {
		luminaiService = &LuminaiService{
			luminaiPrompt: luminaiPrompt,
		}
	})

	return luminaiService
}

// GenerateCommitMessage creates a commit message using AI analysis with UI feedback
func (g *LuminaiService) GenerateCommitMessage(
	client *genai.Client,
	ctx context.Context,
	data *LuminaiPreCommitData,
	opts *LuminaiCommitOptions,
) (string, error) {
	messageChan := make(chan string, 1)

	if !*opts.Quiet {
		if err := spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *opts.Model)).
			Action(func() {
				g.analyzeToChannel(client, ctx, data, opts, messageChan)
			}).
			Run(); err != nil {
			return "", err
		}
	} else {
		g.analyzeToChannel(client, ctx, data, opts, messageChan)
	}

	message := <-messageChan
	if !*opts.Quiet {
		underline := color.New(color.Underline)
		underline.Println("\nChanges analyzed!")
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return "", fmt.Errorf("no commit messages were generated. try again")
	}

	return message, nil
}

// analyzeToChannel performs the actual AI analysis and sends result to channel
func (g *LuminaiService) analyzeToChannel(
	client *genai.Client,
	ctx context.Context,
	data *LuminaiPreCommitData,
	opts *LuminaiCommitOptions,
	messageChan chan string,
) {
	message, err := g.AnalyzeChanges(
		client,
		ctx,
		data.Diff,
		opts.UserContext,
		&data.RelatedFiles,
		opts.Model,
		opts.MaxLength,
		opts.Language,
		&data.Issue,
	)
	if err != nil {
		messageChan <- ""
	} else {
		messageChan <- message
	}
}

func (g *LuminaiService) GetUserPrompt(
	context *string,
	diff string,
	files []string,
	maxLength *int,
	language *string,
	issue *string,
	// lastCommits []string,
) (string, error) {
	if *context != "" {
		temp := fmt.Sprintf("Use the following context to understand intent: %s", *context)
		context = &temp
	} else {
		*context = ""
	}

	prompt := fmt.Sprintf(
		`%s

Code diff:
%s

Neighboring files:
%s

Requirements:
- Maximum commit message length: %d characters
- Language: %s`,
		*context,
		diff,
		strings.Join(files, ", "),
		*maxLength,
		*language,
	)

	if *issue != "" {
		prompt += fmt.Sprintf("\n- Reference issue: %s", *issue)
	}

	return prompt, nil
}

func (g *LuminaiService) AnalyzeChanges(
	geminiClient *genai.Client,
	ctx context.Context,
	diff string,
	userContext *string,
	relatedFiles *map[string]string,
	modelName *string,
	maxLength *int,
	language *string,
	issue *string,
	// lastCommits []string,
) (string, error) {
	// format relatedFiles to be dir : files
	relatedFilesArray := make([]string, 0, len(*relatedFiles))
	for dir, ls := range *relatedFiles {
		relatedFilesArray = append(relatedFilesArray, fmt.Sprintf("%s/%s", dir, ls))
	}

	userPrompt, err := g.GetUserPrompt(userContext, diff, relatedFilesArray, maxLength, language, issue)
	if err != nil {
		return "", err
	}

	// Update system prompt to include language and length requirements
	enhancedluminaiPrompt := g.luminaiPrompt
	if *language != "english" {
		enhancedluminaiPrompt += fmt.Sprintf("\n\nIMPORTANT: Generate the commit message in %s language.", *language)
	}
	enhancedluminaiPrompt += fmt.Sprintf("\n\nIMPORTANT: Keep the commit message under %d characters.", *maxLength)
	if *issue != "" {
		enhancedluminaiPrompt += fmt.Sprintf("\n\nIMPORTANT: Reference issue %s in the commit message.", *issue)
	}

	url := "https://luminai.my.id/"

	payload := map[string]interface{}{
		"prompt":  "",
		"content": userPrompt,
		"user":    "",
	}

	jsonData, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Request gagal: %v", err)
	}
	defer resp.Body.Close()

	// Baca response body
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error:", err)
		return "", nil
	}

	var responseData struct {
		result string `json:"result"`
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		log.Fatalf("Gagal decode JSON response: %v", err)
	}

	result := responseData.result
	result = strings.ReplaceAll(result, "```", "")
	result = strings.TrimSpace(result)

	return result, nil
}
