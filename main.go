package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
	"github.com/ysv/pkg/httputil/handler"
	"github.com/ysv/pkg/logger"
)

const(
	LogoSize = 150
	GeneratedQRImageSize = 500

	DataMaxLength = 256

	ImageMinSize     = 100
	ImageMaxSize     = 1000
	DefaultImageSize = 250
)

func main() {
	var generator qrGenerator

	logoURL, ok := os.LookupEnv("LOGO_URL")
	if ok {
		logo, err := loadPNGImage(logoURL)
		if err != nil {
			logger.Fatalf("Failed to load logo from: %v", logoURL)
		}
		logo = resize.Resize(LogoSize, LogoSize, logo, resize.Lanczos3)
		generator.logo = &logo
	}

	srv := server{qrGenerator: &generator}

	http.Handle("/", srv.qrGenerateHandler())

	addr := ":8080"
	logger.Warnf("listening on %v...", addr)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatalf("http server failed: %v", err)
	}
}

type server struct {
	qrGenerator *qrGenerator
}

func loadPNGImage(logoURL string) (image.Image, error) {
	resp, err := http.Get(logoURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to load logo. status: %v", resp.StatusCode)
	}

	pngImage, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return pngImage, nil
}

func (s *server) qrGenerateHandler() handler.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		rawData := r.URL.Query().Get("data")
		if rawData == "" {
			return handler.NewAPIError(422, "data.missing")
		} else if len(rawData) > DataMaxLength {
			return handler.NewAPIError(422, "data.too_long")
		}

		qrData, err := url.PathUnescape(rawData)
		if err != nil {
			return err
		}

		size, err := strconv.ParseUint(r.URL.Query().Get("size"), 10, 64)
		if err != nil {
			size = DefaultImageSize
		} else if size > ImageMaxSize {
			return handler.NewAPIError(422, "size.too_big")
		} else if size < ImageMinSize {
			return handler.NewAPIError(422, "size.too_small")
		}

		qrCode, err := s.qrGenerator.Generate(qrData, uint(size))
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		if err := png.Encode(buffer, qrCode); err != nil {
			return err
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

		if _, err := w.Write(buffer.Bytes()); err != nil {
			return err
		}
		return nil
	}
}

type qrGenerator struct {
	logo *image.Image
}

func (gen *qrGenerator) Generate(data string, size uint) (image.Image,error) {
	code, err := qrcode.New(data, qrcode.Highest)
	if err != nil {
		return nil, err
	}

	qrWithoutLogo := code.Image(GeneratedQRImageSize)
	if gen.logo == nil {
		return resize.Resize(size, size, qrWithoutLogo, resize.Lanczos3), nil
	}

	qrWithLogo := image.NewRGBA(qrWithoutLogo.Bounds())

	logoPosition0 := GeneratedQRImageSize / 2 - LogoSize / 2
	logoPosition1 := logoPosition0 + LogoSize

	draw.Draw(qrWithLogo, qrWithoutLogo.Bounds(), qrWithoutLogo, image.Point{}, draw.Src)
	draw.Draw(qrWithLogo, image.Rect(logoPosition0, logoPosition0, logoPosition1, logoPosition1), *gen.logo, image.Point{}, draw.Over)

	return resize.Resize(size, size, qrWithLogo, resize.Lanczos3), nil
}
