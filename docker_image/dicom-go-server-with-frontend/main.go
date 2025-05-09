package main

import (
	"encoding/json"
	"fmt"
	"html/template"
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

var indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>DICOM Uploader & Viewer</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body class="bg-light">
  <div class="container py-5">
    <h1 class="mb-4">DICOM Uploader & Viewer</h1>

    <div class="card mb-4">
      <div class="card-body">
        <h5 class="card-title">1) Upload DICOM File</h5>
        <form id="uploadForm">
          <div class="mb-3">
            <input class="form-control" type="file" id="dicomFile" name="dicom" required>
          </div>
          <button type="submit" class="btn btn-primary">Upload</button>
        </form>
        <div id="uploadResult" class="mt-3"></div>
      </div>
    </div>

    <div class="card mb-4">
      <div class="card-body">
        <h5 class="card-title">2) View Image</h5>
        <div id="imageContainer"></div>
      </div>
    </div>

    <div class="card">
      <div class="card-body">
        <h5 class="card-title">3) Search Header Tag</h5>
        <div class="alert alert-info">
          <strong>Tag Formatting:</strong> Per <a href="https://www.dicomlibrary.com/dicom/dicom-tags/" target="_blank">DICOM Library</a>, DICOM tags show like <code>(0002,0000)</code>.
          Please enter the tags here without parentheses and comma, e.g. <code>00020000</code> (group + element, each 4 hex digits).
        </div>
        <form id="headerForm">
          <div class="row mb-3">
            <div class="col">
              <input class="form-control" type="text" id="fileName" placeholder="Filename" required>
            </div>
            <div class="col">
              <input class="form-control" type="text" id="tag" placeholder="Tag (e.g. 00080080)" required>
            </div>
          </div>
          <button type="submit" class="btn btn-secondary">Get Tag</button>
        </form>
        <div id="headerResult" class="mt-3"></div>
      </div>
    </div>
  </div>

  <script>
    // Upload handler
    document.getElementById('uploadForm').addEventListener('submit', async e => {
      e.preventDefault();
      const fileInput = document.getElementById('dicomFile');
      const formData = new FormData();
      formData.append('dicom', fileInput.files[0]);

      const res = await fetch('/upload', { method: 'POST', body: formData });
      const data = await res.json();
      if (res.ok) {
        document.getElementById('uploadResult').innerHTML =
          '<div class="alert alert-success">Uploaded as: ' + data.file + '</div>';
        document.getElementById('fileName').value = data.file;
        // Show image
        const img = document.createElement('img');
        img.src = '/image?file=' + encodeURIComponent(data.file);
		img.className = 'img-fluid';
        document.getElementById('imageContainer').innerHTML = '';
        document.getElementById('imageContainer').appendChild(img);
      } else {
        document.getElementById('uploadResult').innerHTML =
          '<div class="alert alert-danger">' + await res.text() + '</div>';
      }
    });

    // Header lookup handler
    document.getElementById('headerForm').addEventListener('submit', async e => {
      e.preventDefault();
      const file = document.getElementById('fileName').value;
      const tag = document.getElementById('tag').value;
      const res = await fetch('/header?file=' + encodeURIComponent(file) + '&tag=' + encodeURIComponent(tag));
      const text = await res.text();
      document.getElementById('headerResult').innerHTML =
        '<div class="alert ' + (res.ok ? 'alert-info' : 'alert-warning') + '">' + text + '</div>';
    });
  </script>
</body>
</html>`))

func main() {
	// ensure uploads dir
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create upload dir: %v", err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/header", headerHandler)
	http.HandleFunc("/image", imageHandler)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	indexTmpl.Execute(w, nil)
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

	if len(tagStr) != 8 {
		http.Error(w, "tag must be 8-hex chars (e.g. 00100010)", http.StatusBadRequest)
		return
	}
	grp, err1 := strconv.ParseUint(tagStr[:4], 16, 16)
	elm, err2 := strconv.ParseUint(tagStr[4:], 16, 16)
	if err1 != nil || err2 != nil {
		http.Error(w, "invalid hex in tag", http.StatusBadRequest)
		return
	}

	ds, err := dicom.ParseFile(filepath.Join(uploadDir, fileName), nil)
	if err != nil {
		http.Error(w, "failed to parse DICOM", http.StatusInternalServerError)
		return
	}
	elem, err := ds.FindElementByTag(dcmTag.Tag{Group: uint16(grp), Element: uint16(elm)})
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
	png.Encode(w, img)
}
