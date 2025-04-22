package notifier

import (
	"log/slog"

	"github.com/gurebusan/simple-auth/internal/config"
)

type MockNotifier struct {
	log *slog.Logger
	cfg *config.Config
}

func NewMockNotifier(log *slog.Logger, cfg *config.Config) *MockNotifier {
	return &MockNotifier{
		log: log,
		cfg: cfg,
	}
}

func (n *MockNotifier) Send(guid, email, oldIP, newIP string) error {
	log := n.log.With(
		slog.String("From", n.cfg.Email.From),
		slog.String("To", email),
		slog.String("OldIP", oldIP),
		slog.String("New IP", newIP),
	)
	log.Warn("Warning! IP has been changed", slog.String("GUID", guid))
	return nil
}
