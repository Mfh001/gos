package gen_server

type NamingServer struct {
	name_map map[string]GenServer
}

var SharedInstance NamingServer

func SharedNameServer() NamingServer {
	return SharedInstance
}

func StartNamingServer() {
	SharedInstance = new(NamingServer)
	SharedInstance.name_map = make(map[string]GenServer)
	go SharedInstance.loop()
}

func (self *NamingServer) loop() error {
	for {
		select {
		case Packet := <-self.register_channel:
		case Packet := <-self.register_channel:
		}
	}
}

func (self *NamingServer) Terminate(reason string) (err error) {
	return nil
}

func (self *NamingServer) Register(name string, server GenServer) {
	self.name_map[name] = server
}

func (self *NamingServer) Unregister(name string) {
	delete(self.name_map, name)
}

func (self *NamingServer) Get(name string) GenServer {
	return self.name_map[name]
}
