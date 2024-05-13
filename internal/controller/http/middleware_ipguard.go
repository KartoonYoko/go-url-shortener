package http

import (
	"net/http"
	"net/netip"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"go.uber.org/zap"
)

// guardIPMiddleware запрещает/разрешает доступ на основе переданной подсети
func (c *ShortenerController) guardIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.Header.Get("X-Real-IP")
		// если нет заголовка - запрещаем доступ
		if ipStr == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// если подсеть не задана - запрещаем доступ
		if c.conf.TrustedSubnets == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		network, err := netip.ParsePrefix(c.conf.TrustedSubnets)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("middleware ipguard error: ", zap.Error(err))
			return
		}
		ip, err := netip.ParseAddr(ipStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Error("middleware ipguard error: ", zap.Error(err))
			return
		}

		// если не входит в подсеть - запрещаем доступ
		if !network.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
