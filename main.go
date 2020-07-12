package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var workDir string

func index(w http.ResponseWriter, req *http.Request) {
	f, err := os.Open("index.html")
	if err != nil {
		http.Error(w, "Parse index err", http.StatusInternalServerError)
		return
	}
	io.Copy(w, f)
}

func saveFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		http.Error(w, "Url Param 'path' is missing", http.StatusMethodNotAllowed)
		return
	}

	path, ok := r.URL.Query()["path"]
	if !ok {
		http.Error(w, "Url Param 'path' is missing", http.StatusBadRequest)
		return
	}

	filepath := fmt.Sprintf("%s/%s", workDir, path[0])
	file, err := os.OpenFile(filepath, os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Open file err %v\n", err)
		http.Error(w, "Open file err", 400)
		return
	}
	defer file.Close()

	file.Truncate(0)
	n, err := io.Copy(file, r.Body)
	if err != nil {
		log.Printf("Write file err %v\n", err)
		http.Error(w, "Write file err", 400)
		return
	}
	log.Printf("Save to file %s. written %d\n", filepath, n)
	w.WriteHeader(http.StatusNoContent)
}

func preview(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in f", r)
		}
	}()

	overlays, ok := r.URL.Query()["overlay"]
	if !ok {
		http.Error(w, "Url Param 'overlay' is missing", http.StatusBadRequest)
		return
	}

	path := fmt.Sprintf("%s/%s", workDir, overlays[0])
	log.Printf("Preview path %s\n", path)
	cmd := exec.Command("kustomize", "build", path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Preview err %v\n", err)
		http.Error(w, "Preview err", 500)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Printf("Preview %s err. %v\n", path, err)
		http.Error(w, "Preview err", http.StatusInternalServerError)
		return
	}
	io.Copy(w, stdout)
	if err := cmd.Wait(); err != nil {
		log.Printf("Preview %s err. %v\n", path, err)
		http.Error(w, "Preview err", http.StatusInternalServerError)
		return
	}
}

func createFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		http.Error(w, "Url Param 'path' is missing", http.StatusMethodNotAllowed)
		return
	}

	path, ok := r.URL.Query()["path"]
	if !ok {
		http.Error(w, "Url Param 'path' is missing", http.StatusBadRequest)
		return
	}

	filepath := fmt.Sprintf("%s/%s", workDir, path[0])
	file, err := os.Create(filepath)
	if err != nil {
		log.Printf("Create file err %v\n", err)
		http.Error(w, "Create file err", 400)
		return
	}
	defer file.Close()
	log.Printf("Create file %s\n", filepath)
	w.WriteHeader(http.StatusNoContent)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		http.Error(w, "Url Param 'path' is missing", http.StatusMethodNotAllowed)
		return
	}

	path, ok := r.URL.Query()["path"]
	if !ok {
		http.Error(w, "Url Param 'path' is missing", http.StatusBadRequest)
		return
	}

	filepath := fmt.Sprintf("%s/%s", workDir, path[0])
	err := os.Remove(filepath)
	if err != nil {
		log.Printf("Delete file err %v\n", err)
		http.Error(w, "Delete file err", 400)
		return
	}
	log.Printf("Delete file %s\n", filepath)
	w.WriteHeader(http.StatusNoContent)
}

func createOverlay(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		http.Error(w, "Url Param 'path' is missing", http.StatusMethodNotAllowed)
		return
	}

	overlay, ok := r.URL.Query()["overlay"]
	if !ok {
		http.Error(w, "Url Param 'overlay' is missing", http.StatusBadRequest)
		return
	}

	overlayPath := fmt.Sprintf("%s/%s", workDir, overlay[0])
	err := os.Mkdir(overlayPath, 0755)
	if err != nil {
		log.Printf("Create overlay err %v\n", err)
		http.Error(w, "Create overlay err", 400)
		return
	}

	kustomizeFilepath := fmt.Sprintf("%s/kustomization.yaml", overlayPath)
	_, err = os.Create(kustomizeFilepath)
	if err != nil {
		log.Printf("Create kustomize file err %v\n", err)
		http.Error(w, "Create kustomize file err", 400)
		return
	}

	log.Printf("Create overlay %s\n", overlayPath)
	w.WriteHeader(http.StatusNoContent)
}

func renameOverlay(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		http.Error(w, "Url Param 'path' is missing", http.StatusMethodNotAllowed)
		return
	}

	overlay, ok := r.URL.Query()["overlay"]
	if !ok {
		http.Error(w, "Url Param 'overlay' is missing", http.StatusBadRequest)
		return
	}

	newName, ok := r.URL.Query()["name"]
	if !ok {
		http.Error(w, "Url Param 'name' is missing", http.StatusBadRequest)
		return
	}

	overlayPath := fmt.Sprintf("%s/%s", workDir, overlay[0])
	newPath := fmt.Sprintf("%s/%s", workDir, newName[0])

	err := os.Rename(overlayPath, newPath)
	if err != nil {
		log.Printf("Move overlay err %v\n", err)
		http.Error(w, "Move overlay err", 400)
		return
	}

	log.Printf("Move overlay %s to %s\n", overlayPath, newPath)
	w.WriteHeader(http.StatusNoContent)
}

func deleteOverlay(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		http.Error(w, "Url Param 'path' is missing", http.StatusMethodNotAllowed)
		return
	}

	overlay, ok := r.URL.Query()["overlay"]
	if !ok {
		http.Error(w, "Url Param 'overlay' is missing", http.StatusBadRequest)
		return
	}


	overlayPath := fmt.Sprintf("%s/%s", workDir, overlay[0])

	err := removeContents(overlayPath)
	if err != nil {
		log.Printf("Delete overlay err %v\n", err)
		http.Error(w, "Delete overlay err", 400)
		return
	}

	log.Printf("Delete overlay %s\n", overlayPath)
	w.WriteHeader(http.StatusNoContent)
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return os.Remove(dir)
}

func main() {
	port := flag.String("p", "8100", "port to serve on")
	flag.StringVar(&workDir, "d", ".", "kustomize dir")
	flag.Parse()

	if workDir == "." {
		path, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
			return
		}
		workDir = path
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/save", saveFile)
	http.HandleFunc("/preview", preview)
	http.HandleFunc("/create", createFile)
	http.HandleFunc("/delete", deleteFile)
	http.HandleFunc("/create_overlay", createOverlay)
	http.HandleFunc("/rename_overlay", renameOverlay)
	http.HandleFunc("/delete_overlay", deleteOverlay)
	http.Handle("/files/", JsonFileServer(http.Dir(workDir), "/files"))

	http.ListenAndServe(":"+*port, nil)
}
