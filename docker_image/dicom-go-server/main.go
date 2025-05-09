// main.go
package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/suyashkumar/dicom"
	dcmTag "github.com/suyashkumar/dicom/pkg/tag"
)

const uploadDir = "./uploads"

func main() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/header", headerHandler)
	http.HandleFunc("/image", imageHandler)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	file, hdr, err := r.FormFile("dicom")
	if err != nil {
		http.Error(w, "must include form-field ‘dicom’", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// prepend a nanosecond timestamp to guarantee uniqueness
	ts := time.Now().UnixNano()
	storedName := fmt.Sprintf("%d_%s", ts, hdr.Filename)
	dst := filepath.Join(uploadDir, storedName)

	out, err := os.Create(dst)
	if err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "error writing file", http.StatusInternalServerError)
		return
	}

	// return the stored filename so client can refer to it
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"file": storedName})
}

func headerHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	tagStr := r.URL.Query().Get("tag")
	if fileName == "" || tagStr == "" {
		http.Error(w, "need both ‘file’ and ‘tag’ query params", http.StatusBadRequest)
		return
	}

	// parse tag string like "00100010"
	if len(tagStr) != 8 {
		http.Error(w, "tag must be 8-hex chars (e.g. 00100010)", http.StatusBadRequest)
		return
	}
	grp, err := parseHex(tagStr[:4])
	elm, err2 := parseHex(tagStr[4:])
	if err != nil || err2 != nil {
		http.Error(w, "invalid hex in tag", http.StatusBadRequest)
		return
	}

	ds, err := dicom.ParseFile(filepath.Join(uploadDir, fileName), nil)
	if err != nil {
		http.Error(w, "failed to parse DICOM", http.StatusInternalServerError)
		return
	}
	elem, err := ds.FindElementByTag(dcmTag.Tag{Group: grp, Element: elm})
	if err != nil {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Tag %s → %v", tagStr, elem.Value)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "need ‘file’ query param", http.StatusBadRequest)
		return
	}
	ds, err := dicom.ParseFile(filepath.Join(uploadDir, fileName), nil)
	if err != nil {
		http.Error(w, "failed to parse DICOM", http.StatusInternalServerError)
		return
	}

	pde, err := ds.FindElementByTag(dcmTag.PixelData)
	if err != nil {
		http.Error(w, "no PixelData found", http.StatusNotFound)
		return
	}
	info := dicom.MustGetPixelDataInfo(pde.Value)
	if len(info.Frames) == 0 {
		http.Error(w, "no image frames present", http.StatusNotFound)
		return
	}

	img, err := info.Frames[0].GetImage()
	if err != nil {
		http.Error(w, "failed to get image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, img); err != nil {
		http.Error(w, "error encoding PNG", http.StatusInternalServerError)
	}
}

// helper to parse a 4-hex-digit string into uint16
func parseHex(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 16, 16)
	return uint16(v), err
}
