package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/northbright/filehashes"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	addr        = flag.String("addr", "localhost:8080", "http service address")
	ctx, cancel = context.WithCancel(context.Background())
)

func readHashMessages(ctx context.Context, ch <-chan *filehashes.Message, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("readHashMessages() stopped: %v", ctx.Err())
			return
		case m := <-ch:
			buf, err := m.JSON()
			if err != nil {
				log.Printf("m.JSON() error: %v", err)
				continue
			}
			log.Printf("message: %v", string(buf))

			if err := conn.WriteMessage(websocket.TextMessage, buf); err != nil {
				log.Println(err)
				return
			}

		}
	}
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
	// Get the websocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Create a filehashes manager.
	man, chMsg := filehashes.NewManager(
		filehashes.DefaultConcurrency,
		filehashes.DefaultBufferSize,
	)

	// Start a goroutine to read the messages of hashes.
	go func() {
		readHashMessages(ctx, chMsg, conn)
	}()

	for {
		messageType, m, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if messageType == websocket.TextMessage {
			reqs := []*filehashes.Request{}

			if err := json.Unmarshal(m, &reqs); err != nil {
				log.Printf("json.Unmarshal() error: %v", err)
				continue
			}

			man.StartSumFiles(ctx, reqs)
		}
	}
}

func shutdown(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("shutdown server..."))
	// Call cancel func
	cancel()
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/ws")
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	mux := http.NewServeMux()

	mux.HandleFunc("/", home)
	mux.HandleFunc("/ws", hashHandler)
	mux.HandleFunc("/shutdown", shutdown)

	server := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		select {
		case <-sigint:
			// os.Interrupt, call cancel func.
			log.Printf("os.Interrupt received")
			cancel()
		case <-ctx.Done():
			log.Printf("shutdown server from user request")
		}

		// We received an interrupt signal, shut down.
		if err := server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP ListenAndServe: %v", err)
	}

	<-idleConnsClosed
	log.Printf("main() exited")
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
