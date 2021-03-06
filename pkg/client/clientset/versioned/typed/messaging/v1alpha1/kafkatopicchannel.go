/*
Copyright 2021 The Optum Authors
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	scheme "github.com/optum/kafka-topic-channel/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KafkaTopicChannelsGetter has a method to return a KafkaTopicChannelInterface.
// A group's client should implement this interface.
type KafkaTopicChannelsGetter interface {
	KafkaTopicChannels(namespace string) KafkaTopicChannelInterface
}

// KafkaTopicChannelInterface has methods to work with KafkaTopicChannel resources.
type KafkaTopicChannelInterface interface {
	Create(ctx context.Context, kafkaTopicChannel *v1alpha1.KafkaTopicChannel, opts v1.CreateOptions) (*v1alpha1.KafkaTopicChannel, error)
	Update(ctx context.Context, kafkaTopicChannel *v1alpha1.KafkaTopicChannel, opts v1.UpdateOptions) (*v1alpha1.KafkaTopicChannel, error)
	UpdateStatus(ctx context.Context, kafkaTopicChannel *v1alpha1.KafkaTopicChannel, opts v1.UpdateOptions) (*v1alpha1.KafkaTopicChannel, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.KafkaTopicChannel, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.KafkaTopicChannelList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KafkaTopicChannel, err error)
	KafkaTopicChannelExpansion
}

// kafkaTopicChannels implements KafkaTopicChannelInterface
type kafkaTopicChannels struct {
	client rest.Interface
	ns     string
}

// newKafkaTopicChannels returns a KafkaTopicChannels
func newKafkaTopicChannels(c *MessagingV1alpha1Client, namespace string) *kafkaTopicChannels {
	return &kafkaTopicChannels{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kafkaTopicChannel, and returns the corresponding kafkaTopicChannel object, and an error if there is any.
func (c *kafkaTopicChannels) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.KafkaTopicChannel, err error) {
	result = &v1alpha1.KafkaTopicChannel{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KafkaTopicChannels that match those selectors.
func (c *kafkaTopicChannels) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.KafkaTopicChannelList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.KafkaTopicChannelList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kafkaTopicChannels.
func (c *kafkaTopicChannels) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a kafkaTopicChannel and creates it.  Returns the server's representation of the kafkaTopicChannel, and an error, if there is any.
func (c *kafkaTopicChannels) Create(ctx context.Context, kafkaTopicChannel *v1alpha1.KafkaTopicChannel, opts v1.CreateOptions) (result *v1alpha1.KafkaTopicChannel, err error) {
	result = &v1alpha1.KafkaTopicChannel{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kafkaTopicChannel).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a kafkaTopicChannel and updates it. Returns the server's representation of the kafkaTopicChannel, and an error, if there is any.
func (c *kafkaTopicChannels) Update(ctx context.Context, kafkaTopicChannel *v1alpha1.KafkaTopicChannel, opts v1.UpdateOptions) (result *v1alpha1.KafkaTopicChannel, err error) {
	result = &v1alpha1.KafkaTopicChannel{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		Name(kafkaTopicChannel.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kafkaTopicChannel).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *kafkaTopicChannels) UpdateStatus(ctx context.Context, kafkaTopicChannel *v1alpha1.KafkaTopicChannel, opts v1.UpdateOptions) (result *v1alpha1.KafkaTopicChannel, err error) {
	result = &v1alpha1.KafkaTopicChannel{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		Name(kafkaTopicChannel.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kafkaTopicChannel).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the kafkaTopicChannel and deletes it. Returns an error if one occurs.
func (c *kafkaTopicChannels) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kafkaTopicChannels) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched kafkaTopicChannel.
func (c *kafkaTopicChannels) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.KafkaTopicChannel, err error) {
	result = &v1alpha1.KafkaTopicChannel{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kafkatopicchannels").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
