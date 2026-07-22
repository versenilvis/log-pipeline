package main

import "sync"

/*
There's one signal source (Postgres NOTIFY) but there can be
N recipients (N open Dashboard tabs, each tab a separate WebSocket connection)
We need a mechanism: "One incoming signal -> automatically forwards to all N listeners",
without knowing the number of listeners beforehand,
and users can join/leave at any time (open a new tab, close a tab)

Imagine Hub like a radio station: the station only broadcasts once,
but many different radios (many clients) can receive that signal
Hub keeps a list of "who's listening," and when there's a new message,
it iterates through the list and sends it to each person
*/
type Hub struct {
	mu         sync.Mutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

/*
Because Go maps are insecure when multiple goroutines read/write simultaneously (concurrent map write),
Hub runs a hidden goroutine Run()
that manages data exclusively across three channels:
  - register (add new tabs)
  - unregister (delete closed tabs)
  - broadcast (receive data from Listeners to iterate through the map and distribute to each client)
*/
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.Unlock()
		}
	}
}
