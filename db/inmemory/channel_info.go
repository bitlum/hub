package inmemory

import (
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/lightning/lnd"
	"sync"
)

type InfoStorage struct {
	mx    sync.RWMutex
	infos map[lightning.ChannelID]*lnd.ChannelAdditionalInfo
}

func NewInfoStorage() *InfoStorage {
	return &InfoStorage{
		infos: make(map[lightning.ChannelID]*lnd.ChannelAdditionalInfo),
	}
}

func (s *InfoStorage) UpdateChannelAdditionalInfo(info *lnd.
ChannelAdditionalInfo) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.infos[info.ChannelID] = info
	return nil
}

// ...
func (s *InfoStorage) GetChannelAdditionalInfoByID(chanID lightning.ChannelID) (
	*lnd.ChannelAdditionalInfo, error) {
	s.mx.Lock()
	defer s.mx.Unlock()


	info, ok := s.infos[chanID]
	if !ok {
		return nil, lnd.ErrorChannelInfoNotFound
	}

	return info, nil
}

// ...
func (s *InfoStorage) GetChannelAdditionalInfoByShortID(shortChanID uint64) (
	*lnd.ChannelAdditionalInfo, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	for _, info := range s.infos {
		if info.ShortChannelID == shortChanID {
			return info, nil
		}
	}

	return nil, lnd.ErrorChannelInfoNotFound
}
