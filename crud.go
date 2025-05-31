// main.go
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db   *sql.DB
	tmpl = template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>LLM Access Form</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 2rem; background-color: #f5f5f5; }
        form { background: #fff; padding: 2rem; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        input { margin: 0.5rem 0; padding: 0.5rem; width: 100%; }
        button { padding: 0.75rem 1.5rem; background: #007bff; color: white; border: none; border-radius: 4px; }
    </style>
</head>
<body>
    <h2>Access the LLM Tester</h2>
    <form action="/start" method="post">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" required>

        <label for="rut">RUT:</label>
        <input type="text" id="rut" name="rut" required>

        <button type="submit">Continue</button>
    </form>
</body>
</html>
	`))
	mu sync.Mutex
)

func getUserAndModel(rut string) (int, int, string, error) {
	var userID, modelID int
	var modelName string

	err := db.QueryRow("SELECT idUsuario, idModelo FROM Usuario WHERE RUT = ?", rut).Scan(&userID, &modelID)
	if err != nil {
		return 0, 0, "", err
	}

	err = db.QueryRow("SELECT NombreModelo FROM ModeloLLM WHERE idModelo = ?", modelID).Scan(&modelName)
	if err != nil {
		return 0, 0, "", err
	}

	print(userID, modelID, "\n")
	return userID, modelID, modelName, nil
}

func insertPromptAndGetID(prompt string, userID int) (int, error) {
	res, err := db.Exec("INSERT INTO Consulta (Texto, idUsuario) VALUES (?, ?)", prompt, userID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func insertResponse(response string, promptID, modelID int) error {
	_, err := db.Exec("INSERT INTO Respuesta (Texto, idPrompt, idModelo) VALUES (?, ?, ?)", response, promptID, modelID)
	return err
}

func userEntryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST and GET allowed", http.StatusMethodNotAllowed)
		return
	}

	rut := r.FormValue("rut")
	name := r.FormValue("name")

	_, _, model, err := getUserAndModel(rut)
	if err != nil {
		http.Error(w, "User not found or no model assigned", http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, "/chat.html?rut="+rut+"&model="+model+"&name="+name, http.StatusSeeOther)
}

type chatRequest struct {
	Message string `json:"message"`
	Model   string `json:"model"`
	UserID  int    `json:"user_id"`
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	promptID, err := insertPromptAndGetID(req.Message, req.UserID)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	print(req.UserID)
	modelID := 0
	db.QueryRow("SELECT idModelo FROM Usuario WHERE idUsuario = ?", req.UserID).Scan(&modelID)
	print(modelID)

	ollamaBody := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Message,
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(ollamaBody)

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", buf)
	if err != nil {
		http.Error(w, "Ollama error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	var full string
	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err == io.EOF {
			break
		} else if err != nil {
			break
		}
		if content, ok := chunk["response"].(string); ok {
			fmt.Fprintf(w, "data: %s\n\n", content)
			flusher.Flush()
			full += content
		}
	}

	insertResponse(full, promptID, modelID)
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/datasql")
	if err != nil {
		log.Fatalf("DB open error: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	http.HandleFunc("/start", userEntryHandler)
	http.HandleFunc("/chat", chatHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
