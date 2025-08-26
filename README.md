# 🎬 Myt-V — Local Streaming Server

Myt-V is a lightweight Netflix-style media server written in Go.  
It scans your local movie library, generates a catalog, and serves videos via **HLS (m3u8 + ts)** with a simple web UI.

---

## 🚀 Features
- **Cross-platform** (Windows / Linux).
- Catalog stored in **SQLite** (no external DB required).
- **FFprobe** extracts metadata (codec, resolution, duration).
- **HLS streaming** with `ffmpeg`.
- Web UI built with **Fiber + Bootstrap + hls.js**.
- Ready to run inside a VPN-protected homelab.

---

## 📦 Dependencies

### Required
- **Go 1.22+**  
  - [Download](https://go.dev/dl/)  
  - Check with:  
    ```bash
    go version
    ```

- **FFmpeg (with ffprobe)**  
  Used for scanning media and generating HLS segments.  

  **Windows**:
  1. Download FFmpeg static build: [https://ffmpeg.org/download.html](https://ffmpeg.org/download.html)  
  2. Extract to `C:\ffmpeg\` (or similar).  
  3. Add `C:\ffmpeg\bin` to your system **PATH**.  
  4. Verify:  
     ```powershell
     ffmpeg -version
     ffprobe -version
     ```

  **Linux (Debian/Ubuntu):**
  ```bash
  sudo apt update
  sudo apt install -y ffmpeg
  ffmpeg -version
  ffprobe -version
  ```

  **Linux (Fedora/CentOS/RHEL):**
  ```bash
  sudo dnf install ffmpeg ffmpeg-devel
  ```

- **Git** (to clone the repository)  
  ```bash
  git --version
  ```

---

## 🛠️ Project Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/Myt-v.git
   cd Myt-v
   ```

2. Install Go dependencies:
   ```bash
   go mod tidy
   ```

3. Create `.env` file:
   ```ini
   # Example config
   BIND=127.0.0.1:8080
   MEDIA_DIR=D:\Peliculas   # or /home/user/Peliculas
   HLS_DIR=./hls
   APP_ENV=dev
   ```

4. Run the server:
   ```bash
   go run .
   ```

5. Open in browser:
   - Catalog: [http://127.0.0.1:8080/catalog](http://127.0.0.1:8080/catalog)  
   - Player:  [http://127.0.0.1:8080/watch/1](http://127.0.0.1:8080/watch/1)

---

## 📂 Project Structure
```
/internal
   /db        -> SQLite models + migrations
   /scanner   -> Library scanner with ffprobe
   /stream    -> HLS stream generator
/public
   index.html -> Catalog UI
   watch.html -> Video player UI
/hls          -> HLS output (m3u8 + ts)
/media        -> Your movies
main.go       -> App entrypoint
```

---

## ⚡ Notes
- For **best performance**, normalize your media to `h264 + aac` so Myt-V can use `-c copy` instead of transcoding.
- Myt-V is designed to run behind a **VPN**; do not expose it directly to the internet.
- If you want GPU-accelerated transcoding, ensure your FFmpeg build supports NVENC/VAAPI/QuickSync.
