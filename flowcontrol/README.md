Realise Managable interface

Handle messages inside your ServiceManager func via ReadMessage

List of messages and guaranties that FlowController provies defined in protocol.go

Register your service

Now you can access your metadata and populate config and logger

!! WARNING !! Be careful with logging inside your ServiceManager,
wait untill you populate your Logger, otherwise it may cause nil ptr dereference

if we assume that you populate your logger right after registring, this code will do the trick
```
func (you *YourService)Register()ServiceManager{
    return func()error{
        for you.logger == nil{
            time.Sleep(time.Millisecond * 10)
        }
        for{
            Messages Handling Cycle ...
        }
    }
}
func New(flow FlowController)YourService{
    yService := YourService{}
    md, err := flow.Register(yService)
    if err != nil {
        do something
    }
    yService.logger = md.Logger()
}
    
```
