package v1alpha1

import (
	"errors"

	componentsv1alpha1 "github.com/dapr/dapr/pkg/apis/components/v1alpha1"
	kedav1alpha1 "github.com/kedacore/keda/v2/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"

	ofcore "github.com/openfunction/apis/core/v1alpha2"
)

const (
	ComponentVersion    = "v1"
	BindingsKafka       = "bindings.kafka"
	BindingsRedis       = "bindings.redis"
	BindingsCron        = "bindings.cron"
	BindingsMQTT        = "bindings.mqtt"
	ScaleKafka          = "kafka"
	ScaleNatsStreaming  = "stan"
	PubsubNatsStreaming = "pubsub.natsstreaming"
)

type GenericScaleOption struct {
	PollingInterval *int32                            `json:"pollingInterval,omitempty"`
	CooldownPeriod  *int32                            `json:"cooldownPeriod,omitempty"`
	MinReplicaCount *int32                            `json:"minReplicaCount,omitempty"`
	MaxReplicaCount *int32                            `json:"maxReplicaCount,omitempty"`
	Advanced        *kedav1alpha1.AdvancedConfig      `json:"advanced,omitempty"`
	Metadata        map[string]string                 `json:"metadata,omitempty"`
	AuthRef         *kedav1alpha1.ScaledObjectAuthRef `json:"authRef,omitempty"`
}

type NatsStreamingSpec struct {
	NatsURL                 string                    `json:"natsURL"`
	NatsStreamingClusterID  string                    `json:"natsStreamingClusterID"`
	SubscriptionType        string                    `json:"subscriptionType"`
	AckWaitTime             *string                   `json:"ackWaitTime,omitempty"`
	MaxInFlight             *int64                    `json:"maxInFlight,omitempty"`
	DurableSubscriptionName *string                   `json:"durableSubscriptionName,omitempty"`
	DeliverNew              *bool                     `json:"deliverNew,omitempty"`
	StartAtSequence         *int64                    `json:"startAtSequence,omitempty"`
	StartWithLastReceived   *bool                     `json:"startWithLastReceived,omitempty"`
	DeliverAll              *bool                     `json:"deliverAll,omitempty"`
	StartAtTimeDelta        *string                   `json:"startAtTimeDelta,omitempty"`
	StartAtTime             *string                   `json:"startAtTime,omitempty"`
	StartAtTimeFormat       *string                   `json:"startAtTimeFormat,omitempty"`
	ConsumerID              *string                   `json:"consumerID,omitempty"`
	ScaleOption             *NatsStreamingScaleOption `json:"scaleOption,omitempty"`
}

type NatsStreamingScaleOption struct {
	*GenericScaleOption          `json:",inline"`
	NatsServerMonitoringEndpoint string `json:"natsServerMonitoringEndpoint"`
	QueueGroup                   string `json:"queueGroup,omitempty"`
	DurableName                  string `json:"durableName,omitempty"`
	Subject                      string `json:"subject,omitempty"`
	LagThreshold                 string `json:"lagThreshold"`
}

