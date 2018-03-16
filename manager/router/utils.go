package router

func listToMap(channels []*Channel) map[ChannelID]*Channel {
	channelMap := make(map[ChannelID]*Channel, len(channels))
	for _, c := range channels {
		channelMap[c.ChannelID] = c
	}
	return channelMap
}
