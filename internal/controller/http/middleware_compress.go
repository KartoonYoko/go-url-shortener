package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"go.uber.org/zap"
)

type compressWriter struct {
	rw                    http.ResponseWriter // обычный writer для http ответа
	cw                    io.WriteCloser      // writer для сжатия
	contentTypeToCompress []string            // значения заголовка Content-Type, при которых необходимо сжимать данные
	shouldCompress        bool                // нужно ли сжимать данные
}

func newCompressWriter(w http.ResponseWriter) (*compressWriter, error) {
	// изменим алгортим сжатия, чтобы сверить alloc_space профилировщиком
	cw, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
	if err != nil {
		return nil, err
	}

	return &compressWriter{
		rw:                    w,
		cw:                    cw,
		contentTypeToCompress: []string{"application/json", "text/html"},
		shouldCompress:        true,
	}, nil
}

func (c *compressWriter) Write(b []byte) (int, error) {
	logger.Log.Info("compress body by gzip")

	if c.shouldCompress {
		return c.cw.Write(b)
	}

	return c.rw.Write(b)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	shouldCompress := false
	for _, v := range c.rw.Header().Values("content-type") {
		if shouldCompress {
			break
		}
		for _, c := range c.contentTypeToCompress {
			if strings.Contains(v, c) {
				shouldCompress = true
				break
			}
		}
	}
	c.shouldCompress = shouldCompress

	if c.shouldCompress {
		c.rw.Header().Set("Content-Encoding", "gzip")
	}

	c.rw.WriteHeader(statusCode)
}

func (c *compressWriter) Header() http.Header {
	return c.rw.Header()
}

func (c *compressWriter) Close() error {
	if c.shouldCompress {
		return c.cw.Close()
	}

	return nil
}

// skipResponseGZIPCompress определяет нужно ли пропустить сжатие
func skipResponseGZIPCompress(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/debug")
}

func compressResponseGZIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipResponseGZIPCompress(r) {
			next.ServeHTTP(w, r)
			return
		}
		supportsGzip := false
		for _, v := range r.Header.Values("Accept-Encoding") {
			if strings.Contains(v, "gzip") {
				supportsGzip = true
				break
			}
		}

		if !supportsGzip {
			next.ServeHTTP(w, r)
			return
		}

		rw, err := newCompressWriter(w)
		if err != nil {
			logger.Log.Sugar().Errorln("http.compressResponseGZIPMiddleware", zap.Error(err))
			io.WriteString(w, err.Error())
			return
		}
		defer rw.Close()
		next.ServeHTTP(rw, r)
	})
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func decompressRequestGZIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if !sendsGzip {
			next.ServeHTTP(w, r)
			return
		}

		// при чтении вернётся распакованный слайс байт
		cr, err := newCompressReader(r.Body)
		if err != nil {
			logger.Log.Sugar().Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cr.zr.Close()
		r.Body = cr

		logger.Log.Sugar().Infoln("decompress", "gzip")
		next.ServeHTTP(w, r)
	})
}
