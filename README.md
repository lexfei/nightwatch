## 使用说明
1. 编译nightwatch二进制文件：
make

2. 启动nightwatch
./admin.sh start

4. 查看监控是否正常运行：
./nightwatch list
ID: monitor的ID
Name: monitor的名字
Times: monitor已经运行的次数
Status: 当前运行状态
FailedAt: 如果失败，展示首次失败时间

## 其它说明
action.alarm 默认4小时告警一次
taskMonitor: duration = 12(小时) & interval:  240(分钟), 一天最多告警3次
