/**
 * Maze Visualizer - High-Performance Canvas & WebSocket Engine
 */

// Битовые маски стен
const WALL_TOP    = 1;
const WALL_RIGHT  = 2;
const WALL_BOTTOM = 4;
const WALL_LEFT   = 8;
const WALL_ALL    = 15;

class MazeApp {
    constructor() {
        this.canvas = document.getElementById('mazeCanvas');
        this.ctx = this.canvas.getContext('2d');
        this.wrapper = document.getElementById('canvasWrapper');

        this.width = 25;
        this.height = 25;
        this.grid = []; // 2D array: walls bitmask
        this.visited = new Map(); // key "x,y" => { intensity, type }
        this.frontier = new Set(); // "x,y"
        this.path = []; // [[x, y], ...]
        this.ripples = []; // shockwaves: { x, y, radius, maxRadius, alpha }

        this.start = [0, 0];
        this.goal = [24, 24];

        this.eventQueue = [];
        this.isPaused = false;
        this.speed = 35;
        this.accumulatedTime = 0;
        this.lastFrameTime = performance.now();

        this.ws = null;
        this.initDOM();
        this.initGrid();
        this.resizeCanvas();
        this.connectWebSocket();

        window.addEventListener('resize', () => this.resizeCanvas());
        requestAnimationFrame((t) => this.renderLoop(t));
    }

    initDOM() {
        this.genSelect = document.getElementById('genAlgorithm');
        this.solveSelect = document.getElementById('solveAlgorithm');
        this.genHint = document.getElementById('genHint');
        this.solveHint = document.getElementById('solveHint');

        this.sliderGridSize = document.getElementById('sliderGridSize');
        this.valGridSize = document.getElementById('valGridSize');
        this.sliderSpeed = document.getElementById('sliderSpeed');
        this.valSpeed = document.getElementById('valSpeed');

        this.btnGenerate = document.getElementById('btnGenerate');
        this.btnSolve = document.getElementById('btnSolve');
        this.btnPausePlay = document.getElementById('btnPausePlay');
        this.iconPausePlay = document.getElementById('iconPausePlay');
        this.btnStep = document.getElementById('btnStep');
        this.btnClearSolution = document.getElementById('btnClearSolution');
        this.btnReset = document.getElementById('btnReset');

        this.statusText = document.getElementById('statusText');
        this.hudMessage = document.getElementById('hudMessage');
        this.connStatus = document.getElementById('connStatus');

        this.statVisited = document.getElementById('statVisited');
        this.statFrontier = document.getElementById('statFrontier');
        this.statPathLen = document.getElementById('statPathLen');
        this.statTime = document.getElementById('statTime');

        // Подсказки к алгоритмам
        const genHints = {
            backtracker: "Глубокий обход со стеком: строит длинные, извилистые коридоры с высоким коэффициентом блуждания.",
            prim: "Алгоритм Прима: вырастает из стартовой точки в виде плотной ветвистой структуры с множеством коротких тупиков.",
            kruskal: "Алгоритм Краскала (DSU): глобально объединяет разрозненные деревья по всей сетке без образования циклов."
        };
        const solveHints = {
            astar: "Оптимизированный волновой луч: эвристика Манхэттена f(n)=g(n)+h(n) направляет фронт точно к выходу.",
            bfs: "Поиск в ширину: равномерная концентрическая волна уровней, гарантирующая кратчайший путь.",
            dijkstra: "Алгоритм Дейкстры: раскрытие вершин по минимальной стоимости через очередь с приоритетом."
        };

        this.genSelect.addEventListener('change', () => {
            this.genHint.textContent = genHints[this.genSelect.value];
        });
        this.solveSelect.addEventListener('change', () => {
            this.solveHint.textContent = solveHints[this.solveSelect.value];
        });

        // Слайдер размера сетки
        this.sliderGridSize.addEventListener('input', (e) => {
            const size = parseInt(e.target.value);
            this.valGridSize.textContent = `${size} × ${size}`;
            this.width = size;
            this.height = size;
            this.start = [0, 0];
            this.goal = [size - 1, size - 1];
            this.eventQueue = [];
            this.sendCmd({ action: 'init', width: size, height: size });
            this.initGrid();
            this.resizeCanvas();
        });

        // Слайдер скорости
        this.sliderSpeed.addEventListener('input', (e) => {
            this.speed = parseInt(e.target.value);
            let label = "Нормальная";
            if (this.speed < 15) label = "Замедленная (1×)";
            else if (this.speed < 40) label = "Обычная (5×)";
            else if (this.speed < 75) label = "Быстрая (15×)";
            else label = "Молния (MAX)";
            this.valSpeed.textContent = label;
        });

        // Кнопки действий
        this.btnGenerate.addEventListener('click', () => {
            this.clearSolution();
            this.eventQueue = [];
            this.initGrid();
            this.statusText.textContent = "Генерация лабиринта...";
            this.sendCmd({
                action: 'generate',
                algorithm: this.genSelect.value,
                width: this.width,
                height: this.height,
                seed: Date.now()
            });
        });

        this.btnSolve.addEventListener('click', () => {
            this.clearSolution();
            this.eventQueue = [];
            this.statusText.textContent = "Поиск выхода...";
            this.sendCmd({
                action: 'solve',
                algorithm: this.solveSelect.value,
                start_x: this.start[0],
                start_y: this.start[1],
                goal_x: this.goal[0],
                goal_y: this.goal[1]
            });
        });

        this.btnPausePlay.addEventListener('click', () => {
            this.isPaused = !this.isPaused;
            this.iconPausePlay.textContent = this.isPaused ? "▶️ Пуск" : "⏸️ Пауза";
        });

        this.btnStep.addEventListener('click', () => {
            this.isPaused = true;
            this.iconPausePlay.textContent = "▶️ Пуск";
            this.processOneEvent();
        });

        this.btnClearSolution.addEventListener('click', () => {
            this.clearSolution();
        });

        this.btnReset.addEventListener('click', () => {
            this.eventQueue = [];
            this.initGrid();
            this.clearSolution();
            this.sendCmd({ action: 'reset', width: this.width, height: this.height });
        });
    }

