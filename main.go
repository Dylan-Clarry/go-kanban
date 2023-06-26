package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type status int

const divisor = 4

const (
    todo status = iota
    inProgress
    done
)

/* MODEL MANAGEMENT */
var models []tea.Model
const (
    model status = iota
    form
)

/* STYLING */
var (
    columnStyle = lipgloss.NewStyle().
        Padding(1, 2)
    focusedStyle = lipgloss.NewStyle().
        Padding(1, 2).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62"))
    helpStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("241"))
)

/* CUSTOM ITEM */

type Task struct {
    status status
    title string
    description string
}

func NewTask(status status, title, description string) *Task {
    return &Task{status: status, title: title, description: description}
}

func (t *Task) Next() {
    if t.status == done {
        t.status = todo
    } else {
        t.status++
    }
}

func (t *Task) Prev() {
    if t.status == todo {
        t.status = done
    } else {
        t.status--
    }
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
    focused status
    lists []list.Model
    err error
    loaded bool
    quitting bool
}

func New() *Model {
    return &Model{}
}

func (m *Model) MoveToNext() tea.Msg {
    selectedItem := m.lists[m.focused].SelectedItem()
	if selectedItem == nil { // will happen if board is empty
		return nil
	}
	selectedTask := selectedItem.(Task)
	m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
	selectedTask.Next()
	m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items())-1, list.Item(selectedTask))
	return nil
}

// TODO: Go to next list
func (m *Model) Next() {
    if m.focused == done {
        m.focused = todo
    } else {
        m.focused++
    }
}

// TODO: Go to prev list
func (m *Model) Prev() {
    if m.focused == todo {
        m.focused = done
    } else {
        m.focused--
    }
}

// TODO: call this on tea.WindowSizeMsg
func (m *Model) initLists(width, height int) {
    defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width/divisor, height / 2)
    defaultList.SetShowHelp(false)
    m.lists = []list.Model{defaultList, defaultList, defaultList}
    // init todo
    m.lists[todo].Title = "To Do"
    m.lists[todo].SetItems([]list.Item{
        Task{status: todo, title: "buy milk", description: "strawberry milk"},
        Task{status: todo, title: "eat sushi", description: "negitoro roll, miso soup"},
        Task{status: todo, title: "fold laundry", description: "or wear wrinkly t-shirts"},
    })
    // init in progress
    m.lists[inProgress].Title = "In Progress"
    m.lists[inProgress].SetItems([]list.Item{
        Task{status: todo, title: "Write code", description: "don't worry it's Go"},
    })
    // init done
    m.lists[done].Title = "Done"
    m.lists[done].SetItems([]list.Item{
        Task{status: todo, title: "Stay cool", description: "as a cucumber"},
    })
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
        case tea.WindowSizeMsg:
            if !m.loaded {
                columnStyle.Width(msg.Width / divisor)
                focusedStyle.Width(msg.Width / divisor)
                columnStyle.Height(msg.Height / divisor)
                focusedStyle.Height(msg.Height / divisor)
                m.initLists(msg.Width, msg.Height)
                m.loaded = true
            }
        case tea.KeyMsg:
            switch msg.String() {
            case "ctrl+c", "q":
                m.quitting = true
                return m, tea.Quit
            case "left", "h":
                m.Prev()
            case "right", "l":
                m.Next()
            case "enter":
                return m, m.MoveToNext
            case "n":
                models[model] = m // save the state of the current model
                models[form] = NewForm(m.focused)
                return models[form].Update(nil)
            }
        case Task:
            task := msg
            list := &m.lists[task.status]
            return m, list.InsertItem(len(list.Items()), task)
    }
    var cmd tea.Cmd
    m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
    return m, cmd
}

func (m Model) View() string {
    if m.quitting {
        return ""
    }
    if m.loaded {
        todoView := m.lists[todo].View()
        inProgressView := m.lists[inProgress].View()
        doneView := m.lists[done].View()
        switch m.focused {
        case inProgress:
            return lipgloss.JoinHorizontal(
                lipgloss.Left,
                columnStyle.Render(todoView),
                focusedStyle.Render(inProgressView),
                columnStyle.Render(doneView),
            )
        case done:
            return lipgloss.JoinHorizontal(
                lipgloss.Left,
                columnStyle.Render(todoView),
                columnStyle.Render(inProgressView),
                focusedStyle.Render(doneView),
            )
        default:
            return lipgloss.JoinHorizontal(
                lipgloss.Left,
                focusedStyle.Render(todoView),
                columnStyle.Render(inProgressView),
                columnStyle.Render(doneView),
            )
        }
    }
    return "loading..."
}

/* FORM MODEL */
type Form struct {
    focused status
    title textinput.Model
    description textarea.Model
}

func NewForm(focused status) *Form {
    form := Form{
        focused: focused,
        title: textinput.New(),
        description: textarea.New(),
    }
    form.title.Focus()
    return &form
}

func (m Form) CreateTask() tea.Msg {
    // TODO: create a new Task
    return NewTask(m.focused, m.title.Value(), m.description.Value())
}

func (m Form) Init() tea.Cmd {
    return nil
}

func (m Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m,tea.Quit
        case "enter":
            if m.title.Focused() {
                m.title.Blur()
                m.description.Focus()
                return m, textarea.Blink
            } else {
                models[form] = m
                return models[model], m.CreateTask
            }
        }
    }
    if m.title.Focused() {
        m.title, cmd = m.title.Update(msg)
        return m, cmd
    } else {
        m.description, cmd = m.description.Update(msg)
        return m, cmd
    }
}

func (m Form) View() string {
    return lipgloss.JoinVertical(lipgloss.Left, m.title.View(), m.description.View())
}

func main() {
    models := []tea.Model{New(), NewForm(todo)}
    m := models[model]
    p := tea.NewProgram(m)
    if err := p.Start(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
