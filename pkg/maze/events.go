package maze

// EventType определяет тип анимационного события
type EventType string

const (
	EventCarve     EventType = "carve"     // Пробитие стены между ячейками (вспышка/частицы)
	EventVisit     EventType = "visit"     // Посещение ячейки (закраска волны)
	EventFrontier  EventType = "frontier"  // Добавление в открытый список / фронтир (пульсирующий контур)
	EventBacktrack EventType = "backtrack" // Откат при тупике (DFS)
	EventPath      EventType = "path"      // Ячейка финального кратчайшего пути (неоновый луч)
	EventComplete  EventType = "complete"  // Завершение работы алгоритма
	EventReset     EventType = "reset"     // Сброс лабиринта
	EventInfo      EventType = "info"      // Текстовый статус для HUD
)

// Event описывает атомарный шаг визуализации
type Event struct {
	Type      EventType      `json:"type"`
	X1        int            `json:"x1"`
	Y1        int            `json:"y1"`
	X2        int            `json:"x2"`
	Y2        int            `json:"y2"`
	Wall      uint8          `json:"wall,omitempty"`      // Какая стена была разрушена
	Intensity float64        `json:"intensity,omitempty"` // Интенсивность свечения/стоимость (g-score / f-score)
	Path      [][2]int       `json:"path,omitempty"`      // Финальный путь
	Stats     map[string]any `json:"stats,omitempty"`     // Метрики (посещено, длина пути, время)
	Message   string         `json:"message,omitempty"`   // Сообщение для HUD
	StepIndex int            `json:"step_index"`          // Порядковый номер шага
}

// EventListener канал для передачи событий визуализации
type EventListener chan Event
