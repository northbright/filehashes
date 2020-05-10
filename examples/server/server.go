package main

import (
	"context"
	"crypto"
	_ "crypto/md5"
	_ "crypto/sha1"
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
	ctxMap      = map[string]context.Context{}
	cancelMap   = map[string]context.CancelFunc{}
)

type Command struct {
	Action string             `json:"action"`
	Req    filehashes.Request `json:"req"`
}

// readHashMessages reads the messages during compute checksums of files,
// marshals the messages to JSON strings,
// writes the strings as responses to websocket connection.
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

// hashHandler is the handler to process requests to compute file checksums.
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
			cmd := Command{}

			log.Printf("m: %s", string(m))
			if err := json.Unmarshal(m, &cmd); err != nil {
				log.Printf("json.Unmarshal() error: %v", err)
				continue
			}

			reqs := []*filehashes.Request{
				&cmd.Req,
			}

			if cmd.Action == "start" {
				ctxMap[cmd.Req.File], cancelMap[cmd.Req.File] = context.WithCancel(ctx)

				man.Start(ctxMap[cmd.Req.File], reqs)
			} else if cmd.Action == "stop" {
				cancelMap[cmd.Req.File]()
			}
		}
	}
}

// shutdown is the handler to shutdown the HTTP server.
func shutdown(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("shutdown server..."))
	// Call cancel func
	cancel()
}

// home is the handler to render main page.
func home(w http.ResponseWriter, r *http.Request) {
	// Create default request for front-end.
	req := filehashes.NewRequest(
		"../../filehashes.go",
		[]crypto.Hash{
			crypto.MD5,
			crypto.SHA1,
		},
		nil,
	)

	buf, _ := json.Marshal(req)

	data := struct {
		// Websocket address.
		Addr string
		// Default request JSON to send.
		ReqsJSON string
	}{
		"ws://" + r.Host + "/ws",
		string(buf),
	}

	homeTemplate.Execute(w, data)
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
			log.Printf("shutdown server: os.Interrupt received")
			cancel()
		case <-ctx.Done():
			log.Printf("shutdown server: user request")
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
    var req;
    var resume_req;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{ .Addr }}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
	       var msg = JSON.parse(evt.data);
	       if (msg.type === 3) {
	           resume_req = msg.data;
	           console.log("resume_req:" + resume_req);
	       } else {
	           if (msg.type === 6) {
		       resume_req = null;
		   }
               }
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("start").onclick = function(evt) {
        if (!ws) {
            return false;
        }

        var data = {};
	data.action = "start";
	if (!resume_req) { 
	    req = JSON.parse(input.value);
	    data.req = req;
	} else {
	    data.req = resume_req;
	    console.log("data.req: " + data.req);
	}

        var str = JSON.stringify(data);
        print("SEND: " + str);
        ws.send(str);
        return false;
    };
    document.getElementById("stop").onclick = function(evt) {
        if (!ws) {
            return false;
        }

        var data = {};
	data.action = "stop";
	data.req = req;
	var str = JSON.stringify(data);
        print("SEND: " + str);
        ws.send(str);
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
<p><input id="input" type="text" value="{{ .ReqsJSON }}">
<button id="start">Start/Stop</button>
<button id="stop">Stop</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
