package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

type pngLayoutConfig struct {
	headerHeight         int
	rowHeight            int
	quarterWidth         int
	labelWidth           int
	padding              int
	categoryHeaderHeight int
}

type pngRenderContext struct {
	img           *image.RGBA
	quarters      []quarterInfo
	totalQuarters int
	config        pngLayoutConfig
}

// GeneratePNG creates a PNG image of the Gantt chart
func GeneratePNG(chart *Chart) ([]byte, error) {
	quarters := calculateQuarters(chart.StartYear, chart.StartQ, chart.EndYear, chart.EndQ)

	config := pngLayoutConfig{
		headerHeight:         80,
		rowHeight:            40,
		quarterWidth:         120,
		labelWidth:           200,
		padding:              20,
		categoryHeaderHeight: 35,
	}

	totalRows := countTotalRows(chart)
	width := config.labelWidth + len(quarters)*config.quarterWidth + config.padding*2
	height := config.headerHeight + totalRows*config.rowHeight + config.padding*2

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{250, 250, 250, 255}}, image.Point{}, draw.Src)

	ctx := &pngRenderContext{
		img:           img,
		quarters:      quarters,
		totalQuarters: len(quarters),
		config:        config,
	}

	ctx.drawQuarterHeaders(height)
	ctx.drawCategoriesAndTasks(chart)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func countTotalRows(chart *Chart) int {
	totalRows := 0
	for _, cat := range chart.Categories {
		totalRows++ // Category header
		totalRows += len(cat.Tasks)
	}
	return totalRows
}

func (ctx *pngRenderContext) drawQuarterHeaders(height int) {
	for i := range ctx.quarters {
		x := ctx.config.padding + ctx.config.labelWidth + i*ctx.config.quarterWidth
		y := ctx.config.headerHeight

		bgColor := color.RGBA{232, 232, 232, 255}
		if i%2 == 1 {
			bgColor = color.RGBA{245, 245, 245, 255}
		}
		drawRect(ctx.img, x, y-30, ctx.config.quarterWidth, 30, bgColor)
		drawRectBorder(ctx.img, x, y-30, ctx.config.quarterWidth, 30, color.RGBA{204, 204, 204, 255})
		drawVerticalLine(ctx.img, x, y, height-ctx.config.padding, color.RGBA{221, 221, 221, 255})
	}
}

func (ctx *pngRenderContext) drawCategoriesAndTasks(chart *Chart) {
	currentY := ctx.config.headerHeight
	for _, cat := range chart.Categories {
		catColor := parseColor(cat.Color)
		currentY = ctx.drawCategory(cat, catColor, currentY)
	}
}

func (ctx *pngRenderContext) drawCategory(cat Category, catColor color.RGBA, currentY int) int {
	catColorAlpha := color.RGBA{catColor.R, catColor.G, catColor.B, 76}
	drawRect(ctx.img, ctx.config.padding, currentY, ctx.config.labelWidth, ctx.config.categoryHeaderHeight, catColorAlpha)

	catColorLight := color.RGBA{catColor.R, catColor.G, catColor.B, 13}
	drawRect(ctx.img, ctx.config.padding+ctx.config.labelWidth, currentY, ctx.totalQuarters*ctx.config.quarterWidth, ctx.config.categoryHeaderHeight, catColorLight)

	currentY += ctx.config.categoryHeaderHeight
	return ctx.drawTasks(cat.Tasks, catColor, currentY)
}

func (ctx *pngRenderContext) drawTasks(tasks []Task, catColor color.RGBA, currentY int) int {
	for _, task := range tasks {
		drawRect(ctx.img, ctx.config.padding, currentY, ctx.config.labelWidth, ctx.config.rowHeight, color.RGBA{255, 255, 255, 255})
		drawRectBorder(ctx.img, ctx.config.padding, currentY, ctx.config.labelWidth, ctx.config.rowHeight, color.RGBA{221, 221, 221, 255})
		ctx.drawTaskBar(task, catColor, currentY)
		currentY += ctx.config.rowHeight
	}
	return currentY
}

