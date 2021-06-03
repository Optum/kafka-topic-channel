package dispatcher

type ChannelConfig struct {
	Namespace     string
	Name          string
	HostName      string
	Subscriptions []Subscription
}

func (cc ChannelConfig) SubscriptionsUIDs() []string {
	res := make([]string, 0, len(cc.Subscriptions))
	for _, s := range cc.Subscriptions {
		res = append(res, string(s.UID))
	}
	return res
}
