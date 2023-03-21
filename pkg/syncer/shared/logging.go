package shared

import "github.com/go-logr/logr"

const (
	// ReconcilerKey is used to identify a reconciler.
	ReconcilerKey = "reconciler"

	// QueueKeyKey is used to expose the workqueue key being processed.
	QueueKeyKey = "key"
)

// WithReconciler adds the reconciler name to the logger.
func WithReconciler(logger logr.Logger, reconciler string) logr.Logger {
	return logger.WithValues(ReconcilerKey, reconciler)
}

// WithQueueKey adds the queue key to the logger.
func WithQueueKey(logger logr.Logger, key string) logr.Logger {
	return logger.WithValues(QueueKeyKey, key)
}
