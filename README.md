# go-task（类Jenkins的任务编排调度工具）
# 呕心沥血，初心不负

# 1，简介
  使用golang语言实现的一个类似于jenkins的任务编排调度工具，支持任务步骤编排、任务调度执行功能；
   
# 2，模块介绍
    1，整个项目包括一个简易的前端服务，使用vue编写的，便于整体功能展示；
    2，后端服务采用gin框架实现的web服务器，结合gorm框架连接数据库，实现了任务创建、任务编排、任务构建、结果响应等功能；
    3，执行节点服务，可以添加执行节点主机，程序自动埋点与执行节点建立连接，添加实时服务探测、任务负载监听、任务分发、任务执行、结果响应与日志回传；

# 3，整体功能介绍
## 1，任务
  通过指定任务名称创建任务，支持界面话编排任务步骤，支持任务组件包括：git代码拉取、shell脚本支持、http调用；
## 2，节点
  master节点（服务节点）可以执行任务外，可以添加Worker节点，通过配置主机ip、登陆用户名、密码添加工作节点，服务会维护与
工作节点之前的连接，每个节点支持配置调度方式：不进行任务调度，自由任务调度和支持任务调度三种方式；
## 3，任务调度
  任务发起后，服务会根据配置的工作节点选择最空闲的可调度的节点进行任务调度发起；
## 4，执行进度与日志
  worker节点与master节点在任务创建后通过webSocket建立长连接实现执行进度的实时同步与任务日志的回写，web端与server端在任务创建
后也是通过建立webSocket长连接进行任务进度的实时同步前端与任务执行日志的展示；