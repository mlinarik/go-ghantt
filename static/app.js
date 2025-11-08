// State management
let currentChart = null;
let currentCategoryId = null;
let currentTaskId = null;
let editingCategory = null;
let editingTask = null;

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    createNewChart();
});

function setupEventListeners() {
    // Header actions
    document.getElementById('newChartBtn').addEventListener('click', createNewChart);
    document.getElementById('loadChartBtn').addEventListener('click', openLoadChartModal);
    document.getElementById('saveChartBtn').addEventListener('click', openSaveChartModal);
    
    // Export buttons
    document.getElementById('exportSVG').addEventListener('click', () => exportChart('svg'));
    document.getElementById('exportPNG').addEventListener('click', () => exportChart('png'));
    document.getElementById('exportPDF').addEventListener('click', () => exportChart('pdf'));
    
    // Chart settings
    document.getElementById('chartTitle').addEventListener('input', updateChartSettings);
    document.getElementById('startYear').addEventListener('change', updateChartSettings);
    document.getElementById('startQuarter').addEventListener('change', updateChartSettings);
    document.getElementById('endYear').addEventListener('change', updateChartSettings);
    document.getElementById('endQuarter').addEventListener('change', updateChartSettings);
    
    // Category modal
    document.getElementById('addCategoryBtn').addEventListener('click', () => openCategoryModal());
    document.getElementById('saveCategory').addEventListener('click', saveCategory);
    document.getElementById('cancelCategory').addEventListener('click', closeCategoryModal);
    
    // Task modal
    document.getElementById('saveTask').addEventListener('click', saveTask);
    document.getElementById('cancelTask').addEventListener('click', closeTaskModal);
    
    // Load chart modal
    document.getElementById('closeLoadChart').addEventListener('click', closeLoadChartModal);
    document.getElementById('cancelLoadChart').addEventListener('click', closeLoadChartModal);
    
    // Save chart modal
    document.getElementById('closeSaveChart').addEventListener('click', closeSaveChartModal);
    document.getElementById('cancelSaveChart').addEventListener('click', closeSaveChartModal);
    document.getElementById('confirmSaveChart').addEventListener('click', confirmSaveChart);
    document.getElementById('existingChartSelect').addEventListener('change', (e) => {
        if (e.target.value) {
            document.getElementById('saveChartName').value = '';
            document.getElementById('saveChartName').disabled = true;
        } else {
            document.getElementById('saveChartName').disabled = false;
        }
    });
    
    // Close modals when clicking X or outside
    document.querySelectorAll('.modal .close').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.target.closest('.modal').classList.remove('active');
        });
    });
    
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
            }
        });
    });
}

function createNewChart() {
    const now = new Date();
    const year = now.getFullYear();
    const quarter = Math.ceil((now.getMonth() + 1) / 3);
    
    currentChart = {
        id: generateId(),
        title: 'New Gantt Chart',
        startYear: year,
        startQuarter: quarter,
        endYear: year + 1,
        endQuarter: 4,
        categories: []
    };
    
    updateUI();
}

function updateUI() {
    if (!currentChart) return;
    
    // Update form fields
    document.getElementById('chartTitle').value = currentChart.title;
    document.getElementById('startYear').value = currentChart.startYear;
    document.getElementById('startQuarter').value = currentChart.startQuarter;
    document.getElementById('endYear').value = currentChart.endYear;
    document.getElementById('endQuarter').value = currentChart.endQuarter;
    
    // Render categories
    renderCategories();
    
    // Update preview
    updatePreview();
}

