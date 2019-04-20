package pool

import (
	"container/list"
	"goslib/gen_server"
)

type Pool struct {
	server *gen_server.GenServer
}

type TaskHandler func(args interface{}) (interface{}, error)

type Task struct {
	Params interface{}
	Client *gen_server.Request
	Reply  bool
}

type Manager struct {
	tasks *list.List
	idleWorkers *list.List
	workers []*Worker
}

func New(size int, handler TaskHandler) (pool *Pool, err error) {
	pool = &Pool{}
	manager := &Manager{
		workers: make([]*Worker, size),
	}
	// init manager
	pool.server, err = gen_server.New(&Manager{}, size, handler)
	if err != nil {
		return
	}
	// init workers
	for i := 0; i < size; i++ {
		worker, err := NewWorker(pool, i, handler)
		if err != nil {
			return nil, err
		}
		manager.workers[i] = worker
	}

	return
}

func (self *Pool) Process(args interface{}) (interface{}, error) {
	return self.server.ManualCall(&TaskParams{args})
}

func (self *Pool) ProcessAsync(args interface{}) {
	self.server.Cast(&TaskParams{args})
}

type ReturnWorkerParams struct {idx int}
func (self *Pool) ReturnWorker(idx int) {
	self.server.Cast(&ReturnWorkerParams{idx})
}

func (self *Manager) Init(args []interface{}) (err error) {
	return nil
}

func (self *Manager) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *TaskParams:
		task := &Task{
			Params: params.Msg,
			Client: req,
			Reply:  true,
		}
		worker := self.idleWorkers.Front()
		if worker != nil {
			self.idleWorkers.Remove(worker)
			worker.Value.(*Worker).Process(task)
		} else {
			self.tasks.PushBack(task)
		}
	}
	return nil, nil
}

type TaskParams struct {Msg interface{}}

func (self *Manager) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *TaskParams:
		task := &Task{
			Params: params.Msg,
			Client: req,
			Reply:  false,
		}
		worker := self.idleWorkers.Front()
		if worker != nil {
			self.idleWorkers.Remove(worker)
			worker.Value.(*Worker).Process(task)
		} else {
			self.tasks.PushBack(task)
		}
		break
	case *ReturnWorkerParams:
		task := self.tasks.Front()
		if task != nil {
			self.tasks.Remove(task)
			task := task.Value.(*Task)
			self.workers[params.idx].Process(task)
		} else {
			self.idleWorkers.PushBack(self.workers[params.idx])
		}
		break
	}
}

func (self *Manager) Terminate(reason string) (err error) {
	return nil
}
