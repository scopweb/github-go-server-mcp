package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

type MCPServer struct {
	githubClient *github.Client
}

// Protocolo JSON-RPC 2.0
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Estructuras MCP
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"inputSchema"`
}

type ToolInputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallResult struct {
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	server, err := NewMCPServer()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp := server.handleRequest(req)
		output, err := json.Marshal(resp)
		if err != nil {
			continue
		}
		
		fmt.Println(string(output))
	}
}

func NewMCPServer() (*MCPServer, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN required")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	githubClient := github.NewClient(tc)

	return &MCPServer{
		githubClient: githubClient,
	}, nil
}

func (s *MCPServer) handleRequest(req JSONRPCRequest) JSONRPCResponse {
	id := req.ID
	if id == nil {
		id = 0
	}
	
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
	}

	if req.JSONRPC != "2.0" {
		response.Error = &JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request: jsonrpc must be '2.0'",
		}
		return response
	}

	if req.Method == "" {
		response.Error = &JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request: method is required",
		}
		return response
	}

	switch req.Method {
	case "initialize":
		response.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "github-mcp",
				"version": "1.0.0",
			},
		}
	case "initialized":
		response.Result = map[string]interface{}{}
	case "tools/list":
		response.Result = s.listTools()
	case "tools/call":
		result, err := s.callTool(req.Params)
		if err != nil {
			response.Error = &JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	default:
		response.Error = &JSONRPCError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	return response
}

func (s *MCPServer) listTools() ToolsListResult {
	tools := []Tool{
		{
			Name:        "github_list_repos",
			Description: "Lista repositorios del usuario autenticado",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"type": {Type: "string", Description: "Tipo: all, owner, member"},
				},
			},
		},
		{
			Name:        "github_create_repo",
			Description: "Crea un nuevo repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"name":        {Type: "string", Description: "Nombre del repositorio"},
					"description": {Type: "string", Description: "Descripción del repositorio"},
					"private":     {Type: "boolean", Description: "Repositorio privado"},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "github_get_repo",
			Description: "Obtiene información de un repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_list_branches",
			Description: "Lista ramas de un repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_list_prs",
			Description: "Lista pull requests de un repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
					"state": {Type: "string", Description: "Estado: open, closed, all"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_create_pr",
			Description: "Crea un nuevo pull request",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
					"title": {Type: "string", Description: "Título del PR"},
					"body":  {Type: "string", Description: "Descripción del PR"},
					"head":  {Type: "string", Description: "Rama origen"},
					"base":  {Type: "string", Description: "Rama destino"},
				},
				Required: []string{"owner", "repo", "title", "head", "base"},
			},
		},
		{
			Name:        "github_list_issues",
			Description: "Lista issues de un repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
					"state": {Type: "string", Description: "Estado: open, closed, all"},
				},
				Required: []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_create_issue",
			Description: "Crea un nuevo issue",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner": {Type: "string", Description: "Propietario del repositorio"},
					"repo":  {Type: "string", Description: "Nombre del repositorio"},
					"title": {Type: "string", Description: "Título del issue"},
					"body":  {Type: "string", Description: "Descripción del issue"},
				},
				Required: []string{"owner", "repo", "title"},
			},
		},
		{
			Name:        "github_create_file",
			Description: "Crea un nuevo archivo en el repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":   {Type: "string", Description: "Propietario del repositorio"},
					"repo":    {Type: "string", Description: "Nombre del repositorio"},
					"path":    {Type: "string", Description: "Ruta del archivo"},
					"content": {Type: "string", Description: "Contenido del archivo"},
					"message": {Type: "string", Description: "Mensaje del commit"},
					"branch":  {Type: "string", Description: "Rama (opcional, default: main)"},
				},
				Required: []string{"owner", "repo", "path", "content", "message"},
			},
		},
		{
			Name:        "github_update_file",
			Description: "Actualiza un archivo existente en el repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":   {Type: "string", Description: "Propietario del repositorio"},
					"repo":    {Type: "string", Description: "Nombre del repositorio"},
					"path":    {Type: "string", Description: "Ruta del archivo"},
					"content": {Type: "string", Description: "Nuevo contenido del archivo"},
					"message": {Type: "string", Description: "Mensaje del commit"},
					"sha":     {Type: "string", Description: "SHA del archivo actual"},
					"branch":  {Type: "string", Description: "Rama (opcional, default: main)"},
				},
				Required: []string{"owner", "repo", "path", "content", "message", "sha"},
			},
		},
		{
			Name:        "github_get_file",
			Description: "Obtiene el contenido de un archivo del repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":  {Type: "string", Description: "Propietario del repositorio"},
					"repo":   {Type: "string", Description: "Nombre del repositorio"},
					"path":   {Type: "string", Description: "Ruta del archivo"},
					"branch": {Type: "string", Description: "Rama (opcional, default: main)"},
				},
				Required: []string{"owner", "repo", "path"},
			},
		},
		{
			Name:        "github_list_files",
			Description: "Lista archivos y directorios en una ruta del repositorio",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]Property{
					"owner":  {Type: "string", Description: "Propietario del repositorio"},
					"repo":   {Type: "string", Description: "Nombre del repositorio"},
					"path":   {Type: "string", Description: "Ruta del directorio (opcional, default: raíz)"},
					"branch": {Type: "string", Description: "Rama (opcional, default: main)"},
				},
				Required: []string{"owner", "repo"},
			},
		},
	}

	return ToolsListResult{Tools: tools}
}

func (s *MCPServer) callTool(params map[string]interface{}) (ToolCallResult, error) {
	name, ok := params["name"].(string)
	if !ok {
		return ToolCallResult{}, fmt.Errorf("tool name required")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	ctx := context.Background()
	var text string
	var err error

	switch name {
	case "github_list_repos":
		text, err = s.listRepositories(ctx, arguments)
	case "github_create_repo":
		text, err = s.createRepository(ctx, arguments)
	case "github_get_repo":
		text, err = s.getRepository(ctx, arguments)
	case "github_list_branches":
		text, err = s.listBranches(ctx, arguments)
	case "github_list_prs":
		text, err = s.listPullRequests(ctx, arguments)
	case "github_create_pr":
		text, err = s.createPullRequest(ctx, arguments)
	case "github_list_issues":
		text, err = s.listIssues(ctx, arguments)
	case "github_create_issue":
		text, err = s.createIssue(ctx, arguments)
	case "github_create_file":
		text, err = s.createFile(ctx, arguments)
	case "github_update_file":
		text, err = s.updateFile(ctx, arguments)
	case "github_get_file":
		text, err = s.getFile(ctx, arguments)
	case "github_list_files":
		text, err = s.listFiles(ctx, arguments)
	default:
		return ToolCallResult{}, fmt.Errorf("tool not found")
	}

	if err != nil {
		return ToolCallResult{}, err
	}

	return ToolCallResult{
		Content: []Content{{Type: "text", Text: text}},
	}, nil
}

func (s *MCPServer) listRepositories(ctx context.Context, args map[string]interface{}) (string, error) {
	listType := "all"
	if t, ok := args["type"].(string); ok {
		listType = t
	}

	opt := &github.RepositoryListOptions{Type: listType}
	repos, _, err := s.githubClient.Repositories.List(ctx, "", opt)
	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	for _, repo := range repos {
		result = append(result, map[string]interface{}{
			"name":        repo.GetName(),
			"description": repo.GetDescription(),
			"private":     repo.GetPrivate(),
			"url":         repo.GetHTMLURL(),
			"language":    repo.GetLanguage(),
			"stars":       repo.GetStargazersCount(),
		})
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (s *MCPServer) createRepository(ctx context.Context, args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	repo := &github.Repository{Name: github.String(name)}

	if desc, ok := args["description"].(string); ok {
		repo.Description = github.String(desc)
	}

	if private, ok := args["private"].(bool); ok {
		repo.Private = github.Bool(private)
	}

	createdRepo, _, err := s.githubClient.Repositories.Create(ctx, "", repo)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Repository '%s' created successfully: %s",
		createdRepo.GetName(), createdRepo.GetHTMLURL()), nil
}

func (s *MCPServer) getRepository(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	repo, _, err := s.githubClient.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{
		"name":        repo.GetName(),
		"description": repo.GetDescription(),
		"private":     repo.GetPrivate(),
		"url":         repo.GetHTMLURL(),
		"language":    repo.GetLanguage(),
		"stars":       repo.GetStargazersCount(),
		"forks":       repo.GetForksCount(),
		"issues":      repo.GetOpenIssuesCount(),
		"created":     repo.GetCreatedAt(),
		"updated":     repo.GetUpdatedAt(),
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (s *MCPServer) listBranches(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	branches, _, err := s.githubClient.Repositories.ListBranches(ctx, owner, repoName, nil)
	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	for _, branch := range branches {
		result = append(result, map[string]interface{}{
			"name":      branch.GetName(),
			"protected": branch.GetProtected(),
			"sha":       branch.GetCommit().GetSHA(),
		})
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (s *MCPServer) listPullRequests(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	state := "open"
	if s, ok := args["state"].(string); ok {
		state = s
	}

	opt := &github.PullRequestListOptions{State: state}
	prs, _, err := s.githubClient.PullRequests.List(ctx, owner, repoName, opt)
	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	for _, pr := range prs {
		result = append(result, map[string]interface{}{
			"number": pr.GetNumber(),
			"title":  pr.GetTitle(),
			"state":  pr.GetState(),
			"url":    pr.GetHTMLURL(),
			"user":   pr.GetUser().GetLogin(),
			"head":   pr.GetHead().GetRef(),
			"base":   pr.GetBase().GetRef(),
		})
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (s *MCPServer) createPullRequest(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	title, ok := args["title"].(string)
	if !ok {
		return "", fmt.Errorf("title required")
	}

	head, ok := args["head"].(string)
	if !ok {
		return "", fmt.Errorf("head branch required")
	}

	base, ok := args["base"].(string)
	if !ok {
		return "", fmt.Errorf("base branch required")
	}

	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
	}

	if body, ok := args["body"].(string); ok {
		pr.Body = github.String(body)
	}

	createdPR, _, err := s.githubClient.PullRequests.Create(ctx, owner, repoName, pr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Pull Request #%d created: %s",
		createdPR.GetNumber(), createdPR.GetHTMLURL()), nil
}

func (s *MCPServer) listIssues(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	state := "open"
	if s, ok := args["state"].(string); ok {
		state = s
	}

	opt := &github.IssueListByRepoOptions{State: state}
	issues, _, err := s.githubClient.Issues.ListByRepo(ctx, owner, repoName, opt)
	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	for _, issue := range issues {
		result = append(result, map[string]interface{}{
			"number": issue.GetNumber(),
			"title":  issue.GetTitle(),
			"state":  issue.GetState(),
			"url":    issue.GetHTMLURL(),
			"user":   issue.GetUser().GetLogin(),
		})
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (s *MCPServer) createIssue(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	title, ok := args["title"].(string)
	if !ok {
		return "", fmt.Errorf("title required")
	}

	issue := &github.IssueRequest{Title: github.String(title)}

	if body, ok := args["body"].(string); ok {
		issue.Body = github.String(body)
	}

	createdIssue, _, err := s.githubClient.Issues.Create(ctx, owner, repoName, issue)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Issue #%d created: %s",
		createdIssue.GetNumber(), createdIssue.GetHTMLURL()), nil
}

func (s *MCPServer) createFile(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("file path required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("file content required")
	}

	message, ok := args["message"].(string)
	if !ok {
		return "", fmt.Errorf("commit message required")
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	fileOptions := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
	}

	result, _, err := s.githubClient.Repositories.CreateFile(ctx, owner, repoName, path, fileOptions)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("File '%s' created successfully. Commit SHA: %s",
		path, result.Commit.GetSHA()), nil
}

func (s *MCPServer) updateFile(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("file path required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("file content required")
	}

	message, ok := args["message"].(string)
	if !ok {
		return "", fmt.Errorf("commit message required")
	}

	sha, ok := args["sha"].(string)
	if !ok {
		return "", fmt.Errorf("file SHA required")
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	fileOptions := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		SHA:     github.String(sha),
		Branch:  github.String(branch),
	}

	result, _, err := s.githubClient.Repositories.UpdateFile(ctx, owner, repoName, path, fileOptions)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("File '%s' updated successfully. Commit SHA: %s",
		path, result.Commit.GetSHA()), nil
}

func (s *MCPServer) getFile(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("file path required")
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	opt := &github.RepositoryContentGetOptions{Ref: branch}
	file, _, _, err := s.githubClient.Repositories.GetContents(ctx, owner, repoName, path, opt)
	if err != nil {
		return "", err
	}

	content, err := file.GetContent()
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{
		"path":     file.GetPath(),
		"content":  content,
		"sha":      file.GetSHA(),
		"size":     file.GetSize(),
		"encoding": file.GetEncoding(),
		"url":      file.GetHTMLURL(),
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}

func (s *MCPServer) listFiles(ctx context.Context, args map[string]interface{}) (string, error) {
	owner, ok := args["owner"].(string)
	if !ok {
		return "", fmt.Errorf("owner required")
	}

	repoName, ok := args["repo"].(string)
	if !ok {
		return "", fmt.Errorf("repository name required")
	}

	path := ""
	if p, ok := args["path"].(string); ok {
		path = p
	}

	branch := "main"
	if b, ok := args["branch"].(string); ok {
		branch = b
	}

	opt := &github.RepositoryContentGetOptions{Ref: branch}
	_, contents, _, err := s.githubClient.Repositories.GetContents(ctx, owner, repoName, path, opt)
	if err != nil {
		return "", err
	}

	var result []map[string]interface{}
	for _, item := range contents {
		result = append(result, map[string]interface{}{
			"name": item.GetName(),
			"path": item.GetPath(),
			"type": item.GetType(),
			"size": item.GetSize(),
			"sha":  item.GetSHA(),
			"url":  item.GetHTMLURL(),
		})
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output), nil
}