function updateUIFromChart() {
    updateUI();
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function renderCategories() {
    const container = document.getElementById('categoriesList');
    container.innerHTML = '';
    
    if (!currentChart.categories || currentChart.categories.length === 0) {
        container.innerHTML = '<p style="color: #95a5a6; font-size: 0.875rem;">No categories yet</p>';
        return;
    }
    
    currentChart.categories.forEach(category => {
        const categoryEl = document.createElement('div');
        categoryEl.className = 'category-item';
        categoryEl.innerHTML = `
            <div class="category-header">
                <div class="category-title">
                    <div class="category-color" style="background: ${category.color}"></div>
                    <span>${escapeHtml(category.name)}</span>
                </div>
                <div class="category-actions">
                    <button class="btn btn-small btn-primary" onclick="openTaskModal('${category.id}')">+ Task</button>
                    <button class="btn btn-small btn-secondary" onclick="editCategory('${category.id}')">Edit</button>
                    <button class="btn btn-small btn-danger" onclick="deleteCategory('${category.id}')">Delete</button>
                </div>
            </div>
            <div class="task-list" id="tasks-${category.id}">
                ${renderTasks(category)}
            </div>
        `;
        container.appendChild(categoryEl);
    });
}

function renderTasks(category) {
    if (!category.tasks || category.tasks.length === 0) {
        return '<p style="color: #bdc3c7; font-size: 0.75rem; margin-top: 0.5rem;">No tasks</p>';
    }
    
    return category.tasks.map(task => `
        <div class="task-item" draggable="true" ondragstart="dragStart(event, '${category.id}', '${task.id}')" ondragend="dragEnd(event)">
            <div class="task-header">
                <div>
                    <div class="task-title-text">${escapeHtml(task.title)}</div>
                    <div class="task-timeline">Q${task.startQuarter} ${task.startYear} - Q${task.endQuarter} ${task.endYear}</div>
                    ${task.description ? `<div class="task-description">${escapeHtml(task.description)}</div>` : ''}
                </div>
                <div class="task-actions">
                    <button class="btn btn-small btn-secondary" onclick="editTask('${category.id}', '${task.id}')">Edit</button>
                    <button class="btn btn-small btn-danger" onclick="deleteTask('${category.id}', '${task.id}')">Del</button>
                </div>
            </div>
        </div>
    `).join('');
}

function updateChartSettings() {
    if (!currentChart) return;
    
    currentChart.title = document.getElementById('chartTitle').value;
    currentChart.startYear = parseInt(document.getElementById('startYear').value);
    currentChart.startQuarter = parseInt(document.getElementById('startQuarter').value);
    currentChart.endYear = parseInt(document.getElementById('endYear').value);
    currentChart.endQuarter = parseInt(document.getElementById('endQuarter').value);
    
    updatePreview();
}

async function updatePreview() {
    if (!currentChart) return;
    
    const preview = document.getElementById('chartPreview');
    
    if (!currentChart.categories || currentChart.categories.length === 0) {
        preview.innerHTML = `
            <div class="empty-state">
                <h2>Add categories and tasks to see your Gantt chart</h2>
                <p>Use the sidebar to build your chart</p>
            </div>
        `;
        return;
    }
    
    try {
        // For preview, we'll generate the SVG client-side
        const svg = generateClientSVG(currentChart);
        preview.innerHTML = svg;
    } catch (error) {
        console.error('Error generating preview:', error);
        preview.innerHTML = `<div class="empty-state"><h2>Error generating preview</h2></div>`;
    }
}

// Category management
function openCategoryModal(categoryId = null) {
    editingCategory = categoryId;
    const modal = document.getElementById('categoryModal');
    
    if (categoryId) {
        const category = currentChart.categories.find(c => c.id === categoryId);
        document.getElementById('categoryModalTitle').textContent = 'Edit Category';
        document.getElementById('categoryName').value = category.name;
        document.getElementById('categoryColor').value = category.color;
    } else {
        document.getElementById('categoryModalTitle').textContent = 'Add Category';
        document.getElementById('categoryName').value = '';
        document.getElementById('categoryColor').value = '#6495ed';
    }
    
    modal.classList.add('active');
}

function closeCategoryModal() {
    document.getElementById('categoryModal').classList.remove('active');
    editingCategory = null;
}

function saveCategory() {
    const name = document.getElementById('categoryName').value.trim();
    const color = document.getElementById('categoryColor').value;
    
    if (!name) {
        alert('Please enter a category name');
        return;
    }
    
    if (editingCategory) {
        const category = currentChart.categories.find(c => c.id === editingCategory);
        category.name = name;
        category.color = color;
    } else {
        currentChart.categories.push({
            id: generateId(),
            name,
            color,
            tasks: []
        });
    }
    
    closeCategoryModal();
    updateUI();
}

function editCategory(categoryId) {
    openCategoryModal(categoryId);
}

function deleteCategory(categoryId) {
    if (!confirm('Delete this category and all its tasks?')) return;
    
    currentChart.categories = currentChart.categories.filter(c => c.id !== categoryId);
    updateUI();
}

// Task management
function openTaskModal(categoryId, taskId = null) {
    currentCategoryId = categoryId;
    currentTaskId = taskId;
    const modal = document.getElementById('taskModal');
    
    if (taskId) {
        const category = currentChart.categories.find(c => c.id === categoryId);
        const task = category.tasks.find(t => t.id === taskId);
        
        document.getElementById('taskModalTitle').textContent = 'Edit Task';
        document.getElementById('taskTitle').value = task.title;
        document.getElementById('taskDescription').value = task.description || '';
        document.getElementById('taskStartYear').value = task.startYear;
        document.getElementById('taskStartQuarter').value = task.startQuarter;
        document.getElementById('taskEndYear').value = task.endYear;
        document.getElementById('taskEndQuarter').value = task.endQuarter;
        document.getElementById('taskColor').value = task.color || '';
    } else {
        document.getElementById('taskModalTitle').textContent = 'Add Task';
        document.getElementById('taskTitle').value = '';
        document.getElementById('taskDescription').value = '';
        document.getElementById('taskStartYear').value = currentChart.startYear;
        document.getElementById('taskStartQuarter').value = currentChart.startQuarter;
        document.getElementById('taskEndYear').value = currentChart.endYear;
        document.getElementById('taskEndQuarter').value = currentChart.endQuarter;
        document.getElementById('taskColor').value = '';
    }
    
    modal.classList.add('active');
}

function closeTaskModal() {
    document.getElementById('taskModal').classList.remove('active');
    currentCategoryId = null;
    currentTaskId = null;
}

function saveTask() {
    const title = document.getElementById('taskTitle').value.trim();
    const description = document.getElementById('taskDescription').value.trim();
    const startYear = parseInt(document.getElementById('taskStartYear').value);
    const startQuarter = parseInt(document.getElementById('taskStartQuarter').value);
    const endYear = parseInt(document.getElementById('taskEndYear').value);
    const endQuarter = parseInt(document.getElementById('taskEndQuarter').value);
    const color = document.getElementById('taskColor').value;
    
    if (!title) {
        alert('Please enter a task title');
        return;
    }
    
    const category = currentChart.categories.find(c => c.id === currentCategoryId);
    
    if (currentTaskId) {
        const task = category.tasks.find(t => t.id === currentTaskId);
        task.title = title;
        task.description = description;
        task.startYear = startYear;
        task.startQuarter = startQuarter;
        task.endYear = endYear;
        task.endQuarter = endQuarter;
        task.color = color;
    } else {
        category.tasks.push({
            id: generateId(),
            title,
            description,
            startYear,
            startQuarter,
            endYear,
            endQuarter,
            color
        });
    }
    
    closeTaskModal();
    updateUI();
}

function editTask(categoryId, taskId) {
    openTaskModal(categoryId, taskId);
}

function deleteTask(categoryId, taskId) {
    if (!confirm('Delete this task?')) return;
    
    const category = currentChart.categories.find(c => c.id === categoryId);
    category.tasks = category.tasks.filter(t => t.id !== taskId);
    updateUI();
}

// Drag and drop
let draggedTask = null;
let draggedCategory = null;

function dragStart(event, categoryId, taskId) {
    draggedCategory = categoryId;
    draggedTask = taskId;
    event.target.style.opacity = '0.5';
}

function dragEnd(event) {
    event.target.style.opacity = '1';
}

// Save and export
// Load Chart Modal
async function openLoadChartModal() {
    try {
        const response = await fetch('/api/charts');
        if (!response.ok) throw new Error('Failed to load charts');
        
        const charts = await response.json();
        const chartList = document.getElementById('chartList');
        chartList.innerHTML = '';
        
        if (charts.length === 0) {
            chartList.innerHTML = '<p style="text-align: center; color: #6c757d;">No saved charts found</p>';
        } else {
            charts.forEach(chart => {
                const item = document.createElement('div');
                item.className = 'chart-item';
                item.innerHTML = `
                    <div class="chart-item-info">
                        <div class="chart-item-name">${escapeHtml(chart.title || 'Untitled Chart')}</div>
                        <div class="chart-item-date">ID: ${chart.id}</div>
                    </div>
                    <div class="chart-item-actions">
                        <button class="btn btn-sm btn-primary" onclick="loadChart('${chart.id}')">Load</button>
                        <button class="btn btn-sm btn-danger" onclick="deleteChart('${chart.id}')">Delete</button>
                    </div>
                `;
                chartList.appendChild(item);
            });
        }
        
        document.getElementById('loadChartModal').classList.add('active');
    } catch (error) {
        console.error('Error loading charts:', error);
        alert('Error loading charts: ' + error.message);
    }
}

function closeLoadChartModal() {
    document.getElementById('loadChartModal').classList.remove('active');
}

async function loadChart(id) {
    try {
        const response = await fetch(`/api/charts/${id}`);
        if (!response.ok) throw new Error('Failed to load chart');
        
        currentChart = await response.json();
        updateUIFromChart();
        renderChart();
        closeLoadChartModal();
        alert('Chart loaded successfully!');
    } catch (error) {
        console.error('Error loading chart:', error);
        alert('Error loading chart: ' + error.message);
    }
}

async function deleteChart(id) {
    if (!confirm('Are you sure you want to delete this chart?')) {
        return;
    }
    
    try {
        const response = await fetch(`/api/charts/${id}`, { method: 'DELETE' });
        if (!response.ok) throw new Error('Failed to delete chart');
        
        alert('Chart deleted successfully!');
        openLoadChartModal(); // Refresh the list
    } catch (error) {
        console.error('Error deleting chart:', error);
        alert('Error deleting chart: ' + error.message);
    }
}

// Save Chart Modal
async function openSaveChartModal() {
    try {
        // Load existing charts for the dropdown
        const response = await fetch('/api/charts');
        if (response.ok) {
            const charts = await response.json();
            const select = document.getElementById('existingChartSelect');
            select.innerHTML = '<option value="">-- Save as new --</option>';
            charts.forEach(chart => {
                const option = document.createElement('option');
                option.value = chart.id;
                option.textContent = chart.title || 'Untitled Chart';
                select.appendChild(option);
            });
        }
        
        // Pre-fill with current chart title if exists
        document.getElementById('saveChartName').value = currentChart.title || '';
        document.getElementById('saveChartName').disabled = false;
        document.getElementById('existingChartSelect').value = currentChart.id || '';
        
        if (currentChart.id) {
            document.getElementById('existingChartSelect').value = currentChart.id;
            document.getElementById('saveChartName').disabled = true;
        }
        
        document.getElementById('saveChartModal').classList.add('active');
    } catch (error) {
        console.error('Error opening save modal:', error);
    }
}

function closeSaveChartModal() {
    document.getElementById('saveChartModal').classList.remove('active');
}

async function confirmSaveChart() {
    const existingId = document.getElementById('existingChartSelect').value;
    const newName = document.getElementById('saveChartName').value.trim();
    
    if (!existingId && !newName) {
        alert('Please enter a chart name or select an existing chart');
        return;
    }
    
    // Update chart title if saving as new
    if (!existingId && newName) {
        currentChart.title = newName;
    }
    
    // If overwriting existing, use that ID
    if (existingId) {
        currentChart.id = existingId;
    }
    
    await saveChart();
    closeSaveChartModal();
}

async function saveChart() {
    try {
        const method = currentChart.id ? 'PUT' : 'POST';
        const url = currentChart.id ? `/api/charts/${currentChart.id}` : '/api/charts';
        
        const response = await fetch(url, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(currentChart)
        });
        
        if (!response.ok) throw new Error('Failed to save chart');
        
        const data = await response.json();
        currentChart.id = data.id;
        alert('Chart saved successfully!');
    } catch (error) {
        console.error('Error saving chart:', error);
        alert('Error saving chart: ' + error.message);
    }
}

