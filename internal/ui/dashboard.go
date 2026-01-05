package ui

import (
	"fmt"
	"time"

	"github.com/agnivo988/Repo-lyzer/internal/analyzer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardModel struct {
	data       AnalysisResult
	BackToMenu bool
	width      int
	height     int
	showExport bool
	statusMsg  string
}

func NewDashboardModel() DashboardModel {
	return DashboardModel{}
}

func (m DashboardModel) Init() tea.Cmd { return nil }

func (m *DashboardModel) SetData(data AnalysisResult) {
	m.data = data
}

type exportMsg struct {
	err error
	msg string
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case exportMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Export failed: %v", msg.err)
		} else {
			m.statusMsg = msg.msg
		}
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg { return "clear_status" })

	case string:
		if msg == "clear_status" {
			m.statusMsg = ""
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.showExport {
				m.showExport = false
			} else {
				m.BackToMenu = true
			}
		case "e":
			m.showExport = !m.showExport
		case "j":
			if m.showExport {
				return m, func() tea.Msg {
					err := ExportJSON(m.data, "analysis.json")
					return exportMsg{err, "Exported to analysis.json"}
				}
			}
		case "m":
			if m.showExport {
				return m, func() tea.Msg {
					err := ExportMarkdown(m.data, "analysis.md")
					return exportMsg{err, "Exported to analysis.md"}
				}
			}
		}
	}
	return m, nil
}

func (m DashboardModel) View() string {

	// ❌ Error state
	if m.data.Repo == nil && m.statusMsg != "" {
		return errorStateView(m.statusMsg)
	}

	// 📭 Empty state
	if m.data.Repo == nil ||
		(len(m.data.Commits) == 0 &&
			len(m.data.Contributors) == 0 &&
			len(m.data.Languages) == 0) {
		return emptyStateView()
	}

	// Header
	header := TitleStyle.Render(fmt.Sprintf("Analysis for %s", m.data.Repo.FullName))

	// Metrics
	metrics := fmt.Sprintf(
		"Health Score: %d\nBus Factor: %d (%s)\nMaturity: %s (%d)",
		m.data.HealthScore,
		m.data.BusFactor, m.data.BusRisk,
		m.data.MaturityLevel, m.data.MaturityScore,
	)
	metricsBox := BoxStyle.Render(metrics)

	// Charts
	activityData := analyzer.CommitsPerDay(m.data.Commits)
	chart := RenderCommitActivity(activityData, 10)
	chartBox := BoxStyle.Render(chart)

	// File Tree
	treeContent := "📂 Files (Top 10):\n"
	limit := 10
	if len(m.data.FileTree) < limit {
		limit = len(m.data.FileTree)
	}
	for i := 0; i < limit; i++ {
		icon := "📄"
		if m.data.FileTree[i].Type == "tree" {
			icon = "📁"
		}
		treeContent += fmt.Sprintf("%s %s\n", icon, m.data.FileTree[i].Path)
	}
	if len(m.data.FileTree) > limit {
		treeContent += fmt.Sprintf("... and %d more", len(m.data.FileTree)-limit)
	}
	treeBox := BoxStyle.Render(treeContent)

	// Layout
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, metricsBox, chartBox)
	content := lipgloss.JoinVertical(lipgloss.Left, header, row1, treeBox)

	if m.showExport {
		exportMenu := BoxStyle.Render("Export Options:\n[J] JSON\n[M] Markdown")
		content = lipgloss.JoinVertical(lipgloss.Left, content, exportMenu)
	}

	if m.statusMsg != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMsg),
		)
	}

	content += "\n" + SubtleStyle.Render("e: export • q: back")

	if m.width == 0 {
		return content
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

/* ---------- Empty & Error Views ---------- */

func emptyStateView() string {
	return lipgloss.Place(
		60, 10,
		lipgloss.Center, lipgloss.Center,
		BoxStyle.Render(
			"📭 No analysis data available\n\n" +
				"This repository does not have enough data to generate analysis.\n" +
				"Try another repository or come back later.",
		),
	)
}

func errorStateView(msg string) string {
	return lipgloss.Place(
		60, 10,
		lipgloss.Center, lipgloss.Center,
		BoxStyle.Render(
			"❌ Analysis failed\n\n"+msg+"\n\nPress q to return.",
		),
	)
}
