# qcloudddns
背景：由于路由器上有多个PPPOE线路，不同的线路有不同的ip地址，一般的DDNS程序只能自动更新一个IP地址，所以使用腾讯云的DNS解析服务开发了本程序。
特点：可对多个接口进行自动ddns更新。
工作原理：程序每分钟检查一次本地网络接口ip地址，如有变更就用进行更新，用于多线场景
注意：使用是需要先在自己账号创建好对应的记录，以便程序自动进行更新。