async function exportChart(format) {
    if (!currentChart.id) {
        alert('Please save the chart first');
        return;
    }
    
    try {
        const response = await fetch(`/api/charts/${currentChart.id}/export/${format}`);
        if (!response.ok) throw new Error('Export failed');
        
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `chart-${currentChart.id}.${format}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
    } catch (error) {
        console.error('Error exporting chart:', error);
        alert('Error exporting chart: ' + error.message);
    }
}

// Utilities
function generateId() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function generateClientSVG(chart) {
    // Simple client-side SVG generation for preview
    const quarters = [];
    for (let year = chart.startYear; year <= chart.endYear; year++) {
        const startQ = year === chart.startYear ? chart.startQuarter : 1;
        const endQ = year === chart.endYear ? chart.endQuarter : 4;
        for (let q = startQ; q <= endQ; q++) {
            quarters.push({ year, quarter: q });
        }
    }
    
    const headerHeight = 80;
    const baseRowHeight = 40; // minimum row height
    const quarterWidth = 120;
    const labelWidth = 200;
    const padding = 20;
    const categoryHeaderHeight = 35;
    const titleLineHeight = 14;
    const descLineHeight = 12;
    const verticalPaddingPerTask = 8;

    // Helper to wrap text roughly by character count
    function wrapText(text, maxChars) {
        if (!text) return [];
        const words = text.split(/\s+/);
        const lines = [];
        let line = '';
        words.forEach(w => {
            if ((line + ' ' + w).trim().length <= maxChars) {
                line = (line + ' ' + w).trim();
            } else {
                if (line) lines.push(line);
                line = w;
            }
        });
        if (line) lines.push(line);
        return lines;
    }

    // Calculate dynamic heights per task and category
    let totalCategoryHeadersHeight = 0;
    let totalTaskHeight = 0;
    const perTaskHeights = new Map();
    const perCategoryHeights = new Map();
    chart.categories.forEach(cat => {
        const catNameLines = wrapText(cat.name || '', 30);
        let catH = categoryHeaderHeight;
        if (catNameLines.length > 1) {
            catH = 18 + catNameLines.length * 14; // base + lines * lineheight
        }
        perCategoryHeights.set(cat.id, catH);
        totalCategoryHeadersHeight += catH;
        
        cat.tasks.forEach(task => {
            const titleLines = wrapText(task.title || '', 28);
            const descLines = wrapText(task.description || '', 36);
            const h = Math.max(baseRowHeight, titleLines.length * titleLineHeight + descLines.length * descLineHeight + verticalPaddingPerTask);
            perTaskHeights.set(task.id, h);
            totalTaskHeight += h;
        });
    });

    const width = labelWidth + quarters.length * quarterWidth + padding * 2;
    const height = headerHeight + totalCategoryHeadersHeight + totalTaskHeight + padding * 2;
    
    let svg = `<svg width="${width}" height="${height}" xmlns="http://www.w3.org/2000/svg">`;
    svg += `<defs><style>.title{font:bold 20px sans-serif;fill:#333}.header{font:bold 12px sans-serif;fill:#555}.label{font:12px sans-serif;fill:#333}.category{font:bold 14px sans-serif;fill:#222}.desc{font:10px sans-serif;fill:#666}</style></defs>`;
    svg += `<rect width="${width}" height="${height}" fill="#fafafa"/>`;
    svg += `<text x="${padding}" y="${padding + 20}" class="title">${escapeHtml(chart.title)}</text>`;
    
    // Quarter headers
    quarters.forEach((q, i) => {
        const x = padding + labelWidth + i * quarterWidth;
        const y = headerHeight;
        const color = i % 2 === 0 ? '#e8e8e8' : '#f5f5f5';
        svg += `<rect x="${x}" y="${y - 30}" width="${quarterWidth}" height="30" fill="${color}" stroke="#ccc" stroke-width="1"/>`;
        svg += `<text x="${x + quarterWidth / 2}" y="${y - 10}" class="header" text-anchor="middle">Q${q.quarter} ${q.year}</text>`;
        svg += `<line x1="${x}" y1="${y}" x2="${x}" y2="${height - padding}" stroke="#ddd" stroke-width="1"/>`;
    });
    
    // Categories and tasks - render with wrapping
    let currentY = headerHeight;
    chart.categories.forEach(cat => {
        const catH = perCategoryHeights.get(cat.id) || categoryHeaderHeight;
        svg += `<rect x="${padding}" y="${currentY}" width="${labelWidth}" height="${catH}" fill="${cat.color}" opacity="0.3"/>`;
        // wrap category name if long
        const catLines = wrapText(cat.name || '', 30);
        if (catLines.length <= 1) {
            svg += `<text x="${padding + 10}" y="${currentY + 22}" class="category">${escapeHtml(cat.name)}</text>`;
        } else {
            svg += `<text x="${padding + 10}" y="${currentY + 18}" class="category">`;
            catLines.forEach((ln, idx) => {
                const dy = idx === 0 ? 4 : 14;
                svg += `<tspan x="${padding + 10}" dy="${dy}">${escapeHtml(ln)}</tspan>`;
            });
            svg += `</text>`;
        }
        svg += `<rect x="${padding + labelWidth}" y="${currentY}" width="${quarters.length * quarterWidth}" height="${catH}" fill="${cat.color}" opacity="0.05"/>`;
        currentY += catH;

        cat.tasks.forEach(task => {
            const taskH = perTaskHeights.get(task.id) || baseRowHeight;
            svg += `<rect x="${padding}" y="${currentY}" width="${labelWidth}" height="${taskH}" fill="#fff" stroke="#ddd" stroke-width="1"/>`;

            // Title lines
            const titleLines = wrapText(task.title || '', 28);
            const descLines = wrapText(task.description || '', 36);
            let textY = currentY + 14;
            if (titleLines.length > 0) {
                svg += `<text x="${padding + 10}" y="${textY}" class="label">`;
                titleLines.forEach((ln, idx) => {
                    const dy = idx === 0 ? 0 : titleLineHeight;
                    svg += `<tspan x="${padding + 10}" dy="${dy}">${escapeHtml(ln)}</tspan>`;
                });
                svg += `</text>`;
                textY += titleLines.length * titleLineHeight;
            }
            if (descLines.length > 0) {
                svg += `<text x="${padding + 10}" y="${textY + 4}" class="desc">`;
                descLines.forEach((ln, idx) => {
                    const dy = idx === 0 ? 0 : descLineHeight;
                    svg += `<tspan x="${padding + 10}" dy="${dy}">${escapeHtml(ln)}</tspan>`;
                });
                svg += `</text>`;
            }

            const startIdx = quarters.findIndex(q => q.year === task.startYear && q.quarter === task.startQuarter);
            const endIdx = quarters.findIndex(q => q.year === task.endYear && q.quarter === task.endQuarter);
            if (startIdx >= 0 && endIdx >= 0) {
                const barX = padding + labelWidth + startIdx * quarterWidth;
                const barWidth = (endIdx - startIdx + 1) * quarterWidth;
                const barY = currentY + 8;
                const barHeight = Math.max(12, taskH - 16);
                const taskColor = task.color || cat.color;
                svg += `<rect x="${barX + 2}" y="${barY}" width="${barWidth - 4}" height="${barHeight}" fill="${taskColor}" rx="4" opacity="0.8"/>`;
            }

            currentY += taskH;
        });
    });
    
    svg += '</svg>';
    return svg;
}
