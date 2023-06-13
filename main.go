package main

import (
	"os"
    "fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type status int

const (
    todo status = iota
    inProgress
    done
)

/* CUSTOM ITEM */

type Task struct {
    status status
    title string
    description string
}

// implement the list.Item interface
func (t Task) FilterValue() string {
    return t.title
}

func (t Task) Title() string {
    return t.title
}

func (t Task) Description() string {
    return t.description
}

/* MAIN MODEL */

type Model struct {
    lists []list.Model
    err error
}

func New() *Model {
    return &Model{}
}

// TODO: call this on tea.WindowSizeMsg
func (m *Model) initLists(width, height int) {
    defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
    m.lists = []list.Model{defaultList, defaultList, defaultList}
    m.lists.Title = "To Do"
    m.lists.SetItems([]list.Item{
        Task{status: todo, title: "buy milk", description: "strawberry milk"},
        Task{status: todo, title: "seat sushi", description: "negitoro roll, miso soup"},
        Task{status: todo, title: "fild laundry", description: "or wear wrinkly t-shirts"},
    })
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
        case tea.WindowSizeMsg:
            m.initLists(msg.Width, msg.Height)
    }
    var cmd tea.Cmd
    m.lists, cmd = m.lists.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    return m.lists.View()
}

func main() {
    m := New()
    p := tea.NewProgram(m)
    if err := p.Start(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
