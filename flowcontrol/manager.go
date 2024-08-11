package flowcontrol

// # Any [Manageable] Can be registered on [FlowController]
//
//go:generate mockery --filename=mock_managebale.go --name=Manageable --dir=. --structname MockManageable  --inpackage=true
type Manageable interface {

	// # Will be called by [FlowController] in moment of registration
	//
	// # [ServiceManager] will receive tunnel that connected with [FlowController]
	ServiceManager() ServiceManager
}

// Service Manager is a function that listens on tunnel and changes state or behavior of some [Manageable] object
//
// Actual tunnel will be provided insede of [FlowController.Register]
type ServiceManager func(ControlTunnel) error
