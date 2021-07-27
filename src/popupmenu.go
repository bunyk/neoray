package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/sqweek/dialog"
)

var ButtonNames = []string{
	"Cut", "Copy", "Paste", "Select All", "Open File",
}

var MenuButtonEvents = map[string]func(){
	ButtonNames[0]: func() { //cut
		text := EditorSingleton.nvim.cutSelected()
		if text != "" {
			glfw.SetClipboardString(text)
		}
	},
	ButtonNames[1]: func() { //copy
		text := EditorSingleton.nvim.copySelected()
		if text != "" {
			glfw.SetClipboardString(text)
		}
	},
	ButtonNames[2]: func() { //paste
		EditorSingleton.nvim.paste(glfw.GetClipboardString())
	},
	ButtonNames[3]: func() { //select all
		EditorSingleton.nvim.selectAll()
	},
	ButtonNames[4]: func() { //open file
		filename, err := dialog.File().Load()
		if err == nil && filename != "" && filename != " " {
			EditorSingleton.nvim.openFile(filename)
		}
	},
}

type PopupMenu struct {
	pos        IntVec2
	vertexData VertexDataStorage
	hidden     bool
	width      int
	height     int
	cells      [][]rune
}

func CreatePopupMenu() PopupMenu {
	pmenu := PopupMenu{
		hidden: true,
	}
	// Find the longest text.
	longest := 0
	for _, name := range ButtonNames {
		if len(name) > longest {
			longest = len(name)
		}
	}
	// Create cells
	pmenu.width = longest + 2
	pmenu.height = len(ButtonNames)
	pmenu.cells = make([][]rune, pmenu.height, pmenu.height)
	for i := range pmenu.cells {
		pmenu.cells[i] = make([]rune, pmenu.width, pmenu.width)
	}
	pmenu.createCells()
	return pmenu
}

// Only call this function at initializing.
func (pmenu *PopupMenu) createCells() {
	// Loop through all cells and give them correct characters
	for x, row := range pmenu.cells {
		for y := range row {
			var c rune = 0
			if y != 0 && y != pmenu.width-1 {
				if y-1 < len(ButtonNames[x]) {
					c = rune(ButtonNames[x][y-1])
					if c == ' ' {
						c = 0
					}
				}
			}
			pmenu.cells[x][y] = c
		}
	}
}

func (pmenu *PopupMenu) createVertexData() {
	pmenu.vertexData = EditorSingleton.renderer.reserveVertexData(pmenu.width * pmenu.height)
	pmenu.updateChars()
}

func (pmenu *PopupMenu) updateChars() {
	for x, row := range pmenu.cells {
		for y, char := range row {
			cell_id := x*pmenu.width + y
			var atlasPos IntRect
			if char != 0 {
				atlasPos = EditorSingleton.renderer.getCharPos(
					char, false, false, false, false)
				// For multiwidth character.
				if atlasPos.W > EditorSingleton.cellWidth {
					atlasPos.W /= 2
				}
			}
			pmenu.vertexData.setCellTex(cell_id, atlasPos)
		}
	}
}

func (pmenu *PopupMenu) ShowAt(pos IntVec2) {
	pmenu.pos = pos
	fg := EditorSingleton.grid.defaultBg
	bg := EditorSingleton.grid.defaultFg
	for x, row := range pmenu.cells {
		for y := range row {
			cell_id := x*pmenu.width + y
			rect := F32Rect{
				X: float32(pos.X + y*EditorSingleton.cellWidth),
				Y: float32(pos.Y + x*EditorSingleton.cellHeight),
				W: float32(EditorSingleton.cellWidth),
				H: float32(EditorSingleton.cellHeight),
			}
			pmenu.vertexData.setCellPos(cell_id, rect)
			pmenu.vertexData.setCellFg(cell_id, fg)
			pmenu.vertexData.setCellBg(cell_id, bg)
		}
	}
	pmenu.hidden = false
	EditorSingleton.render()
}

func (pmenu *PopupMenu) Hide() {
	for x, row := range pmenu.cells {
		for y := range row {
			cell_id := x*pmenu.width + y
			pmenu.vertexData.setCellPos(cell_id, F32Rect{})
		}
	}
	pmenu.hidden = true
	EditorSingleton.render()
}

func (pmenu *PopupMenu) globalRect() IntRect {
	return IntRect{
		X: pmenu.pos.X,
		Y: pmenu.pos.Y,
		W: pmenu.width * EditorSingleton.cellWidth,
		H: pmenu.height * EditorSingleton.cellHeight,
	}
}

// Returns true if given position intersects with menu,
// and if the position is on the button, returns button index.
func (pmenu *PopupMenu) intersects(pos IntVec2) (bool, int) {
	menuRect := pmenu.globalRect()
	if pos.X >= menuRect.X && pos.Y >= menuRect.Y &&
		pos.X < menuRect.X+menuRect.W && pos.Y < menuRect.Y+menuRect.H {
		// Areas are intersecting. Now we need to find button under the cursor.
		// This is very simple. First we find the cell at the position.
		relativePos := IntVec2{
			X: pos.X - pmenu.pos.X,
			Y: pos.Y - pmenu.pos.Y,
		}
		row := relativePos.Y / EditorSingleton.cellHeight
		col := relativePos.X / EditorSingleton.cellWidth
		if col > 0 && col < pmenu.width-1 {
			return true, row
		}
		return true, -1
	}
	return false, -1
}

// Call this function when mouse moved.
func (pmenu *PopupMenu) mouseMove(pos IntVec2) {
	if !pmenu.hidden {
		ok, index := pmenu.intersects(pos)
		if ok {
			// Fill all cells with default colors.
			for i := 0; i < pmenu.width*pmenu.height; i++ {
				pmenu.vertexData.setCellFg(i, EditorSingleton.grid.defaultBg)
				pmenu.vertexData.setCellBg(i, EditorSingleton.grid.defaultFg)
			}
			if index != -1 {
				row := index
				if row < len(pmenu.cells) {
					// Highlight this row.
					for col := 1; col < pmenu.width-1; col++ {
						cell_id := row*pmenu.width + col
						pmenu.vertexData.setCellFg(cell_id, EditorSingleton.grid.defaultFg)
						pmenu.vertexData.setCellBg(cell_id, EditorSingleton.grid.defaultBg)
					}
				}
				EditorSingleton.render()
			}
		} else {
			// If this uncommented, the context menu will be hidden
			// when cursor goes out from on top of it.
			// pmenu.Hide()
		}
	}
}

// Call this function when mouse clicked.
// If rightbutton is false (left button is pressed) and positions are
// intersecting, this function returns true. This means if this function
// returns true than you shouldn't send button event to neovim.
func (pmenu *PopupMenu) mouseClick(rightbutton bool, pos IntVec2) bool {
	if !rightbutton && !pmenu.hidden {
		// If positions intersects than call button click event, hide popup menu otherwise.
		ok, index := pmenu.intersects(pos)
		if ok {
			if index != -1 {
				MenuButtonEvents[ButtonNames[index]]()
				pmenu.Hide()
			}
			return true
		} else {
			pmenu.Hide()
		}
	} else if rightbutton {
		// Open popup menu on this position
		pmenu.ShowAt(pos)
	}
	return false
}