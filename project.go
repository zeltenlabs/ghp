package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type Project struct {
	help     help.Model
	cursor   int
	projects []ProjectItem
	err      error
	Width    int
	Height   int
}

type ProjectItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func NewProject() *Project {
	help := help.New()
	help.ShowAll = true
	return &Project{cursor: 0, projects: []ProjectItem{}, err: nil}
}

func (m Project) Init() tea.Cmd {
	return loadProjects
}

func (m Project) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width - margin
		m.Height = msg.Height - margin
		m.help.Width = msg.Width - margin

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.projects) > 0 {
				board = NewBoard(m.projects[m.cursor].ID)
				return board.Update(tea.WindowSizeMsg{Height: m.Height, Width: m.Width})
			}

		}
	case error:
		m.err = msg

	case []ProjectItem:
		m.projects = msg
	}
	return m, nil
}

func (m Project) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	if len(m.projects) == 0 {
		return "Loading..."
	}
	if m.cursor >= len(m.projects) {
		m.cursor = len(m.projects) - 1
	}
	projectList := ""
	for i, project := range m.projects {
		if i == m.cursor {
			projectList += fmt.Sprintf("> %s\n", project.Title)
		} else {
			projectList += fmt.Sprintf("  %s\n", project.Title)
		}
	}
	return projectList
}

func loadProjects() tea.Msg {
	// TODO: pagination
	// TODO: dynamic org from config
	queryStr := `{"query":"{organization(login: \"zeltenlabs\") {projectsV2(first: 20) {nodes {id title}}}}"}`

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer([]byte(queryStr)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GITHUB_OAUTH_TOKEN")))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	var data graphqlResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		return err
	}

	projects := make([]ProjectItem, len(data.Data.Organization.ProjectsV2.Nodes))
	for i, node := range data.Data.Organization.ProjectsV2.Nodes {
		projects[i] = node
	}
	return projects
}

type graphqlResponse struct {
	Data struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"nodes"`
			} `json:"projectsV2"`
		} `json:"organization"`
	} `json:"data"`
}
