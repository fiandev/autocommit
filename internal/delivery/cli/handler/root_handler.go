package handler

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/fiandev/autocommit/internal/usecase"
)

type RootHandler struct {
	useCase *usecase.RootUsecase
}

var (
	rootHandlerInstance *RootHandler
	rootHandlerOnce     sync.Once
)

func NewRootHandler() *RootHandler {
	rootHandlerOnce.Do(func() {
		useCase := usecase.NewRootUsecase()

		rootHandlerInstance = &RootHandler{useCase}
	})

	return rootHandlerInstance
}

func (r *RootHandler) RootCommand(
	ctx context.Context,
	stageAll *bool,
	userContext *string,
	model *string,
	noConfirm *bool,
	quiet *bool,
	push *bool,
	dryRun *bool,
	showDiff *bool,
	maxLength *int,
	aiProvider *string,
	language *string,
	issue *string,
	noVerify *bool,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		modelFromConfig := viper.GetString("api.model")
		aiProviderConfig := viper.GetString("ai.provider")

		if aiProviderConfig == "" {
			fmt.Println(
				"Error: You need to choose AI provider, run this command to set AI Provider",
			)

			fmt.Print("\n")
			color.New(color.Bold).Print("geminicommit config provider set ")
			// color.New(color.Italic, color.Bold).Print("provider\n\n")
			os.Exit(1)
		}

		if aiProviderConfig != "" {
			*aiProvider = aiProviderConfig
		}

		if modelFromConfig != "" && *model == "gemini-2.0-flash" {
			*model = modelFromConfig
		}

		if *quiet && !*noConfirm {
			*quiet = false
		}

		apiKey := viper.GetString("api.key")
		if apiKey == "" {
			fmt.Println(
				"Error: API key is still empty, run this command to set your API key",
			)
			fmt.Print("\n")
			color.New(color.Bold).Print("geminicommit config key set ")
			color.New(color.Italic, color.Bold).Print("api_key\n\n")
			os.Exit(1)
		}

		err := r.useCase.RootCommand(ctx, apiKey, stageAll, userContext, model, noConfirm, quiet, push, dryRun, showDiff, maxLength, language, issue, noVerify)
		cobra.CheckErr(err)
	}
}
