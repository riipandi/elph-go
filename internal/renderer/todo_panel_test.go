package renderer

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/riipandi/elph/pkg/ai/provider"
	"github.com/riipandi/elph/pkg/core/agent"
	"github.com/riipandi/elph/pkg/tools/todolist"
	"github.com/stretchr/testify/require"
)

func TestTodoPanelShowsTasksWithStatusMarkers(t *testing.T) {
	m := testInputModel(t)
	m.session.ReplaceTodos([]todolist.Todo{
		{Title: "read auth.go", Status: todolist.StatusDone},
		{Title: "patch handler", Status: todolist.StatusInProgress},
		{Title: "run tests", Status: todolist.StatusPending},
	})

	rendered := stripANSI(m.todoPanelView())
	require.Contains(t, rendered, "Tasks")
	require.Contains(t, rendered, "read auth.go")
	require.Contains(t, rendered, "patch handler")
	require.Contains(t, rendered, "run tests")
}

func TestTodoPanelMatchesInputChromeWidth(t *testing.T) {
	m := testInputModel(t)
	m.width = 80
	m.content.SetWidth(m.targetContentWidth())
	m.session.ReplaceTodos([]todolist.Todo{{Title: "align width", Status: todolist.StatusPending}})

	panelW := lipgloss.Width(m.todoPanelView())
	inputW := lipgloss.Width(m.inputBoxView(false))
	require.Equal(t, inputW, panelW)
}

func TestTodoPanelHiddenWhenEmpty(t *testing.T) {
	m := testInputModel(t)
	require.Empty(t, m.todoPanelView())
}

func TestTodoPanelHiddenWhenAllDone(t *testing.T) {
	m := testInputModel(t)
	m.session.ReplaceTodos([]todolist.Todo{
		{Title: "read auth.go", Status: todolist.StatusDone},
		{Title: "run tests", Status: todolist.StatusDone},
	})
	require.Empty(t, m.todoPanelView())
}

func TestTodoListToolSkipsDetailBox(t *testing.T) {
	m := testInputModel(t)
	call := provider.ToolCall{ID: "call_todo", Name: "TodoList"}
	m = m.beginNativeToolCall(call)
	require.True(t, m.agent.TodoListUpdating)
	require.Empty(t, m.messages)

	m.session.ReplaceTodos([]todolist.Todo{{Title: "a", Status: todolist.StatusPending}})
	m = m.finishNativeToolCall(call, agent.ToolRunResult{Output: "updated"})
	require.False(t, m.agent.TodoListUpdating)
	require.Empty(t, m.messages)
}

func TestTodoListCompletionNotifiesAndClearsPanel(t *testing.T) {
	m := testInputModel(t)
	call := provider.ToolCall{ID: "call_todo", Name: "TodoList"}
	m.session.ReplaceTodos([]todolist.Todo{
		{Title: "read auth.go", Status: todolist.StatusDone},
		{Title: "run tests", Status: todolist.StatusInProgress},
	})
	m.agent.TodoListBefore = append([]todolist.Todo(nil), m.session.Todos()...)
	m.session.ReplaceTodos([]todolist.Todo{
		{Title: "read auth.go", Status: todolist.StatusDone},
		{Title: "run tests", Status: todolist.StatusDone},
	})

	m = m.finishNativeToolCall(call, agent.ToolRunResult{Output: "updated"})
	require.False(t, m.agent.TodoListUpdating)
	require.Empty(t, m.session.Todos())
	require.Empty(t, m.todoPanelView())
	require.Len(t, m.messages, 1)
	require.Contains(t, m.messages[0].text, "All tasks completed")
	require.Contains(t, m.messages[0].text, "run tests")
}
