package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/indiekitai/mrr-cli/db"
	"github.com/indiekitai/mrr-cli/models"
)

// TUI represents the terminal UI
type TUI struct {
	screen    tcell.Screen
	entries   []models.Entry
	selected  int
	offset    int
	width     int
	height    int
	inputMode string
	inputBuf  string
	message   string
	msgStyle  tcell.Style
}

// Run starts the TUI
func Run() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return fmt.Errorf("failed to init screen: %w", err)
	}

	tui := &TUI{
		screen:   screen,
		msgStyle: tcell.StyleDefault,
	}

	defer screen.Fini()

	return tui.run()
}

func (t *TUI) run() error {
	t.refresh()
	t.render()

	for {
		t.screen.Show()

		ev := t.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			t.width, t.height = ev.Size()
			t.screen.Sync()
			t.render()

		case *tcell.EventKey:
			if t.inputMode != "" {
				if t.handleInput(ev) {
					t.inputMode = ""
					t.inputBuf = ""
					t.refresh()
				}
				t.render()
				continue
			}

			switch ev.Key() {
			case tcell.KeyEscape:
				return nil
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					return nil
				case 'j':
					t.moveDown()
				case 'k':
					t.moveUp()
				case 'g':
					t.selected = 0
					t.offset = 0
				case 'G':
					t.selected = len(t.entries) - 1
					t.adjustOffset()
				case 'a':
					t.inputMode = "add"
					t.inputBuf = ""
					t.message = "Add entry (amount): "
				case 'e':
					if len(t.entries) > 0 {
						t.inputMode = "edit"
						t.inputBuf = ""
						t.message = "Edit amount: "
					}
				case 'd':
					if len(t.entries) > 0 {
						t.inputMode = "delete"
						t.message = "Delete this entry? (y/n): "
					}
				case 'r':
					t.refresh()
					t.setMessage("Refreshed", tcell.StyleDefault.Foreground(tcell.ColorGreen))
				}
			}
			t.render()
		}
	}
}

func (t *TUI) handleInput(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		t.message = ""
		return true
	case tcell.KeyEnter:
		return t.submitInput()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(t.inputBuf) > 0 {
			t.inputBuf = t.inputBuf[:len(t.inputBuf)-1]
		}
	case tcell.KeyRune:
		t.inputBuf += string(ev.Rune())
	}
	return false
}

func (t *TUI) submitInput() bool {
	switch t.inputMode {
	case "add":
		amountStr := strings.TrimPrefix(t.inputBuf, "$")
		amountFloat, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			t.setMessage("Invalid amount", tcell.StyleDefault.Foreground(tcell.ColorRed))
			return true
		}
		amountCents := int64(amountFloat * 100)
		_, err = db.AddEntry(amountCents, "manual", "recurring", "", time.Now())
		if err != nil {
			t.setMessage("Error: "+err.Error(), tcell.StyleDefault.Foreground(tcell.ColorRed))
			return true
		}
		t.setMessage("Entry added!", tcell.StyleDefault.Foreground(tcell.ColorGreen))

	case "edit":
		if len(t.entries) == 0 {
			return true
		}
		amountStr := strings.TrimPrefix(t.inputBuf, "$")
		amountFloat, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			t.setMessage("Invalid amount", tcell.StyleDefault.Foreground(tcell.ColorRed))
			return true
		}
		amountCents := int64(amountFloat * 100)
		entry := t.entries[t.selected]
		err = db.UpdateEntry(entry.ID, &amountCents, nil, nil)
		if err != nil {
			t.setMessage("Error: "+err.Error(), tcell.StyleDefault.Foreground(tcell.ColorRed))
			return true
		}
		t.setMessage("Entry updated!", tcell.StyleDefault.Foreground(tcell.ColorGreen))

	case "delete":
		if strings.ToLower(t.inputBuf) == "y" && len(t.entries) > 0 {
			entry := t.entries[t.selected]
			err := db.DeleteEntry(entry.ID)
			if err != nil {
				t.setMessage("Error: "+err.Error(), tcell.StyleDefault.Foreground(tcell.ColorRed))
				return true
			}
			t.setMessage("Entry deleted!", tcell.StyleDefault.Foreground(tcell.ColorGreen))
			if t.selected > 0 {
				t.selected--
			}
		} else {
			t.message = ""
		}
	}
	return true
}

func (t *TUI) refresh() {
	entries, err := db.GetAllEntries()
	if err != nil {
		t.setMessage("Error loading entries: "+err.Error(), tcell.StyleDefault.Foreground(tcell.ColorRed))
		return
	}
	t.entries = entries
	if t.selected >= len(t.entries) && len(t.entries) > 0 {
		t.selected = len(t.entries) - 1
	}
}

