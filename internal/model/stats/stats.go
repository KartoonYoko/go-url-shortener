package stats

// StatsResponse ответ на запрос получения статистики
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
