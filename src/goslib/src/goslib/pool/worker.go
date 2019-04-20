package pool

import (
	"goslib/gen_server"
)

func NewWorker(manager *Pool, idx int, handler TaskHandler) (*Worker, error) {
	worker := &Worker{
		idx: 	 idx,
		manager: manager,
		handler: handler,
	}
	server, err := gen_server.New(worker)
	worker.server = server
	return worker, err
}

type Worker struct {
	idx int
	manager *Pool
	handler TaskHandler
	server *gen_server.GenServer
}

func (self *Worker) Process(args interface{}) {
	self.server.Cast(args)
}

func (self *Worker) Init(args []interface{}) (err error) {
	return nil
}

func (self *Worker) HandleCall(req *gen_server.Request) (interface{}, error) {
	return nil, nil
}

func (self *Worker) HandleCast(req *gen_server.Request) {
	defer self.manager.ReturnWorker(self.idx)
	switch params := req.Msg.(type) {
	case *Task:
		result, err := self.handler(params.Params)
		if params.Reply {
			params.Client.Response(result, err)
		}
	}
}

func (self *Worker) Terminate(reason string) (err error) {
	return nil
}