package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/list"
)

// projectDataToListItems converts the project data into list items for the kanban board
func projectDataToListItems(projectData *ProjectData) []list.Item {
	var items []list.Item
	for _, node := range projectData.Data.Node.Items.Nodes {
		task := Task{
			ID: node.ID,
		}
		for _, fieldValue := range node.FieldValues.Nodes {
			if fieldValue.Field.Name == "Title" {
				task.title = fieldValue.Text
			}
			if fieldValue.Field.Name == "Status" {
				task.status = StrToStatus(fieldValue.Name)
				log.Println("Status:", fieldValue.Name, "task.status:", task.status)

			}
		}
		items = append(items, task)
	}
	return items
}

func sortTasksByStatus(tasks []list.Item) ([]list.Item, []list.Item, []list.Item) {
	var todoTasks, inProgressTasks, doneTasks []list.Item
	for _, taskItem := range tasks {
		task := taskItem.(Task)
		switch task.status {
		case todo:
			todoTasks = append(todoTasks, task)
		case inProgress:
			inProgressTasks = append(inProgressTasks, task)
		case done:
			doneTasks = append(doneTasks, task)
		}
	}
	return todoTasks, inProgressTasks, doneTasks
}

func loadProjectData(projectID string) (*ProjectData, error) {
	// TODO: pagination
	// TODO: handle multiple views
	queryStr := fmt.Sprintf(`{"query":"query{ node(id: \"%s\") { ... on ProjectV2 { items(first: 20) { nodes{ id fieldValues(first: 8) { nodes{ ... on ProjectV2ItemFieldTextValue { text field { ... on ProjectV2FieldCommon {  name }}} ... on ProjectV2ItemFieldDateValue { date field { ... on ProjectV2FieldCommon { name } } } ... on ProjectV2ItemFieldSingleSelectValue { name field { ... on ProjectV2FieldCommon { name }}}}} content{ ... on DraftIssue { title body } ...on Issue { title assignees(first: 10) { nodes{ login }}} ...on PullRequest { title assignees(first: 10) { nodes{ login }}}}}}}}}"}`, projectID)

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer([]byte(queryStr)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GITHUB_OAUTH_TOKEN")))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var data ProjectData
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		return nil, err
	}
	return &data, nil
}

type ProjectData struct {
	Data struct {
		Node struct {
			Items struct {
				Nodes []struct {
					ID          string `json:"id"`
					FieldValues struct {
						Nodes []struct {
							Text  string `json:"text,omitempty"`
							Field struct {
								Name string `json:"name"`
							} `json:"field,omitempty"`
							Name string `json:"name,omitempty"`
						} `json:"nodes"`
					} `json:"fieldValues"`
					Content struct {
						Title     string `json:"title"`
						Assignees struct {
							Nodes []struct {
								Login string `json:"login"`
							} `json:"nodes"`
						} `json:"assignees"`
					} `json:"content,omitempty"`
					Content0 struct {
						Title string `json:"title"`
						Body  string `json:"body"`
					} `json:"content,omitempty"`
				} `json:"nodes"`
			} `json:"items"`
		} `json:"node"`
	} `json:"data"`
}
