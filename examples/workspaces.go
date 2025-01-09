package examples

import (
	"fmt"
	"irmin-sdk/client"
	"irmin-sdk/services"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func TestWorkspaces() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Read values from environment variables
	baseURL := os.Getenv("BASE_URL")
	apiToken := os.Getenv("API_TOKEN")
	locale := os.Getenv("LOCALE")

	if baseURL == "" || apiToken == "" || locale == "" {
		log.Fatalf("Missing required environment variables: BASE_URL, API_TOKEN, or LOCALE")
	}

	// Initialise the client and service
	apiClient := client.NewClient(baseURL, apiToken, locale)
	workspaceService := services.NewWorkspaceService(apiClient)

	// Create a new workspace
	workspace, res, err := workspaceService.CreateWorkspace("Test Workspace", "This is for SDK testing")
	if err != nil {
		fmt.Println("Error creating workspace:", err)
		return
	}
	fmt.Println(res.Message)
	fmt.Printf("Created workspace: %s (%s)\n", workspace.Name, workspace.Slug)

	// Switch to the new workspace
	res, err = workspaceService.SwitchWorkspace(workspace.Slug)
	if err != nil {
		fmt.Println("Error switching workspace:", err)
		return
	}
	fmt.Println(res.Message)

	// Delete the new workspace
	res, err = workspaceService.DeleteWorkspace(workspace.Slug)
	if err != nil {
		fmt.Println("Error deleting workspace:", err)
		return
	}
	fmt.Println(res.Message)

	// Fetch workspaces
	workspaces, _, err := workspaceService.FetchWorkspaces()
	if err != nil {
		fmt.Println("Error fetching workspaces:", err)
		return
	}

	for _, workspace := range workspaces {
		fmt.Printf("Workspace: %s (%s)\n", workspace.Name, workspace.Slug)
	}

	// Switch to the first workspace
	if len(workspaces) == 0 {
		fmt.Println("No workspaces found")
		return
	}
	_, err = workspaceService.SwitchWorkspace(workspaces[0].Slug)
	if err != nil {
		fmt.Println("Error switching workspace:", err)
		return
	}

	// Fetch the current workspace
	currentWorkspace, _, err := workspaceService.FetchWorkspace(workspaces[0].Slug)
	if err != nil {
		fmt.Println("Error fetching workspace:", err)
		return
	}
	fmt.Printf("Current workspace: %s (%s)\n", currentWorkspace.Name, currentWorkspace.Slug)
}
