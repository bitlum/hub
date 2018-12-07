package optimisation

import "github.com/bitlum/hub/lightning"

func listToMap(channels []*lightning.Channel) map[lightning.ChannelID]*lightning.Channel {
	channelMap := make(map[lightning.ChannelID]*lightning.Channel, len(channels))
	for _, c := range channels {
		channelMap[c.ChannelID] = c
	}
	return channelMap
}
