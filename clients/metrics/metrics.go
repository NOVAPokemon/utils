package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	battleDurations = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "client_battle_duration",
	})

	battleMessageDurations = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "client_battle_message_duration",
	})

	notificationMessageDurations = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "client_notification_message_duration",
	})

	tradeMessageDurations = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "client_trade_message_duration",
	})
)

func EmitBattleDuration(duration float64) {
	battleDurations.Observe(duration)
}

func EmitBattleMessageDuration(duration float64) {
	battleMessageDurations.Observe(duration)
}

func EmitNotificationMessageDuration(duration float64) {
	notificationMessageDurations.Observe(duration)
}

func EmittradeMessageDuration(duration float64) {
	tradeMessageDurations.Observe(duration)
}
