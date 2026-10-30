package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"maze_visualizer/pkg/engine"
	"maze_visualizer/pkg/maze"
	"maze_visualizer/web"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем локальные запросы
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка апгрейда соединения в WebSocket: %v", err)
		return
	}
	defer conn.Close()

	var writeMu sync.Mutex
	sendEvent := func(e maze.Event) {
		writeMu.Lock()
		defer writeMu.Unlock()
		data, err := json.Marshal(e)
		if err != nil {
			log.Printf("Ошибка маршалинга события: %v", err)
			return
		}
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Ошибка отправки через WebSocket: %v", err)
		}
	}

	session := engine.NewSession(sendEvent)
	log.Println("⚡ Новое подключение клиента к визуализатору лабиринтов")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket закрыт: %v", err)
			}
			break
		}

		cmd, err := engine.ParseCommand(msg)
		if err != nil {
			log.Printf("Ошибка парсинга команды: %v", err)
			continue
		}

		go session.HandleCommand(cmd)
	}
}

func main() {
	port := flag.Int("port", 8080, "Порт HTTP сервера")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(web.Assets)))
	mux.HandleFunc("/ws", handleWebSocket)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Println("╔════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  ⚡ ВИЗУАЛИЗАТОР ЛАБИРИНТОВ НА GO ЗАПУЩЕН НА http://localhost:%d  ║\n", *port)
	fmt.Println("║  Алгоритмы генерации: Prim, Kruskal (DSU), Recursive Backtracker  ║")
	fmt.Println("║  Алгоритмы решения: A* (Manhattan), BFS, Dijkstra                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════╝")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
