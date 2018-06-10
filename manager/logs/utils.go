package logs

import (
	"time"
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
	"os"
	"github.com/davecgh/go-spew/spew"
	"reflect"
)

// getState converts router topology into state log.
func getState(r router.Router) (*Log, error) {
	// Get the router local network in order to write it on log file,
	// so that external optimisation program could sync it state.
	routerChannels, err := r.Network()
	if err != nil {
		return nil, err
	}

	// TODO(andrew.shvv) As far we have gap between those two operations
	// the free balance might actually change, need to cope with that.

	// Get the number of money under control of router in order to write
	// in the log.
	freeBalance, err := r.FreeBalance()
	if err != nil {
		return nil, err
	}

	// Get the number of money which is in the process of being proceeded by
	// blockchain.
	pendingBalance, err := r.PendingBalance()
	if err != nil {
		return nil, err
	}

	duration, err := r.AverageChangeUpdateDuration()
	if err != nil {
		return nil, err
	}
	milliseconds := duration.Nanoseconds() / int64(time.Millisecond)

	channels := make([]*Channel, len(routerChannels))
	for i, c := range routerChannels {
		channels[i] = &Channel{
			ChannelId:     string(c.ChannelID),
			UserId:        string(c.UserID),
			UserBalance:   uint64(c.UserBalance),
			RouterBalance: uint64(c.RouterBalance),
			IsPending:     c.IsPending,
		}
	}

	return &Log{
		Time: time.Now().UnixNano(),
		Data: &Log_State{
			State: &RouterState{
				FreeBalance:                 uint64(freeBalance),
				PendingBalance:              uint64(pendingBalance),
				AverageChangeUpdateDuration: uint64(milliseconds),
				Channels:                    channels,
			},
		},
	}, nil
}

// UpdateLogFileGoroutine subscribe on routers topology update and update
// the log with the current router state and channel updates.
//
// NOTE: Should be run as goroutine.
func UpdateLogFileGoroutine(r router.Router, path string, errChan chan error) {
	var logEntry *Log

	// Ensure that gRPC structures are printed properly
	pretty := spew.NewDefaultConfig()
	pretty.DisableMethods = true
	pretty.DisablePointerAddresses = true

	logEntry, err := getState(r)
	if err != nil {
		fail(errChan, "unable to get state: %v", err)
		return
	}

	var needWriteState <-chan time.Time
	triggerStateWrite := func() {
		needWriteState = nil
		needWriteState = time.After(3 * time.Second)
	}

	receiver := r.RegisterOnUpdates()
	defer receiver.Stop()

	for {
		if logEntry != nil {
			// NOTE: If move open/close of the file out of this cycle than this
			// would lead to optimisation third-party program unable to get and
			// log update via watchdog package.
			log.Tracef("Open update log file(%v) to write an update: %v",
				path, pretty.Sdump(logEntry))
			updateLogFile, err := os.OpenFile(path, os.O_APPEND | os.O_RDWR|
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
				log.Info("Router update channel close, " +
					"exiting log update goroutine")
				return
			}

			switch u := update.(type) {
			case *router.UpdateChannelClosing:
				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_closing,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							UserBalance:   0,
							RouterBalance: 0,
							Fee:           uint64(u.Fee),
						},
					},
				}

			case *router.UpdateChannelClosed:
				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_closed,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							UserBalance:   0,
							RouterBalance: 0,
							Fee:           uint64(u.Fee),
						},
					},
				}

			case *router.UpdateChannelOpening:
				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_openning,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							UserBalance:   uint64(u.UserBalance),
							RouterBalance: uint64(u.RouterBalance),
							Fee:           uint64(u.Fee),
						},
					},
				}

			case *router.UpdateChannelOpened:
				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_opened,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							UserBalance:   uint64(u.UserBalance),
							RouterBalance: uint64(u.RouterBalance),
							Fee:           uint64(u.Fee),
						},
					},
				}

			case *router.UpdateChannelUpdating:
				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_updating,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							UserBalance:   uint64(u.UserBalance),
							RouterBalance: uint64(u.RouterBalance),
							Fee:           uint64(u.Fee),
						},
					},
				}

			case *router.UpdateChannelUpdated:
				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_ChannelChange{
						ChannelChange: &ChannelChange{
							Type:          ChannelChangeType_updated,
							ChannelId:     string(u.ChannelID),
							UserId:        string(u.UserID),
							UserBalance:   uint64(u.UserBalance),
							RouterBalance: uint64(u.RouterBalance),
							Fee:           uint64(u.Fee),
						},
					},
				}

			case *router.UpdateLinkAverageUpdateDuration:
				// With this update we just trigger state update
			case *router.UpdatePayment:
				var status PaymentStatus
				switch u.Status {
				case router.InsufficientFunds:
					status = PaymentStatus_unsufficient_funds
				case router.Successful:
					status = PaymentStatus_success
				case router.ExternalFail:
					status = PaymentStatus_external_fail
				default:
					fail(errChan, "unknown status: %v", u.Status)
					return
				}

				logEntry = &Log{
					Time: time.Now().UnixNano(),
					Data: &Log_Payment{
						Payment: &Payment{
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
			// router
			triggerStateWrite()

		case <-needWriteState:
			log.Info("Synchronise state of the router and write state in the log")

			logEntry, err = getState(r)
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
