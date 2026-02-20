package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"my-go-vibe/internal/domain"
	"my-go-vibe/internal/processor"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	app fyne.App
	win fyne.Window

	inputFileList  *widget.List
	outputFileList *widget.List
	jsonOutput     *widget.Entry
	patchInput     *widget.Entry

	inputFiles  []string
	outputFiles []string

	selectedInputFile  string
	selectedOutputFile string

	rootPath string
}

func newGUI() *GUI {
	a := app.New()
	w := a.NewWindow("my-go-vibe GUI")

	root, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("❌ Error getting absolute path: %v", err)
	}

	g := &GUI{
		app:      a,
		win:      w,
		rootPath: root,
	}

	g.setupUI()

	w.SetContent(g.createContent())
	w.Resize(fyne.NewSize(1200, 800))

	return g
}

func (g *GUI) run() {
	g.win.ShowAndRun()
}

func StartCustomGUI() {
	g := newGUI()
	g.run()
}

func (g *GUI) setupUI() {
	// Input files
	g.inputFileList = widget.NewList(
		func() int {
			return len(g.inputFiles)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(g.inputFiles[i])
		},
	)
	g.inputFileList.OnSelected = func(id widget.ListItemID) {
		g.selectedInputFile = g.inputFiles[id]
	}

	// Output files
	g.outputFileList = widget.NewList(
		func() int {
			return len(g.outputFiles)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(g.outputFiles[i])
		},
	)
	g.outputFileList.OnSelected = func(id widget.ListItemID) {
		g.selectedOutputFile = g.outputFiles[id]
	}

	g.jsonOutput = widget.NewMultiLineEntry()
	g.jsonOutput.Wrapping = fyne.TextWrapWord
	g.jsonOutput.Disable()

	g.patchInput = widget.NewMultiLineEntry()
	g.patchInput.Wrapping = fyne.TextWrapWord
	g.patchInput.SetPlaceHolder("Paste Patch JSON here...")

	g.refreshInputFiles()
	g.refreshOutputFiles()
}

func (g *GUI) createContent() fyne.CanvasObject {
	// Left Panel
	leftPanel := container.NewBorder(
		nil,
		widget.NewButton("Refresh", g.refreshInputFiles),
		nil,
		nil,
		container.NewScroll(g.inputFileList),
	)

	// Center Panel
	centerPanel := container.NewBorder(
		container.NewGridWithColumns(2,
			widget.NewButton("Scan File", g.scanFile),
			widget.NewButton("Copy JSON", func() {
				g.win.Clipboard().SetContent(g.jsonOutput.Text)
			}),
		),
		nil,
		nil,
		nil,
		container.NewScroll(g.jsonOutput),
	)

	// Right Panel
	rightPanel := container.NewBorder(
		widget.NewButton("Apply Patch", g.applyPatch),
		nil,
		nil,
		nil,
		container.NewScroll(g.patchInput),
	)

	// Bottom Panel
	bottomPanel := container.NewBorder(
		nil,
		container.NewGridWithColumns(2,
			widget.NewButton("Open Input Folder", g.openInputFolder),
			widget.NewButton("Open Output Folder", g.openOutputFolder),
		),
		nil,
		nil,
		container.NewScroll(g.outputFileList),
	)

	split := container.NewHSplit(
		leftPanel,
		container.NewHSplit(
			centerPanel,
			rightPanel,
		),
	)
	split.Offset = 0.2

	return container.NewBorder(nil, bottomPanel, nil, nil, split)
}

func (g *GUI) refreshInputFiles() {
	inputDir := filepath.Join(g.rootPath, "data", "input")
	g.inputFiles = g.findFiles(inputDir, ".docx")
	g.inputFileList.Refresh()
}

func (g *GUI) refreshOutputFiles() {
	outputDir := filepath.Join(g.rootPath, "data", "output")
	g.outputFiles = g.findFiles(outputDir, ".docx")
	g.outputFileList.Refresh()
}

func (g *GUI) findFiles(dir, suffix string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), suffix) {
			// We only want the base name, not the full path
			files = append(files, info.Name())
		}
		return nil
	})
	if err != nil {
		if !os.IsNotExist(err) { // Don't show error if the directory just doesn't exist yet
			dialog.ShowError(err, g.win)
		}
	}
	return files
}

func (g *GUI) scanFile() {
	if g.selectedInputFile == "" {
		dialog.ShowInformation("Information", "Please select a file to scan", g.win)
		return
	}

	filePath := filepath.Join(g.rootPath, "data", "input", g.selectedInputFile)
	content, err := processor.ExtractContent(filePath)
	if err != nil {
		dialog.ShowError(err, g.win)
		return
	}

	jsonBytes, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		dialog.ShowError(err, g.win)
		return
	}

	g.jsonOutput.SetText(string(jsonBytes))
	g.jsonOutput.Enable()
}

func (g *GUI) applyPatch() {
	if g.selectedInputFile == "" {
		dialog.ShowInformation("Information", "Please select a file to patch", g.win)
		return
	}
	if g.patchInput.Text == "" {
		dialog.ShowInformation("Information", "Please paste the patch JSON", g.win)
		return
	}

	var patch domain.PatchResponse
	err := json.Unmarshal([]byte(g.patchInput.Text), &patch)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid patch JSON: %w", err), g.win)
		return
	}

	inputPath := filepath.Join(g.rootPath, "data", "input", g.selectedInputFile)
	outputFile := strings.TrimSuffix(g.selectedInputFile, ".docx") + "_patched.docx"
	outPath := filepath.Join(g.rootPath, "data", "output", outputFile)

	if err := processor.ApplyPatch(inputPath, patch, outPath); err != nil {
		dialog.ShowError(err, g.win)
		return
	}

	g.refreshOutputFiles()
	dialog.ShowInformation("Success", fmt.Sprintf("Patched file saved to %s", outPath), g.win)
}

func (g *GUI) openInputFolder() {
	path := filepath.Join(g.rootPath, "data", "input")
	cmd := exec.Command("explorer", path)
	if err := cmd.Run(); err != nil {
		dialog.ShowError(err, g.win)
	}
}

func (g *GUI) openOutputFolder() {
	path := filepath.Join(g.rootPath, "data", "output")
	cmd := exec.Command("explorer", path)
	if err := cmd.Run(); err != nil {
		dialog.ShowError(err, g.win)
	}
}
