package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
)

type compressWriter struct {
	rw                    http.ResponseWriter // обычный writer
	cw                    io.WriteCloser      // writer для сжатия
	contentTypeToCompress []string            // значения заголовка Content-Type, при которых необходимо сжимать данные
}

func newCompressWriter(w http.ResponseWriter) (*compressWriter, error) {
	cw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	return &compressWriter{
		rw:                    w,
		cw:                    cw,
		contentTypeToCompress: []string{"application/json", "text/html"},
	}, nil
}

func (c *compressWriter) Write(b []byte) (int, error) {
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

	logger.Log.Sugar().Infoln("compressed", "gzip")

	if shouldCompress {
		c.rw.Header().Set("Content-Encoding", "gzip")
		return c.cw.Write(b)
	}

	return c.rw.Write(b)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.rw.WriteHeader(statusCode)
}

func (c *compressWriter) Header() http.Header {
	return c.rw.Header()
}

func (c *compressWriter) Close() error {
    return c.cw.Close()
}

func compressResponseGZIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			logger.Log.Sugar().Errorln(err)
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
