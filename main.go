package main

import (
	"bytes"
    "image"
    "image/color"
    "image/draw"
    "image/png"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
	"time"
	"math/rand"
	
	"golang.org/x/image/font"
    "golang.org/x/image/font/basicfont"
    "golang.org/x/image/math/fixed"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

var (
	totalRequests   uint64
	totalUploaded   uint64
	totalDownload	uint64
)

// Gambar garis sederhana dengan algoritma Bresenham
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
    dx := abs(x2 - x1)
    dy := -abs(y2 - y1)
    sx := 1
    if x1 >= x2 { sx = -1 }
    sy := 1
    if y1 >= y2 { sy = -1 }
    err := dx + dy

    for {
        img.Set(x1, y1, col)
        if x1 == x2 && y1 == y2 {
            break
        }
        e2 := 2 * err
        if e2 >= dy {
            err += dy
            x1 += sx
        }
        if e2 <= dx {
            err += dx
            y1 += sy
        }
    }
}

func abs(x int) int {
    if x < 0 { return -x }
    return x
}

// Menambahkan teks ke image
func addLabel(img *image.RGBA, x, y int, label string) {
    col := color.RGBA{0, 0, 0, 255} // hitam
    face := basicfont.Face7x13
    d := &font.Drawer{
        Dst:  img,
        Src:  image.NewUniform(col),
        Face: face,
        Dot:  fixed.P(x, y),
    }
    d.DrawString(label)
}

