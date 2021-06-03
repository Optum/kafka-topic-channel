/*
Copyright 2021 The Optum Authors
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// KafkaTopicChannelLister helps list KafkaTopicChannels.
// All objects returned here must be treated as read-only.
type KafkaTopicChannelLister interface {
	// List lists all KafkaTopicChannels in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.KafkaTopicChannel, err error)
	// KafkaTopicChannels returns an object that can list and get KafkaTopicChannels.
	KafkaTopicChannels(namespace string) KafkaTopicChannelNamespaceLister
	KafkaTopicChannelListerExpansion
}

// kafkaTopicChannelLister implements the KafkaTopicChannelLister interface.
type kafkaTopicChannelLister struct {
	indexer cache.Indexer
}

// NewKafkaTopicChannelLister returns a new KafkaTopicChannelLister.
func NewKafkaTopicChannelLister(indexer cache.Indexer) KafkaTopicChannelLister {
	return &kafkaTopicChannelLister{indexer: indexer}
}

// List lists all KafkaTopicChannels in the indexer.
func (s *kafkaTopicChannelLister) List(selector labels.Selector) (ret []*v1alpha1.KafkaTopicChannel, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.KafkaTopicChannel))
	})
	return ret, err
}

// KafkaTopicChannels returns an object that can list and get KafkaTopicChannels.
func (s *kafkaTopicChannelLister) KafkaTopicChannels(namespace string) KafkaTopicChannelNamespaceLister {
	return kafkaTopicChannelNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// KafkaTopicChannelNamespaceLister helps list and get KafkaTopicChannels.
// All objects returned here must be treated as read-only.
type KafkaTopicChannelNamespaceLister interface {
	// List lists all KafkaTopicChannels in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.KafkaTopicChannel, err error)
	// Get retrieves the KafkaTopicChannel from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.KafkaTopicChannel, error)
	KafkaTopicChannelNamespaceListerExpansion
}

// kafkaTopicChannelNamespaceLister implements the KafkaTopicChannelNamespaceLister
// interface.
type kafkaTopicChannelNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all KafkaTopicChannels in the indexer for a given namespace.
func (s kafkaTopicChannelNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.KafkaTopicChannel, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.KafkaTopicChannel))
	})
	return ret, err
}

// Get retrieves the KafkaTopicChannel from the indexer for a given namespace and name.
func (s kafkaTopicChannelNamespaceLister) Get(name string) (*v1alpha1.KafkaTopicChannel, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("kafkatopicchannel"), name)
	}
	return obj.(*v1alpha1.KafkaTopicChannel), nil
}
