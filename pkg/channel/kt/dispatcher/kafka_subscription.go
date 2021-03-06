package dispatcher

import (
	"sync"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

type KafkaSubscription struct {
	logger *zap.SugaredLogger
	subs   sets.String
	// readySubscriptionsLock must be used to synchronize access to channelReadySubscriptions
	readySubscriptionsLock    sync.RWMutex
	channelReadySubscriptions map[string]sets.Int32
}

func NewKafkaSubscription(logger *zap.SugaredLogger) *KafkaSubscription {
	return &KafkaSubscription{
		logger:                    logger,
		subs:                      sets.NewString(),
		channelReadySubscriptions: map[string]sets.Int32{},
	}
}

// SetReady will mark the subid in the KafkaSubscription and call any registered callbacks
func (ks *KafkaSubscription) SetReady(subID types.UID, partition int32, ready bool) {
	ks.logger.Debugw("Setting subscription readiness", zap.Any("subscription", subID), zap.Bool("ready", ready))
	ks.readySubscriptionsLock.Lock()
	defer ks.readySubscriptionsLock.Unlock()
	if ready {
		if subs, ok := ks.channelReadySubscriptions[string(subID)]; ok {
			ks.logger.Debugw("Adding ready ready partition to cached subscription", zap.Any("subscription", subID), zap.Int32("partition", partition))
			subs.Insert(partition)
		} else {
			ks.logger.Debugw("Caching ready subscription", zap.Any("subscription", subID), zap.Int32("partition", partition))
			ks.channelReadySubscriptions[string(subID)] = sets.NewInt32(partition)
		}
	} else {
		if subs, ok := ks.channelReadySubscriptions[string(subID)]; ok {
			ks.logger.Debugw("Ejecting cached ready subscription", zap.Any("subscription", subID), zap.Int32("partition", partition))
			subs.Delete(partition)
			if subs.Len() == 0 {
				delete(ks.channelReadySubscriptions, string(subID))
			}
		}
	}
}