func (ctx *pngRenderContext) drawTaskBar(task Task, catColor color.RGBA, currentY int) {
	startIdx := findQuarterIndex(ctx.quarters, task.StartYear, task.StartQ)
	endIdx := findQuarterIndex(ctx.quarters, task.EndYear, task.EndQ)

	if startIdx < 0 || endIdx < 0 {
		return
	}

	barX := ctx.config.padding + ctx.config.labelWidth + startIdx*ctx.config.quarterWidth
	barWidth := (endIdx - startIdx + 1) * ctx.config.quarterWidth
	barY := currentY + 8
	barHeight := ctx.config.rowHeight - 16

	taskColor := parseColor(task.Color)
	if task.Color == "" {
		taskColor = catColor
	}

	taskColorAlpha := color.RGBA{taskColor.R, taskColor.G, taskColor.B, 204}
	drawRoundedRect(ctx.img, barX+2, barY, barWidth-4, barHeight, 4, taskColorAlpha)
}

type pdfLayoutConfig struct {
	headerHeight         float64
	rowHeight            float64
	quarterWidth         float64
	labelWidth           float64
	padding              float64
	categoryHeaderHeight float64
}

type pdfRenderContext struct {
	pdf      *gofpdf.Fpdf
	quarters []quarterInfo
	config   pdfLayoutConfig
}

// GeneratePDF creates a PDF document of the Gantt chart
func GeneratePDF(chart *Chart) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	quarters := calculateQuarters(chart.StartYear, chart.StartQ, chart.EndYear, chart.EndQ)

	config := pdfLayoutConfig{
		headerHeight:         20.0,
		rowHeight:            10.0,
		quarterWidth:         25.0,
		labelWidth:           50.0,
		padding:              10.0,
		categoryHeaderHeight: 8.0,
	}

	ctx := &pdfRenderContext{
		pdf:      pdf,
		quarters: quarters,
		config:   config,
	}

	ctx.writeTitle(chart.Title)
	ctx.writeQuarterHeaders()
	ctx.writeCategoriesAndTasks(chart)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ctx *pdfRenderContext) writeTitle(title string) {
	ctx.pdf.SetFont("Arial", "B", 16)
	ctx.pdf.SetXY(ctx.config.padding, ctx.config.padding)
	ctx.pdf.Cell(0, 10, title)
}

func (ctx *pdfRenderContext) writeQuarterHeaders() {
	ctx.pdf.SetFont("Arial", "B", 9)
	for i, q := range ctx.quarters {
		x := ctx.config.padding + ctx.config.labelWidth + float64(i)*ctx.config.quarterWidth
		y := ctx.config.headerHeight

		if i%2 == 0 {
			ctx.pdf.SetFillColor(232, 232, 232)
		} else {
			ctx.pdf.SetFillColor(245, 245, 245)
		}
		ctx.pdf.Rect(x, y-8, ctx.config.quarterWidth, 8, "F")

		ctx.pdf.SetXY(x, y-7)
		ctx.pdf.Cell(ctx.config.quarterWidth, 6, fmt.Sprintf("Q%d %d", q.quarter, q.year))
	}
}

func (ctx *pdfRenderContext) writeCategoriesAndTasks(chart *Chart) {
	currentY := ctx.config.headerHeight
	ctx.pdf.SetFont("Arial", "", 8)

	for _, cat := range chart.Categories {
		currentY = ctx.writeCategory(cat, currentY)
	}
}

func (ctx *pdfRenderContext) writeCategory(cat Category, currentY float64) float64 {
	catR, catG, catB := parseColorRGB(cat.Color)
	ctx.pdf.SetFillColor(catR, catG, catB)
	ctx.pdf.SetAlpha(0.3, "Normal")
	ctx.pdf.Rect(ctx.config.padding, currentY, ctx.config.labelWidth, ctx.config.categoryHeaderHeight, "F")
	ctx.pdf.SetAlpha(1.0, "Normal")

	ctx.pdf.SetFont("Arial", "B", 9)
	ctx.pdf.SetXY(ctx.config.padding+2, currentY+2)
	ctx.pdf.Cell(ctx.config.labelWidth-4, ctx.config.categoryHeaderHeight-4, cat.Name)
	ctx.pdf.SetFont("Arial", "", 8)

	currentY += ctx.config.categoryHeaderHeight

	for _, task := range cat.Tasks {
		currentY = ctx.writeTask(task, cat.Color, currentY)
	}

	return currentY
}

