package logs

import (
	"github.com/bitlum/hub/lightning"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"os"
	"reflect"
	"time"
)

// getState converts nodes topology into state log.
func getState(client lightning.Client) (*Log, error) {
	// Get the node's local network in order to write it on log file,
	// so that external optimisation program could sync it state.
	nodeChannels, err := client.Channels()
	if err != nil {
		return nil, err
	}

	// TODO(andrew.shvv) As far we have gap between those two operations
	// the free balance might actually change, need to cope with that.

	// Get the number of money under control of lightning node in order to write
	// in the log.
	freeBalance, err := client.FreeBalance()
	if err != nil {
		return nil, err
	}

	// Get the number of money which is in the process of being proceeded by
	// blockchain.
	pendingBalance, err := client.PendingBalance()
	if err != nil {
		return nil, err
	}

	channels := make([]*Channel, len(nodeChannels))
	for i, c := range nodeChannels {
		channels[i] = &Channel{
			ChannelId:     string(c.ChannelID),
			UserId:        string(c.UserID),
			RemoteBalance: uint64(c.RemoteBalance),
			LocalBalance:  uint64(c.LocalBalance),
			IsPending:     c.IsPending(),
		}
	}

	return &Log{
		Time: time.Now().UnixNano(),
		Data: &Log_State{
			State: &NodeState{
				FreeBalance:    uint64(freeBalance),
				PendingBalance: uint64(pendingBalance),
				Channels:       channels,
			},
		},
	}, nil
}

// UpdateLogFileGoroutine subscribe on lightning node topology updates and
// update the log with the current node state and channel updates.
//
// NOTE: Should be run as goroutine.
func UpdateLogFileGoroutine(client lightning.Client, path string, errChan chan error) {
	defer func() {
		log.Infof("Stopped update log file goroutine, log path(%v)", path)
	}()

	log.Infof("Start update log file goroutine, log path(%v)", path)

	var logEntry *Log

	// Ensure that gRPC structures are printed properly
	pretty := spew.NewDefaultConfig()
	pretty.DisableMethods = true
	pretty.DisablePointerAddresses = true

	logEntry, err := getState(client)
	if err != nil {
		fail(errChan, "unable to get state: %v", err)
		return
	}

	var needWriteState <-chan time.Time
	triggerStateWrite := func() {
		if needWriteState == nil {
			needWriteState = time.After(3 * time.Second)
		}
	}

	receiver := client.RegisterOnUpdates()
	defer receiver.Stop()

	for {
		if logEntry != nil {
			// NOTE: If move open/close of the file out of this cycle than this
			// would lead to optimisation third-party program unable to get and
			// log update via watchdog package.
			log.Tracef("Open update log file(%v) to write an update: %v",
				path, pretty.Sdump(logEntry))
			updateLogFile, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|
				os.O_CREATE, 0666)
			if err != nil {
				fail(errChan, "unable to open update log file: %v", err)
				return
			}

			if err := WriteLog(updateLogFile, logEntry); err != nil {
				fail(errChan, "unable to write new log entry: %v", err)
				return
			}

			if err := updateLogFile.Close(); err != nil {
				fail(errChan, "unable to close log file: %v", err)
				return
			}
			logEntry = nil
		}

		select {
		case update, ok := <-receiver.Read():
			if !ok {
				log.Info("Client update channel close, " +
					"exiting log update goroutine")
				return
			}

			switch u := update.(type) {
			case *lightning.UpdateChannelClosing:
				log.Infof("Update(%v) received, logging", u)

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_closing,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							RemoteBalance: 0,
							LocalBalance:  0,
							Fee:           uint64(u.Fee),
							Duration:      0,
						},
					},
				}

			case *lightning.UpdateChannelClosed:
				log.Infof("Update(%v) received, logging", u)

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_closed,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							RemoteBalance: 0,
							LocalBalance:  0,
							Fee:           uint64(u.Fee),
							Duration:      time.Unix(0, u.Duration).Unix(),
						},
					},
				}

			case *lightning.UpdateChannelOpening:
				log.Infof("Update(%v) received, logging", u)

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_openning,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							RemoteBalance: uint64(u.RemoteBalance),
							LocalBalance:  uint64(u.LocalBalance),
							Fee:           uint64(u.Fee),
							Duration:      0,
						},
					},
				}

			case *lightning.UpdateChannelOpened:
				log.Infof("Update(%v) received, logging", u)

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_opened,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							RemoteBalance: uint64(u.RemoteBalance),
							LocalBalance:  uint64(u.LocalBalance),
							Fee:           uint64(u.Fee),
							Duration:      time.Unix(0, u.Duration).Unix(),
						},
					},
				}

			case *lightning.UpdateChannelUpdating:
				log.Infof("Update(%v) received, logging", u)

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_updating,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							RemoteBalance: uint64(u.RemoteBalance),
							LocalBalance:  uint64(u.LocalBalance),
							Fee:           uint64(u.Fee),
							Duration:      0,
						},
					},
				}

			case *lightning.UpdateChannelUpdated:
				log.Infof("Update(%v) received, logging", u)

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_updated,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							RemoteBalance: uint64(u.RemoteBalance),
							LocalBalance:  uint64(u.LocalBalance),
							Fee:           uint64(u.Fee),
							Duration:      time.Unix(0, u.Duration).Unix(),
						},
					},
				}

			case *lightning.UpdateUserConnected:
				log.Infof("Update(%v) received, logging", u)

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_UserChange{
						UserChange: &UserChange{
							UserId:      string(u.User),
							IsConnected: u.IsConnected,
						},
					},
				}

			case *lightning.UpdatePayment:
				log.Infof("Update(%v) received, logging", u)

				var status PaymentStatus
				switch u.Status {
				case lightning.InsufficientFunds:
					status = PaymentStatus_unsufficient_funds
				case lightning.Successful:
					status = PaymentStatus_success
				case lightning.ExternalFail:
					status = PaymentStatus_external_fail
				case lightning.UserLocalFail:
					status = PaymentStatus_user_local_fail
				default:
					fail(errChan, "unknown status: %v", u.Status)
					return
				}

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_Payment{
						Payment: &Payment{
							Id:       u.ID,
							Status:   status,
							Sender:   string(u.Sender),
							Receiver: string(u.Receiver),
							Amount:   uint64(u.Amount),
							Earned:   int64(u.Earned),
						},
					},
				}

			default:
				log.Errorf("unhandled type of update: %v",
					reflect.TypeOf(u))
				continue
			}

			// After the any log update we have to dump the state of the
			// lightning node.
			triggerStateWrite()

		case <-needWriteState:
			log.Info("Synchronise state of the lightning node and write state" +
				" in the log")

			logEntry, err = getState(client)
			if err != nil {
				fail(errChan, "unable to get state: %v", err)
				return
			}

			needWriteState = nil
		}
	}
}

func fail(errChan chan error, format string, params ...interface{}) {
	err := errors.Errorf(format, params...)
	select {
	case _, ok := <-errChan:
		if !ok {
			return
		}
	default:
	}

	select {
	case errChan <- err:
	default:
	}
}