func (spec *NatsStreamingSpec) ConvertToMetadataMap() []map[string]interface{} {
	var m []map[string]interface{}

	// Handling mandatory parameters
	m = append(m, map[string]interface{}{"name": "natsURL", "value": spec.NatsURL})
	m = append(m, map[string]interface{}{"name": "natsStreamingClusterID", "value": spec.NatsStreamingClusterID})
	m = append(m, map[string]interface{}{"name": "subscriptionType", "value": spec.SubscriptionType})

	// Handling optional parameters
	if spec.AckWaitTime != nil {
		m = append(m, map[string]interface{}{"name": "ackWaitTime", "value": *spec.AckWaitTime})
	}
	if spec.MaxInFlight != nil {
		m = append(m, map[string]interface{}{"name": "maxInFlight", "value": *spec.MaxInFlight})
	}
	if spec.DurableSubscriptionName != nil {
		m = append(m, map[string]interface{}{"name": "durableSubscriptionName", "value": *spec.DurableSubscriptionName})
	}
	if spec.DeliverNew != nil {
		m = append(m, map[string]interface{}{"name": "deliverNew", "value": *spec.DeliverNew})
	}
	if spec.StartAtSequence != nil {
		m = append(m, map[string]interface{}{"name": "startAtSequence", "value": *spec.StartAtSequence})
	}
	if spec.StartWithLastReceived != nil {
		m = append(m, map[string]interface{}{"name": "startWithLastReceived", "value": *spec.StartWithLastReceived})
	}
	if spec.DeliverAll != nil {
		m = append(m, map[string]interface{}{"name": "deliverAll", "value": *spec.DeliverAll})
	}
	if spec.StartAtTimeDelta != nil {
		m = append(m, map[string]interface{}{"name": "startAtTimeDelta", "value": *spec.StartAtTimeDelta})
	}
	if spec.StartAtTime != nil {
		m = append(m, map[string]interface{}{"name": "startAtTime", "value": *spec.StartAtTime})
	}
	if spec.StartAtTimeFormat != nil {
		m = append(m, map[string]interface{}{"name": "startAtTimeFormat", "value": *spec.StartAtTimeFormat})
	}
	return m
}

func (spec *NatsStreamingSpec) GenComponent(namespace string, name string, metadataMap []map[string]interface{}) (*componentsv1alpha1.Component, error) {
	component := &componentsv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	component.Spec.Type = PubsubNatsStreaming
	component.Spec.Version = ComponentVersion

	var metadataItems []componentsv1alpha1.MetadataItem
	metadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metadataBytes, &metadataItems)
	if err != nil {
		return nil, err
	}
	component.Spec.Metadata = metadataItems

	return component, nil
}

func (spec *NatsStreamingSpec) GenEventBusScaledObject(subjects []string, consumerID string) (*ofcore.KedaScaledObject, error) {
	if spec.ScaleOption == nil {
		return nil, nil
	}
	scaledObject := &ofcore.KedaScaledObject{}
	scaledObject.Triggers = []kedav1alpha1.ScaleTriggers{}

	scaledObject.MinReplicaCount = spec.ScaleOption.MinReplicaCount
	scaledObject.MaxReplicaCount = spec.ScaleOption.MaxReplicaCount
	scaledObject.CooldownPeriod = spec.ScaleOption.CooldownPeriod
	scaledObject.PollingInterval = spec.ScaleOption.PollingInterval

	for _, subject := range subjects {
		trigger := &kedav1alpha1.ScaleTriggers{}
		trigger.Type = ScaleNatsStreaming
		trigger.Metadata = map[string]string{}
		trigger.Metadata["natsServerMonitoringEndpoint"] = spec.ScaleOption.NatsServerMonitoringEndpoint
		trigger.Metadata["lagThreshold"] = spec.ScaleOption.LagThreshold
		trigger.Metadata["subject"] = subject
		if spec.ConsumerID != nil {
			trigger.Metadata["queueGroup"] = *spec.ConsumerID
		} else {
			trigger.Metadata["queueGroup"] = consumerID
		}
		if spec.DurableSubscriptionName != nil {
			trigger.Metadata["durableName"] = *spec.DurableSubscriptionName
		} else {
			return nil, errors.New("durableSubscriptionName must be set in order to use the scale feature")
		}
		scaledObject.Triggers = append(scaledObject.Triggers, *trigger)
	}
	if scaledObject.Triggers != nil {
		return scaledObject, nil
	}
	return nil, errors.New("scaleOption has no trigger")
}

type KafkaSpec struct {
	Brokers         string            `json:"brokers"`
	AuthRequired    bool              `json:"authRequired"`
	Topic           string            `json:"topic,omitempty"`
	SaslUsername    *string           `json:"saslUsername,omitempty"`
	SaslPassword    *string           `json:"saslPassword,omitempty"`
	MaxMessageBytes *int64            `json:"maxMessageBytes,omitempty"`
	ScaleOption     *KafkaScaleOption `json:"scaleOption,omitempty"`
}