func (ctx *pdfRenderContext) writeTask(task Task, catColor string, currentY float64) float64 {
	ctx.pdf.SetDrawColor(221, 221, 221)
	ctx.pdf.Rect(ctx.config.padding, currentY, ctx.config.labelWidth, ctx.config.rowHeight, "D")

	ctx.pdf.SetXY(ctx.config.padding+2, currentY+2)
	ctx.pdf.Cell(ctx.config.labelWidth-4, 4, truncate(task.Title, 20))

	ctx.writeTaskBar(task, catColor, currentY)

	return currentY + ctx.config.rowHeight
}

func (ctx *pdfRenderContext) writeTaskBar(task Task, catColor string, currentY float64) {
	startIdx := findQuarterIndex(ctx.quarters, task.StartYear, task.StartQ)
	endIdx := findQuarterIndex(ctx.quarters, task.EndYear, task.EndQ)

	if startIdx < 0 || endIdx < 0 {
		return
	}

	barX := ctx.config.padding + ctx.config.labelWidth + float64(startIdx)*ctx.config.quarterWidth
	barWidth := float64(endIdx-startIdx+1) * ctx.config.quarterWidth
	barY := currentY + 2
	barHeight := ctx.config.rowHeight - 4

	taskColor := task.Color
	if taskColor == "" {
		taskColor = catColor
	}

	taskR, taskG, taskB := parseColorRGB(taskColor)
	ctx.pdf.SetFillColor(taskR, taskG, taskB)
	ctx.pdf.SetAlpha(0.8, "Normal")
	ctx.pdf.Rect(barX+1, barY, barWidth-2, barHeight, "F")
	ctx.pdf.SetAlpha(1.0, "Normal")
}

// Helper functions for image drawing
func drawRect(img *image.RGBA, x, y, width, height int, col color.RGBA) {
	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			if i >= 0 && i < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
				img.Set(i, j, col)
			}
		}
	}
}

func drawRectBorder(img *image.RGBA, x, y, width, height int, col color.RGBA) {
	drawHorizontalBorders(img, x, y, width, height, col)
	drawVerticalBorders(img, x, y, width, height, col)
}

func drawHorizontalBorders(img *image.RGBA, x, y, width, height int, col color.RGBA) {
	bounds := img.Bounds()
	for i := x; i < x+width; i++ {
		if i >= 0 && i < bounds.Dx() {
			if y >= 0 && y < bounds.Dy() {
				img.Set(i, y, col)
			}
			bottomY := y + height
			if bottomY >= 0 && bottomY < bounds.Dy() {
				img.Set(i, bottomY, col)
			}
		}
	}
}

func drawVerticalBorders(img *image.RGBA, x, y, width, height int, col color.RGBA) {
	bounds := img.Bounds()
	for j := y; j < y+height; j++ {
		if j >= 0 && j < bounds.Dy() {
			if x >= 0 && x < bounds.Dx() {
				img.Set(x, j, col)
			}
			rightX := x + width
			if rightX >= 0 && rightX < bounds.Dx() {
				img.Set(rightX, j, col)
			}
		}
	}
}

func drawRoundedRect(img *image.RGBA, x, y, width, height, radius int, col color.RGBA) {
	// Simplified rounded rectangle - just draw a regular rectangle for now
	drawRect(img, x, y, width, height, col)
}

func drawVerticalLine(img *image.RGBA, x, y1, y2 int, col color.RGBA) {
	for j := y1; j <= y2; j++ {
		if x >= 0 && x < img.Bounds().Dx() && j >= 0 && j < img.Bounds().Dy() {
			img.Set(x, j, col)
		}
	}
}

func parseColor(hexColor string) color.RGBA {
	if hexColor == "" {
		return color.RGBA{100, 149, 237, 255} // Default blue
	}

	hexColor = strings.TrimPrefix(hexColor, "#")
	if len(hexColor) != 6 {
		return color.RGBA{100, 149, 237, 255}
	}

	r, _ := strconv.ParseUint(hexColor[0:2], 16, 8)
	g, _ := strconv.ParseUint(hexColor[2:4], 16, 8)
	b, _ := strconv.ParseUint(hexColor[4:6], 16, 8)

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func parseColorRGB(hexColor string) (int, int, int) {
	col := parseColor(hexColor)
	return int(col.R), int(col.G), int(col.B)
}