func main() {
	r := router.New()

	// Root page
	r.ANY("/", func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Method()) {
		case fasthttp.MethodGet:
			ip := string(ctx.Request.Header.Peek("X-Forwarded-For"))
			if ip == "" {
				ip = ctx.RemoteAddr().String()
			}
			ctx.SetContentType("text/html; charset=utf-8")
			ctx.SetBodyString(fmt.Sprintf(`
				<!DOCTYPE html>
				<html>
				<head>
				<title>HPTest</title>
				<link rel="stylesheet" href="/style.css">
                <script src="/app.js"></script>
				</head>
				<body>
					<h1>HPTest Backend</h1>
					<ul>
						<li><a href="/html">Sample HTML</a></li>
						<li><a href="/genfile?size=1024">Generate 1KB File</a></li>
						<li><a href="/genfile?size=1048576">Generate 1MB File</a></li>
					</ul>
					<img src="/img/logo.png">
                	<p>Remote IP: %s</p>
					<h2>Upload Form</h2>
					<form enctype="multipart/form-data" action="/upload" method="post">
						<input type="file" name="file" />
						<input type="submit" value="Upload" />
					</form>
				</body>
				</html>
			`,ip))
		default:
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			ctx.SetBodyString("Method not allowed\n")
		}
	})

	    // Generate dummy static content
    r.GET("/style.css", func(ctx *fasthttp.RequestCtx) {
        data := []byte(`body { font-family: Arial; background: #eee; } h1 { color: #333; }`)
        totalDownload += uint64(len(data))
        ctx.SetContentType("text/css")
        ctx.SetBody(data)
    })

    r.GET("/app.js", func(ctx *fasthttp.RequestCtx) {
        data := []byte(`console.log("Hello from dummy JS");`)
        totalDownload += uint64(len(data))
        ctx.SetContentType("application/javascript")
        ctx.SetBody(data)
    })

    // Dummy images
    r.GET("/img/{name}", func(ctx *fasthttp.RequestCtx) {
		name := ctx.UserValue("name").(string)

		rand.Seed(time.Now().UnixNano())
		const (
			width  = 300
			height = 200
		)

		// Background putih
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

		// Gambar beberapa garis acak
		for i := 0; i < 10; i++ {
			x1 := rand.Intn(width)
			y1 := rand.Intn(height)
			x2 := rand.Intn(width)
			y2 := rand.Intn(height)

			col := color.RGBA{
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)),
				uint8(rand.Intn(256)),
				255,
			}
			drawLine(img, x1, y1, x2, y2, col)
		}

		// Tulis nama request di tengah gambar
		addLabel(img, 20, height/2, name)

		// Encode PNG ke buffer
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			ctx.Error("Failed to encode image", fasthttp.StatusInternalServerError)
			return
		}

		data := buf.Bytes()
		totalDownload += uint64(len(data))
		ctx.SetContentType("image/png")
		ctx.SetBody(data)
	})

    // Directory listing dummy
    r.GET("/dir/", func(ctx *fasthttp.RequestCtx) {
        ctx.SetContentType("text/html; charset=utf-8")
        fmt.Fprintf(ctx, "<h1>Directory listing</h1><ul>")
        for i := 1; i <= 5; i++ {
            fmt.Fprintf(ctx, `<li><a href="/dir/file%d.txt">file%d.txt</a></li>`, i, i)
        }
        fmt.Fprintf(ctx, "</ul>")
    })

    r.GET("/dir/{filename}", func(ctx *fasthttp.RequestCtx) {
        data := make([]byte, 1024*rand.Intn(50)+1024) // 1KB-50KB dummy
        rand.Read(data)
        totalDownload += uint64(len(data))
        ctx.SetBody(data)
    })
	r.GET("/echo/{code}", func(ctx *fasthttp.RequestCtx) {
		codeStr := ctx.UserValue("code").(string)
		var code int
		if _, err := fmt.Sscanf(codeStr, "%d", &code); err != nil {
			ctx.Error("Invalid status code", fasthttp.StatusBadRequest)
			return
		}

		// Optional: body singkat
		ctx.SetStatusCode(code)
		ctx.SetContentType("text/plain; charset=utf-8")
		fmt.Fprintf(ctx, "Echo HTTP Status: %d\n", code)
	})
	// Catch-all route untuk menerima semua path
	r.GET("/path/{any:*}", func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		ip := string(ctx.Request.Header.Peek("X-Forwarded-For"))
		if ip == "" {
			ip = ctx.RemoteAddr().String()
		}

		// Bisa men-generate response dinamis atau dummy content
		ctx.SetContentType("text/plain; charset=utf-8")
		fmt.Fprintf(ctx, "Path: %s, Remote IP: %s\n", path, ip)
	})

	// Simple HTML
	r.ANY("/html", func(ctx *fasthttp.RequestCtx) {
		ctx.SetContentType("text/html")
		ctx.SetBodyString("<html><body><h1>Hello from HPTest</h1></body></html>")
		trackStats(ctx, len(ctx.Response.Body()), 0)
	})

	// Generate file with given size
	r.ANY("/genfile", func(ctx *fasthttp.RequestCtx) {
		sizeStr := string(ctx.QueryArgs().Peek("size"))
		size, _ := strconv.Atoi(sizeStr)
		if size <= 0 {
			size = 1024
		}
		data := make([]byte, size)
		for i := 0; i < size; i++ {
			data[i] = 'A'
		}
		ctx.SetContentType("application/octet-stream")
		ctx.SetBody(data)
		trackStats(ctx, size, 0)
	})

	// File upload
	r.POST("/upload", func(ctx *fasthttp.RequestCtx) {
		fileHeader, err := ctx.FormFile("file")
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			fmt.Fprintf(ctx, "failed to get form file: %v", err)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			fmt.Fprintf(ctx, "failed to open uploaded file: %v", err)
			return
		}
		defer file.Close()

		// Discard semua data, tidak disimpan ke disk
		written, err := io.Copy(io.Discard, file)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			fmt.Fprintf(ctx, "failed to process file: %v", err)
			return
		}

		ctx.SetStatusCode(fasthttp.StatusOK)
		fmt.Fprintf(ctx, "uploaded %s (%d bytes, discarded)\n", fileHeader.Filename, written)

		// catat statistik upload
		trackStats(ctx, 0, int(written))
	})

	// Statistics
	r.ANY("/stats", func(ctx *fasthttp.RequestCtx) {
		reqs := atomic.LoadUint64(&totalRequests)
		down := atomic.LoadUint64(&totalDownload)
		up := atomic.LoadUint64(&totalUploaded)
		ctx.SetContentType("text/plain")
		ctx.SetBodyString(fmt.Sprintf(
			"Requests: %d\nDownloaded: %d bytes\nUploaded: %d bytes\n",
			reqs, down, up,
		))
	})

	// Print stats every 10s
	go func() {
		for {
			time.Sleep(10 * time.Second)
			reqs := atomic.LoadUint64(&totalRequests)
			down := atomic.LoadUint64(&totalDownload)
			up := atomic.LoadUint64(&totalUploaded)
			log.Printf("[STATS] requests=%d downloaded=%dB uploaded=%dB\n", reqs, down, up)
		}
	}()

	log.Println("Server started on :5000")
	s := &fasthttp.Server{
		Handler:            r.Handler,
		MaxRequestBodySize: 101 * 1024 * 1024, // atau sesuai kebutuhan
	}
	if err := s.ListenAndServe(":5000"); err != nil {
		log.Fatal(err)
	}
}

func trackStats(ctx *fasthttp.RequestCtx, downloaded int, uploaded int) {
	atomic.AddUint64(&totalRequests, 1)
	atomic.AddUint64(&totalDownload, uint64(downloaded))
	atomic.AddUint64(&totalUploaded, uint64(uploaded))

	clientIP := string(ctx.Request.Header.Peek("X-Forwarded-For"))
	if clientIP == "" {
		clientIP, _, _ = net.SplitHostPort(ctx.RemoteAddr().String())
	}
	log.Printf("REQ %s %s from %s d=%dB u=%dB\n",
		ctx.Method(), ctx.Path(), clientIP, downloaded, uploaded)
}