type KafkaScaleOption struct {
	*GenericScaleOption `json:",inline"`
	ConsumerGroup       string `json:"consumerGroup,omitempty"`
	Topic               string `json:"topic,omitempty"`
	LagThreshold        string `json:"lagThreshold"`
}

func (spec *KafkaSpec) ConvertToMetadataMap() []map[string]interface{} {
	var m []map[string]interface{}

	// Handling mandatory parameters
	m = append(m, map[string]interface{}{"name": "brokers", "value": spec.Brokers})
	m = append(m, map[string]interface{}{"name": "publishTopic", "value": spec.Topic})
	m = append(m, map[string]interface{}{"name": "topics", "value": spec.Topic})
	m = append(m, map[string]interface{}{"name": "authRequired", "value": spec.AuthRequired})

	// Handling optional parameters
	if spec.SaslUsername != nil {
		m = append(m, map[string]interface{}{"name": "saslUsername", "value": *spec.SaslUsername})
	}
	if spec.SaslPassword != nil {
		m = append(m, map[string]interface{}{"name": "saslPassword", "value": *spec.SaslPassword})
	}
	if spec.MaxMessageBytes != nil {
		m = append(m, map[string]interface{}{"name": "maxMessageBytes", "value": *spec.MaxMessageBytes})
	}
	return m
}

func (spec *KafkaSpec) GenComponent(namespace string, name string, metadataMap []map[string]interface{}) (*componentsv1alpha1.Component, error) {
	component := &componentsv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	component.Spec.Type = BindingsKafka
	component.Spec.Version = ComponentVersion

	var metadataItems []componentsv1alpha1.MetadataItem
	metadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metadataBytes, &metadataItems)
	if err != nil {
		return nil, err
	}
	component.Spec.Metadata = metadataItems

	return component, nil
}

func (spec *KafkaSpec) GenScaledObject() (*ofcore.KedaScaledObject, error) {
	if spec.ScaleOption == nil {
		return nil, nil
	}
	scaledObject := &ofcore.KedaScaledObject{}
	trigger := &kedav1alpha1.ScaleTriggers{}

	scaledObject.MinReplicaCount = spec.ScaleOption.MinReplicaCount
	scaledObject.MaxReplicaCount = spec.ScaleOption.MaxReplicaCount
	scaledObject.CooldownPeriod = spec.ScaleOption.CooldownPeriod
	scaledObject.PollingInterval = spec.ScaleOption.PollingInterval
	trigger.Type = ScaleKafka

	if spec.ScaleOption.Metadata != nil {
		trigger.Metadata = spec.ScaleOption.Metadata
	} else {
		trigger.Metadata = map[string]string{}
	}
	trigger.Metadata["bootstrapServers"] = spec.Brokers
	trigger.Metadata["topic"] = spec.Topic
	scaledObject.Triggers = []kedav1alpha1.ScaleTriggers{*trigger}
	return scaledObject, nil
}

