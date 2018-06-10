package optimisation

import "github.com/bitlum/hub/manager/router"

func listToMap(channels []*router.Channel) map[router.ChannelID]*router.Channel {
	channelMap := make(map[router.ChannelID]*router.Channel, len(channels))
	for _, c := range channels {
		channelMap[c.ChannelID] = c
	}
	return channelMap
}
