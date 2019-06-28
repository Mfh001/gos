#编程环境
brew isntall python3
brew install node

#安装依赖
pip3 install xlrd

#导出JSON配置文件
./tools/gen_excels

#更新配置文件
redis-cli -x SET __gs_configs__  < configData.json.gz

#发布更新事件
redis-cli PUBLISH __channel_reload_config__ update