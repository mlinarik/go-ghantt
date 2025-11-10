package main

import (
	"bytes"
	"fmt"
	"strings"
)

type svgLayoutConfig struct {
	headerHeight           int
	baseRowHeight          int
	quarterWidth           int
	labelWidth             int
	padding                int
	categoryHeaderHeight   int
	titleLineHeight        int
	descLineHeight         int
	verticalPaddingPerTask int
}

type svgRenderContext struct {
	buf                *bytes.Buffer
	quarters           []quarterInfo
	totalQuarters      int
	config             svgLayoutConfig
	perTaskHeights     map[string]int
	perCategoryHeights map[string]int
}

// GenerateSVG creates an SVG representation of the Gantt chart
func GenerateSVG(chart *Chart) (string, error) {
	quarters := calculateQuarters(chart.StartYear, chart.StartQ, chart.EndYear, chart.EndQ)

	config := svgLayoutConfig{
		headerHeight:           80,
		baseRowHeight:          40,
		quarterWidth:           120,
		labelWidth:             200,
		padding:                20,
		categoryHeaderHeight:   35,
		titleLineHeight:        14,
		descLineHeight:         12,
		verticalPaddingPerTask: 8,
	}

	perTaskHeights, perCategoryHeights := calculateDynamicHeights(chart, config)

	totalCategoryHeadersHeight := sumCategoryHeights(chart, perCategoryHeights)
	totalTaskHeight := sumTaskHeights(chart, perTaskHeights)

	width := config.labelWidth + len(quarters)*config.quarterWidth + config.padding*2
	height := config.headerHeight + totalCategoryHeadersHeight + totalTaskHeight + config.padding*2

	var buf bytes.Buffer
	ctx := &svgRenderContext{
		buf:                &buf,
		quarters:           quarters,
		totalQuarters:      len(quarters),
		config:             config,
		perTaskHeights:     perTaskHeights,
		perCategoryHeights: perCategoryHeights,
	}

	ctx.writeSVGHeader(width, height)
	ctx.writeBackground(width, height)
	ctx.writeTitle(chart.Title)
	ctx.writeQuarterHeaders(height)
	ctx.writeCategoriesAndTasks(chart)
	buf.WriteString(`</svg>`)

	return buf.String(), nil
}

func calculateDynamicHeights(chart *Chart, config svgLayoutConfig) (map[string]int, map[string]int) {
	perTaskHeights := make(map[string]int)
	perCategoryHeights := make(map[string]int)

	for _, cat := range chart.Categories {
		catNameLines := wrapText(cat.Name, 30)
		catH := config.categoryHeaderHeight
		if len(catNameLines) > 1 {
			catH = 18 + len(catNameLines)*14
		}
		perCategoryHeights[cat.ID] = catH

		for _, task := range cat.Tasks {
			titleLines := wrapText(task.Title, 28)
			descLines := wrapText(task.Description, 36)
			h := config.baseRowHeight
			calc := len(titleLines)*config.titleLineHeight + len(descLines)*config.descLineHeight + config.verticalPaddingPerTask
			if calc > h {
				h = calc
			}
			perTaskHeights[task.ID] = h
		}
	}

	return perTaskHeights, perCategoryHeights
}

func sumCategoryHeights(chart *Chart, perCategoryHeights map[string]int) int {
	total := 0
	for _, cat := range chart.Categories {
		if h, ok := perCategoryHeights[cat.ID]; ok {
			total += h
		}
	}
	return total
}

func sumTaskHeights(chart *Chart, perTaskHeights map[string]int) int {
	total := 0
	for _, cat := range chart.Categories {
		for _, task := range cat.Tasks {
			if h, ok := perTaskHeights[task.ID]; ok {
				total += h
			}
		}
	}
	return total
}

func (ctx *svgRenderContext) writeSVGHeader(width, height int) {
	ctx.buf.WriteString(fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height))
	ctx.buf.WriteString(`<defs><style>.title{font:bold 20px sans-serif;fill:#333}.header{font:bold 12px sans-serif;fill:#555}.label{font:12px sans-serif;fill:#333}.category{font:bold 14px sans-serif;fill:#222}.desc{font:10px sans-serif;fill:#666}</style></defs>`)
}

func (ctx *svgRenderContext) writeBackground(width, height int) {
	ctx.buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#fafafa"/>`, width, height))
}

func (ctx *svgRenderContext) writeTitle(title string) {
	ctx.buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="title">%s</text>`,
		ctx.config.padding, ctx.config.padding+20, escapeXML(title)))
}

func (ctx *svgRenderContext) writeQuarterHeaders(height int) {
	for i, q := range ctx.quarters {
		x := ctx.config.padding + ctx.config.labelWidth + i*ctx.config.quarterWidth
		y := ctx.config.headerHeight

		color := "#e8e8e8"
		if i%2 == 1 {
			color = "#f5f5f5"
		}
		ctx.buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" stroke="#ccc" stroke-width="1"/>`,
			x, y-30, ctx.config.quarterWidth, 30, color))

		ctx.buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="header" text-anchor="middle">Q%d %d</text>`,
			x+ctx.config.quarterWidth/2, y-10, q.quarter, q.year))

		ctx.buf.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ddd" stroke-width="1"/>`,
			x, y, x, height-ctx.config.padding))
	}
}

func (ctx *svgRenderContext) writeCategoriesAndTasks(chart *Chart) {
	currentY := ctx.config.headerHeight
	for _, cat := range chart.Categories {
		currentY = ctx.writeCategory(cat, currentY)
	}
}