    initGrid() {
        this.grid = [];
        for (let y = 0; y < this.height; y++) {
            const row = new Uint8Array(this.width);
            row.fill(WALL_ALL);
            this.grid.push(row);
        }
    }

    clearSolution() {
        this.visited.clear();
        this.frontier.clear();
        this.path = [];
        this.statVisited.textContent = "0";
        this.statFrontier.textContent = "0";
        this.statPathLen.textContent = "0";
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            this.connStatus.textContent = "ONLINE";
            this.connStatus.className = "status-indicator online";
            this.sendCmd({ action: 'init', width: this.width, height: this.height });
        };

        this.ws.onmessage = (event) => {
            try {
                const e = JSON.parse(event.data);
                this.eventQueue.push(e);
            } catch (err) {
                console.error("Ошибка парсинга WebSocket события:", err);
            }
        };

        this.ws.onclose = () => {
            this.connStatus.textContent = "OFFLINE";
            this.connStatus.className = "status-indicator offline";
            setTimeout(() => this.connectWebSocket(), 2000);
        };
    }

    sendCmd(cmd) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(cmd));
        }
    }

    resizeCanvas() {
        const rect = this.wrapper.getBoundingClientRect();
        const dpr = window.devicePixelRatio || 1;
        const availableW = rect.width - 20;
        const availableH = rect.height - 20;

        // Квадратное или пропорциональное полотно
        const size = Math.min(availableW, availableH);
        this.canvas.width = size * dpr;
        this.canvas.height = size * dpr;
        this.canvas.style.width = `${size}px`;
        this.canvas.style.height = `${size}px`;

        this.ctx.resetTransform();
        this.ctx.scale(dpr, dpr);
        this.displaySize = size;
    }

    // Обработка одного события из очереди
    processOneEvent() {
        if (this.eventQueue.length === 0) return;
        const e = this.eventQueue.shift();

        try {
            if (e.message) {
                this.hudMessage.textContent = e.message;
            }

            if (e.stats) {
                if (e.stats.visited !== undefined) this.statVisited.textContent = e.stats.visited;
                if (e.stats.stack_depth !== undefined) this.statFrontier.textContent = e.stats.stack_depth;
                if (e.stats.frontier_size !== undefined) this.statFrontier.textContent = e.stats.frontier_size;
                if (e.stats.pq_size !== undefined) this.statFrontier.textContent = e.stats.pq_size;
                if (e.stats.queue_size !== undefined) this.statFrontier.textContent = e.stats.queue_size;
                if (e.stats.components !== undefined) this.statFrontier.textContent = e.stats.components;
                if (e.stats.path_length !== undefined) this.statPathLen.textContent = e.stats.path_length;
                if (e.stats.time_ms !== undefined) this.statTime.textContent = `${e.stats.time_ms} ms`;
            }

            const x1 = Number(e.x1 ?? 0);
            const y1 = Number(e.y1 ?? 0);
            const x2 = Number(e.x2 ?? 0);
            const y2 = Number(e.y2 ?? 0);

            switch (e.type) {
                case 'carve': {
                    const dx = x2 - x1;
                    const dy = y2 - y1;

                    if (this.grid[y1] && this.grid[y2]) {
                        if (dx === 1) {
                            this.grid[y1][x1] &= ~WALL_RIGHT;
                            this.grid[y2][x2] &= ~WALL_LEFT;
                        } else if (dx === -1) {
                            this.grid[y1][x1] &= ~WALL_LEFT;
                            this.grid[y2][x2] &= ~WALL_RIGHT;
                        } else if (dy === 1) {
                            this.grid[y1][x1] &= ~WALL_BOTTOM;
                            this.grid[y2][x2] &= ~WALL_TOP;
                        } else if (dy === -1) {
                            this.grid[y1][x1] &= ~WALL_TOP;
                            this.grid[y2][x2] &= ~WALL_BOTTOM;
                        }
                        // Вспышка и частица разрушения стены (Shockwave ripple)
                        this.addRipple((x1 + x2) / 2, (y1 + y2) / 2);
                    }
                    break;
                }

                case 'visit': {
                    const key = `${x1},${y1}`;
                    this.visited.set(key, { intensity: e.intensity || this.visited.size, type: 'visit' });
                    this.frontier.delete(key);
                    break;
                }

                case 'frontier': {
                    const key = `${x1},${y1}`;
                    if (!this.visited.has(key)) {
                        this.frontier.add(key);
                    }
                    break;
                }

                case 'backtrack': {
                    const key = `${x1},${y1}`;
                    const item = this.visited.get(key);
                    if (item) item.type = 'backtrack';
                    break;
                }

                case 'path': {
                    this.path.push([x1, y1]);
                    break;
                }

                case 'complete': {
                    if (e.path && e.path.length > 0) {
                        this.path = e.path;
                        this.statusText.textContent = "Путь найден!";
                    } else {
                        this.statusText.textContent = "Готов";
                    }
                    break;
                }

                case 'reset': {
                    if (e.stats && e.stats.width) {
                        this.width = e.stats.width;
                        this.height = e.stats.height;
                        this.start = [0, 0];
                        this.goal = [this.width - 1, this.height - 1];
                    }
                    this.initGrid();
                    this.clearSolution();
                    break;
                }
            }
        } catch (err) {
            console.error("Ошибка при обработке события:", err, e);
        }
    }

    addRipple(gx, gy) {
        this.ripples.push({
            gx: gx,
            gy: gy,
            progress: 0,
            maxRadius: 24,
            alpha: 1.0
        });
    }

    renderLoop(timestamp) {
        try {
            const delta = timestamp - this.lastFrameTime;
            this.lastFrameTime = timestamp;

            // Поглощение событий из очереди в зависимости от скорости
            if (!this.isPaused && this.eventQueue.length > 0) {
                let eventsToProcess = 1;
                if (this.speed >= 80) {
                    eventsToProcess = Math.min(this.eventQueue.length, 50);
                } else if (this.speed >= 50) {
                    eventsToProcess = 10;
                } else if (this.speed >= 30) {
                    eventsToProcess = 4;
                } else if (this.speed >= 15) {
                    eventsToProcess = 2;
                } else {
                    this.accumulatedTime += delta;
                    const interval = (15 - this.speed) * 15;
                    if (this.accumulatedTime >= interval) {
                        eventsToProcess = 1;
                        this.accumulatedTime = 0;
                    } else {
                        eventsToProcess = 0;
                    }
                }

                for (let i = 0; i < eventsToProcess; i++) {
                    this.processOneEvent();
                }
            }

            this.draw();
        } catch (err) {
            console.error("Ошибка в renderLoop:", err);
        } finally {
            requestAnimationFrame((t) => this.renderLoop(t));
        }
    }

    draw() {
        const ctx = this.ctx;
        const size = this.displaySize;
        if (!size) return;

        ctx.clearRect(0, 0, size, size);

        const pad = 12;
        const drawW = size - pad * 2;
        const drawH = size - pad * 2;
        const cellW = drawW / this.width;
        const cellH = drawH / this.height;

        // 1. Отрисовка исследованных ячеек (Волновой фронт / Heatmap)
        for (const [key, val] of this.visited.entries()) {
            const [x, y] = key.split(',').map(Number);
            const cx = pad + x * cellW;
            const cy = pad + y * cellH;

            if (val.type === 'backtrack') {
                ctx.fillStyle = 'rgba(255, 0, 85, 0.12)';
            } else {
                // Плавный градиент волны от циана к фиолетовому
                const t = Math.min(1.0, val.intensity / (this.width * this.height * 0.7));
                const r = Math.round(0 * (1 - t) + 138 * t);
                const g = Math.round(240 * (1 - t) + 43 * t);
                const b = Math.round(255 * (1 - t) + 226 * t);
                ctx.fillStyle = `rgba(${r}, ${g}, ${b}, 0.28)`;
            }
            ctx.fillRect(cx, cy, cellW, cellH);
        }

        // 2. Отрисовка фронтира (пульсирующий контур)
        if (this.frontier.size > 0) {
            ctx.fillStyle = 'rgba(0, 240, 255, 0.4)';
            for (const key of this.frontier) {
                const [x, y] = key.split(',').map(Number);
                const cx = pad + x * cellW;
                const cy = pad + y * cellH;
                ctx.fillRect(cx + 1, cy + 1, cellW - 2, cellH - 2);
            }
        }

        // 3. Отрисовка волн разрушения стен (Shockwave Ripples)
        for (let i = this.ripples.length - 1; i >= 0; i--) {
            const rip = this.ripples[i];
            rip.progress += 0.08;
            rip.alpha = 1.0 - rip.progress;

            if (rip.progress >= 1.0) {
                this.ripples.splice(i, 1);
                continue;
            }

            const rx = pad + rip.gx * cellW + cellW / 2;
            const ry = pad + rip.gy * cellH + cellH / 2;
            const radius = rip.progress * rip.maxRadius;

            ctx.save();
            ctx.beginPath();
            ctx.arc(rx, ry, radius, 0, Math.PI * 2);
            ctx.strokeStyle = `rgba(0, 240, 255, ${rip.alpha * 0.8})`;
            ctx.lineWidth = 2;
            ctx.shadowColor = '#00f0ff';
            ctx.shadowBlur = 8;
            ctx.stroke();
            ctx.restore();
        }

        // 4. Отрисовка стен лабиринта (Неоновые перегородки)
        ctx.strokeStyle = '#2d416b';
        ctx.lineWidth = Math.max(1.5, Math.min(4, cellW * 0.12));
        ctx.lineCap = 'round';

        for (let y = 0; y < this.height; y++) {
            for (let x = 0; x < this.width; x++) {
                const walls = this.grid[y][x];
                const cx = pad + x * cellW;
                const cy = pad + y * cellH;

                ctx.beginPath();
                if (walls & WALL_TOP) {
                    ctx.moveTo(cx, cy);
                    ctx.lineTo(cx + cellW, cy);
                }
                if (walls & WALL_RIGHT) {
                    ctx.moveTo(cx + cellW, cy);
                    ctx.lineTo(cx + cellW, cy + cellH);
                }
                if (walls & WALL_BOTTOM) {
                    ctx.moveTo(cx, cy + cellH);
                    ctx.lineTo(cx + cellW, cy + cellH);
                }
                if (walls & WALL_LEFT) {
                    ctx.moveTo(cx, cy);
                    ctx.lineTo(cx, cy + cellH);
                }
                ctx.stroke();
            }
        }

        // 5. Отрисовка старта и финиша (Glowing Beacons)
        const renderBeacon = (pos, color, glow) => {
            const cx = pad + pos[0] * cellW + cellW / 2;
            const cy = pad + pos[1] * cellH + cellH / 2;
            const r = Math.max(3, Math.min(cellW, cellH) * 0.35);

            ctx.save();
            ctx.beginPath();
            ctx.arc(cx, cy, r, 0, Math.PI * 2);
            ctx.fillStyle = color;
            ctx.shadowColor = glow;
            ctx.shadowBlur = 14;
            ctx.fill();

            // Внешнее пульсирующее кольцо
            ctx.beginPath();
            ctx.arc(cx, cy, r + 3, 0, Math.PI * 2);
            ctx.strokeStyle = color;
            ctx.lineWidth = 1.5;
            ctx.stroke();
            ctx.restore();
        };

        renderBeacon(this.start, '#00ff88', '#00ff88');
        renderBeacon(this.goal, '#ff0055', '#ff0055');

        // 6. Отрисовка кратчайшего пути (Неоновый золотой луч)
        if (this.path.length > 1) {
            ctx.save();
            ctx.beginPath();
            for (let i = 0; i < this.path.length; i++) {
                const pt = this.path[i];
                const cx = pad + pt[0] * cellW + cellW / 2;
                const cy = pad + pt[1] * cellH + cellH / 2;
                if (i === 0) ctx.moveTo(cx, cy);
                else ctx.lineTo(cx, cy);
            }
            ctx.strokeStyle = '#ffb700';
            ctx.lineWidth = Math.max(2, cellW * 0.28);
            ctx.lineCap = 'round';
            ctx.lineJoin = 'round';
            ctx.shadowColor = '#ffb700';
            ctx.shadowBlur = 16;
            ctx.stroke();
            ctx.restore();
        }
    }
}

// Запуск приложения при загрузке страницы
window.addEventListener('DOMContentLoaded', () => {
    window.app = new MazeApp();
});