func (t *TUI) moveDown() {
	if t.selected < len(t.entries)-1 {
		t.selected++
		t.adjustOffset()
	}
}

func (t *TUI) moveUp() {
	if t.selected > 0 {
		t.selected--
		t.adjustOffset()
	}
}

func (t *TUI) adjustOffset() {
	visibleRows := t.height - 6 // header + footer
	if visibleRows < 1 {
		visibleRows = 1
	}
	if t.selected < t.offset {
		t.offset = t.selected
	}
	if t.selected >= t.offset+visibleRows {
		t.offset = t.selected - visibleRows + 1
	}
}

func (t *TUI) setMessage(msg string, style tcell.Style) {
	t.message = msg
	t.msgStyle = style
}

func (t *TUI) render() {
	t.screen.Clear()
	t.width, t.height = t.screen.Size()

	// Styles
	headerStyle := tcell.StyleDefault.Bold(true).Foreground(tcell.ColorTeal)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	normalStyle := tcell.StyleDefault
	amountStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	sourceStyle := tcell.StyleDefault.Foreground(tcell.ColorPurple)
	helpStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	// Title
	title := "💰 MRR Tracker"
	t.drawString(2, 0, title, headerStyle)

	// Calculate total MRR
	var totalMRR int64
	for _, e := range t.entries {
		if e.Type == "recurring" {
			totalMRR += e.Amount
		}
	}
	mrrStr := fmt.Sprintf("MRR: %s", models.FormatAmount(totalMRR, "USD"))
	t.drawString(t.width-len(mrrStr)-2, 0, mrrStr, amountStyle.Bold(true))

	// Column headers
	y := 2
	t.drawString(2, y, "ID", headerStyle)
	t.drawString(8, y, "Date", headerStyle)
	t.drawString(20, y, "Amount", headerStyle)
	t.drawString(32, y, "Source", headerStyle)
	t.drawString(44, y, "Type", headerStyle)
	t.drawString(56, y, "Note", headerStyle)

	// Separator
	y++
	for x := 2; x < t.width-2; x++ {
		t.screen.SetContent(x, y, '─', nil, tcell.StyleDefault.Foreground(tcell.ColorGray))
	}

	// Entries
	visibleRows := t.height - 7
	if visibleRows < 1 {
		visibleRows = 1
	}

	for i := t.offset; i < len(t.entries) && i-t.offset < visibleRows; i++ {
		y++
		e := t.entries[i]

		style := normalStyle
		if i == t.selected {
			style = selectedStyle
			// Clear the row
			for x := 0; x < t.width; x++ {
				t.screen.SetContent(x, y, ' ', nil, selectedStyle)
			}
		}

		t.drawString(2, y, fmt.Sprintf("%d", e.ID), style)
		t.drawString(8, y, e.Date.Format("2006-01-02"), style)
		
		if i == t.selected {
			t.drawString(20, y, models.FormatAmount(e.Amount, "USD"), selectedStyle)
			t.drawString(32, y, e.Source, selectedStyle)
		} else {
			t.drawString(20, y, models.FormatAmount(e.Amount, "USD"), amountStyle)
			t.drawString(32, y, e.Source, sourceStyle)
		}
		
		typeStyle := style
		if e.Type == "one-time" && i != t.selected {
			typeStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
		}
		t.drawString(44, y, e.Type, typeStyle)

		note := e.Note
		maxNoteLen := t.width - 58
		if maxNoteLen < 0 {
			maxNoteLen = 0
		}
		if len(note) > maxNoteLen {
			note = note[:maxNoteLen-3] + "..."
		}
		t.drawString(56, y, note, style)
	}

	// Footer / Help
	footerY := t.height - 2
	if t.message != "" || t.inputMode != "" {
		msg := t.message
		if t.inputMode != "" {
			msg = t.message + t.inputBuf + "_"
		}
		t.drawString(2, footerY, msg, t.msgStyle)
	} else {
		help := "j/k: navigate | a: add | e: edit | d: delete | r: refresh | q: quit"
		t.drawString(2, footerY, help, helpStyle)
	}

	// Status bar
	statusY := t.height - 1
	status := fmt.Sprintf(" %d entries ", len(t.entries))
	if len(t.entries) > 0 {
		status += fmt.Sprintf("| %d/%d ", t.selected+1, len(t.entries))
	}
	t.drawString(0, statusY, status, tcell.StyleDefault.Background(tcell.ColorDarkGray).Foreground(tcell.ColorWhite))
}

func (t *TUI) drawString(x, y int, str string, style tcell.Style) {
	for i, r := range str {
		if x+i >= t.width {
			break
		}
		t.screen.SetContent(x+i, y, r, nil, style)
	}
}
