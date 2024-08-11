package flowcontrol

import (
	"context"
	"errors"
	"flowcontroller/config"
	flowcfg "flowcontroller/flowcontrol/flowcore/flowconfig"
	"flowcontroller/generics/safemap"
	"flowcontroller/generics/safeslice"
	"flowcontroller/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// # [FlowController] is crucial part of communication between services
//
// # It provides API to read and reply to messages, realise custom logic over read messages
//
// # It also privides storage for dynamic [ServiceMetadata] of registered services
//
//go:generate mockery --filename=mock_flowcontroller.go --name=FlowController --dir=. --structname MockFlowController  --inpackage=true
type FlowController interface {

	// # Returns unique service identifier, do not lose it!
	//
	// Also populates [ControlTunnel]
	Register(service Manageable) (SID, error)

	// # Accepts Additional Options
	//
	// If opts is a zero value works identical to [FlowController.Register]
	RegisterOpt(service Manageable, opts Options) (SID, error)

	// # Dynamicaly get your metadata in case of [MetadataUpdated]
	Metadata(identifier SID) (ServiceMetadata, error)
}

// Service Identifier
type SID uuid.UUID

type controller struct {
	selfSID    SID
	selfConfig config.Config
	selfLogger logger.Logger

	md        *safemap.Map[SID, ServiceMetadata]
	broadcast *safeslice.Slice[chan internalMsg] /*  []chan internalMsg */

	mx *sync.RWMutex
}
type internalMsg struct {
	data MessageFrom
}

// # Registers object in flow controller
//
// # !! Repeatable registration of same [Manageable] is not an error !!
//
// Nothing will break on controller side and completely new [ServiceMetadata] will be created with
// new tunnel assigned to [ControlTunnel] inside [ServiceManager]
//
// However client side may have severe problems if state [Manageable.ServiceManager] operating with
// are not dynamicaly changing per every following registration
// and internal logic have shared, thread unsafe resources, like slices or maps
//
// # In the best case if will cause handling messages multiple time, if logic is idempotent and secure
//
// In the worst case if will cause race condition or deadlock
func (c *controller) Register(service Manageable) (SID, error) {

	c.mx.Lock()
	defer c.mx.Unlock()

	cfg, err := flowcfg.ReadConfig(noexport_CONFIG_KEY)
	if err != nil {
		return SID{}, err
	}

	metadata := newMD(context.TODO(), cfg, c.selfLogger)
	sid := SID(uuid.New())
	c.md.Store(sid, metadata)

	tunnel := ControlTunnel{
		tun: tunnel{
			from: make(chan MessageFrom, noexport_MESSAGE_CHANNEL_BUF),
			to:   make(chan MessageTo, noexport_MESSAGE_CHANNEL_BUF),
		},
	}
	internalCh := make(chan internalMsg, noexport_INTERNAL_MESSAGE_CHANNEL_BUF)
	c.broadcast.Append(internalCh)
	go bridge(internalCh, tunnel, c.selfLogger) // TODO , utilize bridge error
	go service.ServiceManager()(tunnel)         // TODO , utilize servmanager error

	return sid, nil
}

// Transmits
func bridge(internalCh chan internalMsg, tunnel ControlTunnel, logger logger.Logger) error {

	fromClient := tunnel.tun.to // TODO  убрать комменты на русском канал ответов клиента
	toClient := tunnel.tun.from

	deadClientTimer := time.NewTimer(time.Duration(0))

	for {
		stopAndDrain(deadClientTimer)

		internalmsg := <-internalCh // Listening for broadcast messages
		deadClientTimer.Reset(noexport_DEAD_CLIENT_TIMEOUT)

		toClient <- internalmsg.data // Send message to client

		select {
		case reply := <-fromClient:
			handleReply(reply, toClient, fromClient, logger)
		case <-deadClientTimer.C:
			logger.Error(ErrDeadClient)
			return ErrDeadClient
		}
	}

}
func handleReply(reply MessageTo, toClient chan<- MessageFrom, fromClient <-chan MessageTo, logger logger.Logger) {
	switch reply {
	case OK:
	case Pending:
		// pending()
		fn1 := func() MessageTo {
			time.Sleep(STATUS_SPAM_TIMEOUT)
			toClient <- Status
			return <-fromClient

		}
		for reply == Pending {
			reply = fn1()
		}
		// reply is not pending any more
		// only one level of recursion
		handleReply(reply, toClient, fromClient, logger)
	case Error:
		logger.Error(ErrClientsideError)
	case Unimplemented:
		logger.Info(ErrClientsideUnimplemented)
	default:
		logger.Error(ErrUnknownReply)
	}

}
func stopAndDrain(timer *time.Timer) {
	if !timer.Stop() {
		<-timer.C
	}
}

// Flow Controller self manager
func (c *controller) ServiceManager() ServiceManager {
	selfManager := func(tun ControlTunnel) error {

		for {
			_, selfreply := tun.ReadMessage()
			selfreply(OK)

		}
	}
	return selfManager
}

// Returns dynamic metadata, or [ErrMetadataNotFound] if service is not registered
func (c *controller) Metadata(sid SID) (ServiceMetadata, error) {

	return c.md.Load(sid)
	//	if err != nil {
	//		return ServiceMetadata{}, err
	//	}
	//
	// return metadata, nil
}

// Registeres service and binds message flow to service's [ServiceManager]
func (c *controller) RegisterOpt(service Manageable, opts Options) (SID, error) {
	c.selfLogger.Info("TODO RegisterOpt still unimplemented, any opts will be ignored")
	return c.Register(service)
}

// Serve implements FlowController.
func gracefulShutdownHandler(broadcastTo *safeslice.Slice[chan internalMsg], logger logger.Logger) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {

		sig := <-sigs
		logger.Info("Got signal", sig)
		sslice, runlock := broadcastTo.GetRead()
		for _, b := range sslice {
			b <- internalMsg{
				data: GracefulShutdown,
			}
		}
		runlock()

	}
	// kill  --signal 2 __PID__ or CTRL+C
	// kill  --signal 9 __PID__
	// kill  --signal 15 __PID__

}

