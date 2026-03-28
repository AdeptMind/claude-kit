package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	projectService := &ProjectService{}
	roleService := &RoleService{}
	fileService := &FileService{}
	profileService := &ProfileService{}
	workflowService := &WorkflowService{}
	storyService := &StoryService{}
	terminalService := &TerminalService{}
	initService := &InitService{}
	mcpService := &MCPService{}
	mcpServerService := &MCPServerService{}
	knowledgeService := &KnowledgeService{}
	settingsService := &SettingsService{}

	err := wails.Run(&options.App{
		Title:  "Claude Kit",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 13, G: 17, B: 23, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			projectService.startup(ctx)
			workflowService.startup(ctx)
			terminalService.startup(ctx)
			mcpService.startup(ctx)
			mcpServerService.startup(ctx)
			knowledgeService.startup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			terminalService.StopAll()
			knowledgeService.shutdown()
			mcpServerService.Shutdown()
		},
		Bind: []interface{}{
			app,
			projectService,
			roleService,
			fileService,
			profileService,
			workflowService,
			storyService,
			terminalService,
			initService,
			mcpService,
			mcpServerService,
			knowledgeService,
			settingsService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