type RedisSpec struct {
	RedisHost             string  `json:"redisHost"`
	RedisPassword         string  `json:"redisPassword"`
	EnableTLS             *bool   `json:"enableTLS,omitempty"`
	Failover              *bool   `json:"failover,omitempty"`
	SentinelMasterName    *string `json:"sentinelMasterName,omitempty"`
	RedeliverInterval     *string `json:"redeliverInterval,omitempty"`
	ProcessingTimeout     *string `json:"processingTimeout,omitempty"`
	RedisType             *string `json:"redisType,omitempty"`
	RedisDB               *int64  `json:"redisDB,omitempty"`
	RedisMaxRetries       *int64  `json:"redisMaxRetries,omitempty"`
	RedisMinRetryInterval *string `json:"redisMinRetryInterval,omitempty"`
	RedisMaxRetryInterval *string `json:"redisMaxRetryInterval,omitempty"`
	DialTimeout           *string `json:"dialTimeout,omitempty"`
	ReadTimeout           *string `json:"readTimeout,omitempty"`
	WriteTimeout          *string `json:"writeTimeout,omitempty"`
	PoolSize              *int64  `json:"poolSize,omitempty"`
	PoolTimeout           *string `json:"poolTimeout,omitempty"`
	MaxConnAge            *string `json:"maxConnAge,omitempty"`
	MinIdleConns          *int64  `json:"minIdleConns,omitempty"`
	IdleCheckFrequency    *string `json:"idleCheckFrequency,omitempty"`
	IdleTimeout           *string `json:"idleTimeout,omitempty"`
}

func (spec *RedisSpec) ConvertToMetadataMap() []map[string]interface{} {
	var m []map[string]interface{}

	// Handling mandatory parameters
	m = append(m, map[string]interface{}{"name": "redisHost", "value": spec.RedisHost})
	m = append(m, map[string]interface{}{"name": "redisPassword", "value": spec.RedisPassword})

	// Handling optional parameters
	if spec.EnableTLS != nil {
		m = append(m, map[string]interface{}{"name": "enableTLS", "value": *spec.EnableTLS})
	}
	if spec.Failover != nil {
		m = append(m, map[string]interface{}{"name": "failover", "value": *spec.Failover})
	}
	if spec.SentinelMasterName != nil {
		m = append(m, map[string]interface{}{"name": "sentinelMasterName", "value": *spec.SentinelMasterName})
	}
	if spec.RedeliverInterval != nil {
		m = append(m, map[string]interface{}{"name": "redeliverInterval", "value": *spec.RedeliverInterval})
	}
	if spec.ProcessingTimeout != nil {
		m = append(m, map[string]interface{}{"name": "processingTimeout", "value": *spec.ProcessingTimeout})
	}
	if spec.RedisType != nil {
		m = append(m, map[string]interface{}{"name": "redisType", "value": *spec.RedisType})
	}
	if spec.RedisDB != nil {
		m = append(m, map[string]interface{}{"name": "redisDB", "value": *spec.RedisDB})
	}
	if spec.RedisMaxRetries != nil {
		m = append(m, map[string]interface{}{"name": "redisMaxRetries", "value": *spec.RedisMaxRetries})
	}
	if spec.RedisMinRetryInterval != nil {
		m = append(m, map[string]interface{}{"name": "redisMinRetryInterval", "value": *spec.RedisMinRetryInterval})
	}
	if spec.RedisMaxRetryInterval != nil {
		m = append(m, map[string]interface{}{"name": "redisMaxRetryInterval", "value": *spec.RedisMaxRetryInterval})
	}
	if spec.DialTimeout != nil {
		m = append(m, map[string]interface{}{"name": "dialTimeout", "value": *spec.DialTimeout})
	}
	if spec.ReadTimeout != nil {
		m = append(m, map[string]interface{}{"name": "readTimeout", "value": *spec.ReadTimeout})
	}
	if spec.WriteTimeout != nil {
		m = append(m, map[string]interface{}{"name": "writeTimeout", "value": *spec.WriteTimeout})
	}
	if spec.PoolSize != nil {
		m = append(m, map[string]interface{}{"name": "poolSize", "value": *spec.PoolSize})
	}
	if spec.PoolTimeout != nil {
		m = append(m, map[string]interface{}{"name": "poolTimeout", "value": *spec.PoolTimeout})
	}
	if spec.MaxConnAge != nil {
		m = append(m, map[string]interface{}{"name": "maxConnAge", "value": *spec.MaxConnAge})
	}
	if spec.MinIdleConns != nil {
		m = append(m, map[string]interface{}{"name": "minIdleConns", "value": *spec.MinIdleConns})
	}
	if spec.IdleCheckFrequency != nil {
		m = append(m, map[string]interface{}{"name": "idleCheckFrequency", "value": *spec.IdleCheckFrequency})
	}
	if spec.IdleTimeout != nil {
		m = append(m, map[string]interface{}{"name": "idleTimeout", "value": *spec.IdleTimeout})
	}
	return m
}