const (
	noexport_MESSAGE_CHANNEL_BUF = 1

	noexport_INTERNAL_MESSAGE_CHANNEL_BUF = 1 // changing to 0 will make broadcast synchronous
	noexport_METADATA_STORE_PREALLOC_SIZE = 10

	noexport_CONFIG_KEY          = `gD/33YoHZP3BezxvWeGaIw==`
	noexport_DEAD_CLIENT_TIMEOUT = 10 * time.Second

	noexport_BROADCAST_SLICE_INITIAL_LEN = 0
	noexport_BROADCAST_SLICE_INITIAL_CAP = 10
)

func newMD(ctx context.Context, cfg config.Config, l logger.Logger) ServiceMetadata {
	return ServiceMetadata{
		logger: l,
		cfg:    cfg,
		ctx:    ctx,
	}
}

// TODO refactor, too much happening
func New(ctx context.Context) (FlowController, error) {

	cfg, err := flowcfg.ReadConfig(noexport_CONFIG_KEY)
	if err != nil {
		if errors.Is(err, flowcfg.ErrIncorrectConfigKey) {
			return nil, errors.Join(ErrHardcodedConfigKeysMissmatch, err)
		}
		return nil, err // init config internal error
	}

	// TODO, find a way to use logging levels
	logger, _, err := logger.AssembleLogger(cfg)
	if err != nil {
		return nil, err
	}

	ctrlr := &controller{
		selfLogger: logger,
		selfConfig: cfg,
		md:         safemap.Make[SID, ServiceMetadata](noexport_METADATA_STORE_PREALLOC_SIZE),
		mx:         &sync.RWMutex{},
		broadcast: safeslice.Make[chan internalMsg](
			noexport_BROADCAST_SLICE_INITIAL_LEN,
			noexport_BROADCAST_SLICE_INITIAL_CAP,
		),
	}

	sid, err := ctrlr.Register(ctrlr)
	if err != nil {
		return nil, err
	}
	ctrlr.selfSID = sid

	go gracefulShutdownHandler(ctrlr.broadcast, logger)
	return ctrlr, nil

}
