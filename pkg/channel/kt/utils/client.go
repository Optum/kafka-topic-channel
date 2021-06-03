package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/kelseyhightower/envconfig"

	"knative.dev/eventing-kafka/pkg/common/client"
)

type AdapterSASL struct {
	Enable   bool   `envconfig:"KAFKA_NET_SASL_ENABLE" required:"false"`
	User     string `envconfig:"KAFKA_NET_SASL_USER" required:"false"`
	Password string `envconfig:"KAFKA_NET_SASL_PASSWORD" required:"false"`
	Type     string `envconfig:"KAFKA_NET_SASL_TYPE" required:"false"`
}

type AdapterTLS struct {
	Enable bool   `envconfig:"KAFKA_NET_TLS_ENABLE" required:"false"`
	Cert   string `envconfig:"KAFKA_NET_TLS_CERT" required:"false"`
	Key    string `envconfig:"KAFKA_NET_TLS_KEY" required:"false"`
	CACert string `envconfig:"KAFKA_NET_TLS_CA_CERT" required:"false"`
}

type AdapterNet struct {
	SASL AdapterSASL
	TLS  AdapterTLS
}

type KafkaEnvConfig struct {
	// KafkaConfigJson is the environment variable that's passed to adapter by the controller.
	// It contains configuration from the Kafka configmap.
	KafkaConfigJson  string   `envconfig:"K_KAFKA_CONFIG"`
	BootstrapServers []string `envconfig:"KAFKA_BOOTSTRAP_SERVERS" required:"true"`
	Net              AdapterNet
}

// NewConfig extracts the Kafka configuration from the environment.
func NewConfigFromEnv(ctx context.Context) ([]string, *sarama.Config, error) {
	var env KafkaEnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, nil, err
	}

	return NewConfigWithEnv(ctx, &env)
}

// NewConfig extracts the Kafka configuration from the environment.
func NewConfigWithEnv(ctx context.Context, env *KafkaEnvConfig) ([]string, *sarama.Config, error) {
	kafkaAuthConfig := &client.KafkaAuthConfig{}

	if env.Net.TLS.Enable {
		kafkaAuthConfig.TLS = &client.KafkaTlsConfig{
			Cacert:   env.Net.TLS.CACert,
			Usercert: env.Net.TLS.Cert,
			Userkey:  env.Net.TLS.Key,
		}
	}

	if env.Net.SASL.Enable {
		kafkaAuthConfig.SASL = &client.KafkaSaslConfig{
			User:     env.Net.SASL.User,
			Password: env.Net.SASL.Password,
			SaslType: env.Net.SASL.Type,
		}
	}

	configBuilder := client.NewConfigBuilder().
		WithDefaults().
		WithAuth(kafkaAuthConfig)

	if env.KafkaConfigJson != "" {
		kafkaCfg := &KafkaConfig{}
		err := json.Unmarshal([]byte(env.KafkaConfigJson), kafkaCfg)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing Kafka config from environment: %w", err)
		}

		configBuilder = configBuilder.FromYaml(kafkaCfg.SaramaYamlString)
	}

	cfg, err := configBuilder.Build(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating Sarama config: %w", err)
	}

	return env.BootstrapServers, cfg, nil
}

// NewProducer is a helper method for constructing a client for producing kafka methods.
func NewProducer(ctx context.Context) (sarama.Client, error) {
	bs, cfg, err := NewConfigFromEnv(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: specific here. can we do this for all?
	cfg.Producer.Return.Errors = true

	return sarama.NewClient(bs, cfg)
}

type KafkaConfig struct {
	SaramaYamlString string
}