func (spec *RedisSpec) GenComponent(namespace string, name string, metadataMap []map[string]interface{}) (*componentsv1alpha1.Component, error) {
	component := &componentsv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	component.Spec.Type = BindingsRedis
	component.Spec.Version = ComponentVersion

	var metadataItems []componentsv1alpha1.MetadataItem
	metadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metadataBytes, &metadataItems)
	if err != nil {
		return nil, err
	}
	component.Spec.Metadata = metadataItems

	return component, nil
}

type CronSpec struct {
	Schedule string `json:"schedule"`
}

func (spec *CronSpec) ConvertToMetadataMap() []map[string]interface{} {
	var m []map[string]interface{}

	// Handling mandatory parameters
	m = append(m, map[string]interface{}{"name": "schedule", "value": spec.Schedule})
	return m
}

func (spec *CronSpec) GenComponent(namespace string, name string, metadataMap []map[string]interface{}) (*componentsv1alpha1.Component, error) {
	component := &componentsv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	component.Spec.Type = BindingsCron
	component.Spec.Version = ComponentVersion

	var metadataItems []componentsv1alpha1.MetadataItem
	metadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metadataBytes, &metadataItems)
	if err != nil {
		return nil, err
	}
	component.Spec.Metadata = metadataItems

	return component, nil
}

type MQTTSpec struct {
	Url          string  `json:"url"`
	Topic        string  `json:"topic"`
	ConsumerID   *string `json:"consumerID,omitempty"`
	Qos          *int64  `json:"qos,omitempty"`
	Retain       *bool   `json:"retain,omitempty"`
	CleanSession *bool   `json:"cleanSession,omitempty"`
	CaCert       *string `json:"caCert,omitempty"`
	ClientCert   *string `json:"clientCert,omitempty"`
	ClientKey    *string `json:"clientKey,omitempty"`
}

func (spec *MQTTSpec) ConvertToMetadataMap() []map[string]interface{} {
	var m []map[string]interface{}

	// Handling mandatory parameters
	m = append(m, map[string]interface{}{"name": "url", "value": spec.Url})
	m = append(m, map[string]interface{}{"name": "topic", "value": spec.Topic})

	// Handling optional parameters
	if spec.Qos != nil {
		m = append(m, map[string]interface{}{"name": "qos", "value": *spec.Qos})
	}
	if spec.Retain != nil {
		m = append(m, map[string]interface{}{"name": "retain", "value": *spec.Retain})
	}
	if spec.CleanSession != nil {
		m = append(m, map[string]interface{}{"name": "cleanSession", "value": *spec.CleanSession})
	}
	if spec.CaCert != nil {
		m = append(m, map[string]interface{}{"name": "caCert", "value": *spec.CaCert})
	}
	if spec.ClientCert != nil {
		m = append(m, map[string]interface{}{"name": "clientCert", "value": *spec.ClientCert})
	}
	if spec.ClientKey != nil {
		m = append(m, map[string]interface{}{"name": "clientKey", "value": *spec.ClientKey})
	}
	return m
}

func (spec *MQTTSpec) GenComponent(namespace string, name string, metadataMap []map[string]interface{}) (*componentsv1alpha1.Component, error) {
	component := &componentsv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	component.Spec.Type = BindingsMQTT
	component.Spec.Version = ComponentVersion

	var metadataItems []componentsv1alpha1.MetadataItem
	metadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metadataBytes, &metadataItems)
	if err != nil {
		return nil, err
	}
	component.Spec.Metadata = metadataItems

	return component, nil
}