func (ctx *svgRenderContext) writeCategory(cat Category, currentY int) int {
	catH := ctx.perCategoryHeights[cat.ID]
	if catH == 0 {
		catH = ctx.config.categoryHeaderHeight
	}

	ctx.buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="0.3"/>`,
		ctx.config.padding, currentY, ctx.config.labelWidth, catH, cat.Color))

	ctx.writeCategoryName(cat.Name, currentY, catH)

	ctx.buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" opacity="0.05"/>`,
		ctx.config.padding+ctx.config.labelWidth, currentY, ctx.totalQuarters*ctx.config.quarterWidth, catH, cat.Color))

	currentY += catH
	return ctx.writeTasks(cat.Tasks, cat.Color, currentY)
}

func (ctx *svgRenderContext) writeCategoryName(name string, currentY, catH int) {
	catLines := wrapText(name, 30)
	if len(catLines) <= 1 {
		ctx.buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="category">%s</text>`,
			ctx.config.padding+10, currentY+22, escapeXML(name)))
	} else {
		ctx.buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="category">`, ctx.config.padding+10, currentY+18))
		for i, ln := range catLines {
			dy := 4
			if i > 0 {
				dy = 14
			}
			ctx.buf.WriteString(fmt.Sprintf(`<tspan x="%d" dy="%d">%s</tspan>`, ctx.config.padding+10, dy, escapeXML(ln)))
		}
		ctx.buf.WriteString(`</text>`)
	}
}

func (ctx *svgRenderContext) writeTasks(tasks []Task, catColor string, currentY int) int {
	for _, task := range tasks {
		h := ctx.perTaskHeights[task.ID]
		if h == 0 {
			h = ctx.config.baseRowHeight
		}

		ctx.buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#fff" stroke="#ddd" stroke-width="1"/>`,
			ctx.config.padding, currentY, ctx.config.labelWidth, h))

		ctx.writeTaskText(task, currentY)
		ctx.writeTaskBar(task, catColor, currentY, h)

		currentY += h
	}
	return currentY
}

func (ctx *svgRenderContext) writeTaskText(task Task, currentY int) int {
	titleLines := wrapText(task.Title, 28)
	descLines := wrapText(task.Description, 36)
	textY := currentY + 14

	if len(titleLines) > 0 {
		ctx.buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="label">`, ctx.config.padding+10, textY))
		for i, ln := range titleLines {
			dy := 0
			if i > 0 {
				dy = ctx.config.titleLineHeight
			}
			ctx.buf.WriteString(fmt.Sprintf(`<tspan x="%d" dy="%d">%s</tspan>`, ctx.config.padding+10, dy, escapeXML(ln)))
		}
		ctx.buf.WriteString(`</text>`)
		textY += len(titleLines) * ctx.config.titleLineHeight
	}

	if len(descLines) > 0 {
		ctx.buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="desc">`, ctx.config.padding+10, textY+4))
		for i, ln := range descLines {
			dy := 0
			if i > 0 {
				dy = ctx.config.descLineHeight
			}
			ctx.buf.WriteString(fmt.Sprintf(`<tspan x="%d" dy="%d">%s</tspan>`, ctx.config.padding+10, dy, escapeXML(ln)))
		}
		ctx.buf.WriteString(`</text>`)
	}

	return textY
}

func (ctx *svgRenderContext) writeTaskBar(task Task, catColor string, currentY, h int) {
	startIdx := findQuarterIndex(ctx.quarters, task.StartYear, task.StartQ)
	endIdx := findQuarterIndex(ctx.quarters, task.EndYear, task.EndQ)

	if startIdx < 0 || endIdx < 0 {
		return
	}

	barX := ctx.config.padding + ctx.config.labelWidth + startIdx*ctx.config.quarterWidth
	barWidth := (endIdx - startIdx + 1) * ctx.config.quarterWidth
	barY := currentY + 8
	barHeight := h - 16

	taskColor := task.Color
	if taskColor == "" {
		taskColor = catColor
	}

	if barHeight < 12 {
		barHeight = 12
	}

	ctx.buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="4" opacity="0.8"/>`,
		barX+2, barY, barWidth-4, barHeight, taskColor))
	ctx.buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="%s" stroke-width="2" rx="4"/>`,
		barX+2, barY, barWidth-4, barHeight, darken(taskColor)))
}

func wrapText(s string, maxChars int) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Fields(s)
	var lines []string
	line := ""
	for _, w := range parts {
		if len((line + " " + w)) <= maxChars {
			if line == "" {
				line = w
			} else {
				line = line + " " + w
			}
		} else {
			if line != "" {
				lines = append(lines, line)
			}
			line = w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

type quarterInfo struct {
	year    int
	quarter int
}

func calculateQuarters(startYear, startQ, endYear, endQ int) []quarterInfo {
	var quarters []quarterInfo

	for year := startYear; year <= endYear; year++ {
		startQuarter := 1
		endQuarter := 4

		if year == startYear {
			startQuarter = startQ
		}
		if year == endYear {
			endQuarter = endQ
		}

		for q := startQuarter; q <= endQuarter; q++ {
			quarters = append(quarters, quarterInfo{year: year, quarter: q})
		}
	}

	return quarters
}

func findQuarterIndex(quarters []quarterInfo, year, quarter int) int {
	for i, q := range quarters {
		if q.year == year && q.quarter == quarter {
			return i
		}
	}
	return -1
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func darken(color string) string {
	// Simple darkening - in production, you'd want proper color manipulation
	if color == "" {
		return "#333"
	}
	return color
}
