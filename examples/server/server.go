package main

import (
	"context"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/northbright/filehashes"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	addr = flag.String("addr", "localhost:8080", "http service address")
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

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	man, chMsg := filehashes.NewManager(
		filehashes.DefaultConcurrency,
		filehashes.DefaultBufferSize,
	)
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

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/ws")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/ws", websocketHandler)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
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
