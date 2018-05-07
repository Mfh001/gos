package root_mgr

/*
Manage the whole distributed servers

Hardware:
	Deploy: hardwares by hardwareInitConf or commands from manager(people)
	Automatic: deploy hardwares when distributed servers face heavy load

Application:
	Deploy: applications to target Hardware and start it
	Monit: deployed applications health

Scene:
	1.CRUD scenes by manager's commands
*/

/*
使用分析：
手动部署：
	1.world服务双机+负载均衡
	2.auth服务双机+负载均衡
自动部署：
	1.agent
	2.game

服务所需软件：
	Game: mysql-client

外部服务：
	MySQL
	Redis
*/
func Start() {
}
