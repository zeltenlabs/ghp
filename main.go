package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type status int

func (s status) getNext() status {
	if s == done {
		return todo
	}
	return s + 1
}

func (s status) getPrev() status {
	if s == todo {
		return done
	}
	return s - 1
}

// status to string
func (s status) String() string {
	switch s {
	case todo:
		return "todo"
	case inProgress:
		return "inProgress"
	case done:
		return "done"
	}
	return "unknown"
}

// string to status
func StrToStatus(s string) status {
	switch s {
	case "todo":
	case "Todo":
		return todo
	case "inProgress":
	case "In Progress":
		return inProgress
	case "done":
	case "Done":
		return done
	}
	return todo
}

const margin = 4

var board *Board

const (
	todo status = iota
	inProgress
	done
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	project := NewProject()
	p := tea.NewProgram(project)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
