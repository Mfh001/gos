package gen_server

type NamingPacket struct {
    action string
    name string
    server GenServer
    exists bool
}

type NamingServer struct {
	name_map map[string]GenServer
    channel_in chan NamingPacket
    channel_out chan NamingPacket
}

var SharedInstance *NamingServer

func SharedNameServer() *NamingServer {
	return SharedInstance
}

func StartNamingServer() {
	SharedInstance = new(NamingServer)
	SharedInstance.name_map = make(map[string]GenServer)
    SharedInstance.channel_in  = make(chan NamingPacket)
    SharedInstance.channel_out  = make(chan NamingPacket)
	go SharedInstance.loop()
}

func GetGenServer(name string) (GenServer, bool) {
    SharedInstance.channel_in <- NamingPacket{action: "get", name: name}
    packet := <-SharedInstance.channel_out
    return packet.server, packet.exists
}

func SetGenServer(name string, server GenServer) {
    SharedInstance.channel_in <- NamingPacket{action: "set", name: name, server: server}
}

func DelGenServer(name string) {
    SharedInstance.channel_in <- NamingPacket{action: "del", name: name}
}

func (self *NamingServer) loop() error {
	for {
      packet := <-self.channel_in
      switch packet.action {
      case "set":
        self.set(packet.name, packet.server)
      case "get":
        server, exists := self.get(packet.name)
        self.channel_out <- NamingPacket{server: server, exists: exists}
      case "del":
        self.del(packet.name)
      }
	}
}

func (self *NamingServer) set(name string, server GenServer) {
	self.name_map[name] = server
}

func (self *NamingServer) del(name string) {
	delete(self.name_map, name)
}

func (self *NamingServer) get(name string) (GenServer, bool) {
    gen_server, exists := self.name_map[name]
    return gen_server, exists
}
