package store

type packet struct {
	action byte
	key    string
	value  interface{}
}

type Ets struct {
	channel_in  chan *packet
	channel_out chan interface{}
	store       map[string]interface{}
}

const (
	GET = 1
	SET = 2
	DEL = 3
)

var sharedInstance *Ets

func InitSharedInstance() {
	sharedInstance = New()
}

func New() *Ets {
	e := &Ets{
		channel_in:  make(chan *packet),
		channel_out: make(chan interface{}),
		store:       make(map[string]interface{}),
	}
	go e.loop()
	return e
}

func Get(key string) {
	sharedInstance.Get(key)
}
func Set(key string, value interface{}) {
	sharedInstance.Set(key, value)
}
func Del(key string) {
	sharedInstance.Del(key)
}

func (e *Ets) Get(key string) interface{} {
	e.channel_in <- &packet{GET, key, nil}
	value := <-e.channel_out
	return value
}

func (e *Ets) Set(key string, value interface{}) {
	e.channel_in <- &packet{SET, key, value}
}

func (e *Ets) Del(key string) {
	e.channel_in <- &packet{DEL, key, nil}
}

func (e *Ets) loop() {
	for {
		data := <-e.channel_in
		switch data.action {
		case GET:
			e.channel_out <- e.store[data.key]
		case SET:
			e.store[data.key] = data.value
		case DEL:
			delete(e.store, data.key)
		}
	}
}
